package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"iter"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	pggen "git.apsolutions.ru/aps/Internal/streaming-platform/source-code/libs/pg-composite-parser-gen.git"
)

const driverValueToStringName = "__intrinsic_computeStringFromDriverValuer"
const driverValueToString = `
func __intrinsic_computeStringFromDriverValuer(value driver.Value) string {
	switch v := value.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "t"
		} else {
			return "f"
		}
	case []byte:
		return strconv.Quote("\\x" + hex.EncodeToString(v))
	case string:
		return strconv.Quote(v)
	case time.Time:
		return strconv.Quote(v.String())
	default:
		panic("all is bad: bad value happened inside '__intrinsic_computeStringFromDriverValuer' intrinsic: " + fmt.Sprintf("%#v", value))
	}
}
`
const readStringName = "__intrinsic_readString"
const readString = `
func __intrinsic_readString(b []byte, delim byte) (res string, read int, err error) {
	switch b[0] {
	case '"':
		var buf bytes.Buffer
		for i := 1; i < len(b); i++ {
			var n = b[i]
			switch n {
			case '"':
				i++
				if i == len(b) {
					return buf.String(), i, nil
				} else if n = b[i]; n == '"' {
					buf.WriteByte('"')
				} else {
					return buf.String(), i, nil
				}
			case '\\':
				i++
				if i == len(b) {
					return "", 0, fmt.Errorf("bad escape sequence: %q", string(b))
				}
				n = b[i]
				switch n {
				case '"':
					buf.WriteByte('"')
				case '\\':
					buf.WriteByte('\\')
				default:
					return "", 0, fmt.Errorf("bad escape sequence: %q", string(b))
				}
			default:
				buf.WriteByte(n)
			}
		}
		return "", 0, fmt.Errorf("bad escaped string literal: %s", string(b))
	default:
		for i := range b {
			if b[i] == delim || b[i] == ',' {
				return string(b[:i]), i, nil
			}
		}
		return "", 0, fmt.Errorf(` + "`readString should be called only for '\"', '{' or '(' - delimited strings or records or array literals`" + `)
	}
}
`
const shouldAddFilenameFromError = true

type Module struct {
	Path      string
	Version   string
	Replace   *Module
	Time      *time.Time
	Main      bool
	Indirect  bool
	Dir       string
	GoMod     string
	GoVersion string
	Error     *struct {
		Err string
	}
}

// Fields must match go list;
// see $GOROOT/src/cmd/go/internal/load/pkg.go.
//
// I cannot be said to have stolen this shit
// until a legitimate court rules otherwise.
type jsonPackage struct {
	ImportPath        string
	Dir               string
	Name              string
	Target            string
	Export            string
	GoFiles           []string
	CompiledGoFiles   []string
	IgnoredGoFiles    []string
	IgnoredOtherFiles []string
	EmbedPatterns     []string
	EmbedFiles        []string
	CFiles            []string
	CgoFiles          []string
	CXXFiles          []string
	MFiles            []string
	HFiles            []string
	FFiles            []string
	SFiles            []string
	SwigFiles         []string
	SwigCXXFiles      []string
	SysoFiles         []string
	Imports           []string
	ImportMap         map[string]string
	Deps              []string
	Module            *Module
	TestGoFiles       []string
	TestImports       []string
	XTestGoFiles      []string
	XTestImports      []string
	ForTest           string // q in a "p [q.test]" package, else ""
	DepOnly           bool

	Error *struct {
		ImportStack []string
		Pos         string
		Err         string
	}
	DepsErrors []*struct {
		ImportStack []string
		Pos         string
		Err         string
	}
}

func absolutize(directory string, fs ...[]string) (res []string) {
	for _, files := range fs {
		for _, file := range files {
			if !filepath.IsAbs(file) {
				file = filepath.Join(directory, file)
			}
			res = append(res, file)
		}
	}
	return res
}

var goListBuffer bytes.Buffer

// NOT THREAD SAFE
func runGoToolAndIterateResults(ppath ...string) iter.Seq2[*jsonPackage, error] {
	var cmd = exec.Command("go", append([]string{"list", "-e", "-compiled", "-export=true", "-json", "--"}, ppath...)...)
	goListBuffer.Grow(1024 * 1024)
	cmd.Stdout = &goListBuffer
	if err := cmd.Run(); err != nil {
		return func(yield func(*jsonPackage, error) bool) { yield(nil, err) }
	}
	return func(yield func(*jsonPackage, error) bool) {
		// var dec = json.NewDecoder(io.TeeReader(&buf, os.Stdout))
		var dec = json.NewDecoder(&goListBuffer)
		for dec.More() {
			var out = new(jsonPackage)
			if !yield(out, dec.Decode(out)) {
				return
			}
		}
	}
}

type needs struct {
	scanner bool
	valuer  bool
}
type typeNeedsMap map[*types.Package]map[string]needs
type generatorState struct {
	pggen.AstAcc

	nowPkg *types.Package

	seen    map[string]*jsonPackage
	pkgs    map[string]*Package
	checked map[*Package]struct{}
	fset    *token.FileSet

	generatedPackages map[*types.Package]struct{}
	// named types that we encounter and that we need to generate
	// scanner and valuer implementations for.
	typeNeeds typeNeedsMap
	seenTypes typeNeedsMap

	curPlusNOne *ast.AssignStmt
	curPlusN    *ast.AssignStmt
}

func (g *generatorState) unsafeBytesRender(e ast.Expr) *ast.CallExpr {
	return g.ImportAndCall("unsafe", "Slice")(
		g.ImportAndCall("unsafe", "StringData")(e),
		g.Len(e),
	)
}
func (g *generatorState) unsafeStringRender(e ast.Expr) *ast.CallExpr {
	return g.ImportAndCall("unsafe", "String")(
		g.ImportAndCall("unsafe", "SliceData")(e),
		g.Len(e),
	)
}

func (g *generatorState) GenerateScanForStruct(name string, spec *types.Struct, prologue []ast.Stmt) ([]ast.Stmt, error) {
	var checkSrc = g.If(g.Or(
		g.Neq(g.IndexInt(g.I("sourceBytes"), 0), g.AsLitChar("(")),
		g.Neq(g.Index(g.I("sourceBytes"), g.SubConst(g.Len(g.I("source")), 1)), g.AsLitChar(")")),
	))(g.Return(g.ImportAndCall("fmt", "Errorf")(g.AsLit("bad source string: %q"), g.I("source"))))
	var declCur = g.VarDeclInit("cur")(g.AsLitInt(1))
	prologue = append(prologue, checkSrc, declCur)

	for field := range spec.Fields() {
		var fname = field.Name()
		var ftype = field.Type()
		prologue = append(prologue, g.renderScanner(
			ftype,
			g.Selector(g.I("v"), fname),
			g.SliceExpr1(g.I("sourceBytes"), g.I("cur"))))
	}

	var ignoreCur = g.Assign(g.Blank)(g.I("cur"))
	return append(prologue, ignoreCur, g.ReturnNil), nil
}

var (
	scannerIface = types.NewInterfaceType(
		[]*types.Func{types.NewFunc(
			token.NoPos, nil, "Scan", types.NewSignatureType(nil, nil, nil,
				types.NewTuple(
					types.NewVar(token.NoPos, nil, "src", types.Universe.Lookup("any").Type())),
				types.NewTuple(
					types.NewVar(token.NoPos, nil, "err", types.Universe.Lookup("error").Type())), false),
		)}, nil).Complete()
	valuerIface = types.NewInterfaceType(
		[]*types.Func{types.NewFunc(
			token.NoPos, nil, "Value", types.NewSignatureType(nil, nil, nil, nil,
				types.NewTuple(
					types.NewVar(token.NoPos, nil, "", types.NewTypeName(
						token.NoPos, types.NewPackage("database/sql/driver", "driver"), "Value", types.Universe.Lookup("any").Type(),
					).Type()),
					types.NewVar(token.NoPos, nil, "", types.Universe.Lookup("error").Type())), false),
		)}, nil).Complete()
)

func impl(t types.Type, i *types.Interface) bool {
	return types.Implements(t, i) || types.Implements(types.NewPointer(t), i)
}

func (g *generatorState) renderValuerBasic(typ *types.Basic, what, where ast.Expr) ast.Stmt {
	switch typ.Kind() {
	case types.UntypedString, types.String:
		return g.Assign(where)(g.ImportAndCall("strconv", "Quote")(what))
	case types.Bool, types.UntypedBool:
		return g.IfElse(what)(
			g.Assign(where)(g.AsLit("t")),
		)(
			g.Assign(where)(g.AsLit("f")),
		)
	case types.Complex64, types.Complex128, types.UntypedComplex:
		panic("postgres can't handle a complex number")
	case types.Float32:
		return g.Assign(where)(g.ImportAndCall("strconv", "FormatFloat")(
			g.Cast(what, g.Int64),
			g.AsLitChar("f"),
			g.AsLitInt(-1),
			g.AsLitInt(32)))
	case types.Float64, types.UntypedFloat:
		return g.Assign(where)(g.ImportAndCall("strconv", "FormatFloat")(
			g.Cast(what, g.Int64),
			g.AsLitChar("f"),
			g.AsLitInt(-1),
			g.AsLitInt(64)))
	case types.Int, types.UntypedInt, types.Int16,
		types.Int32, types.UntypedRune, types.Int64,
		types.Int8:
		return g.Assign(where)(g.ImportAndCall("strconv", "FormatInt")(g.Cast(what, g.Int64), g.AsLitInt(10)))
	case types.Uint, types.Uint8,
		types.Uint16, types.Uint32, types.Uint64,
		types.Uintptr:
		return g.Assign(where)(g.ImportAndCall("strconv", "FormatUint")(g.Cast(what, g.Uint64), g.AsLitInt(10)))
	case types.UnsafePointer:
		return g.Assign(where)(g.ImportAndCall("strconv", "FormatUint")(g.Cast(g.Cast(what, g.Uintptr), g.Uint64)), g.AsLitInt(10))
	default:
		panic("unreachable")
	}
}
func (g *generatorState) renderValuableValuerForField(what, where ast.Expr) ast.Stmt {
	return g.Block(
		g.VarDeclInit("value", "err")(g.MethodCall(what, "Value")()),
		g.If(g.Neq(g.Err, g.Nil))(g.Return(g.Nil, g.Err)),
		g.Assign(where)(g.FuncCall(driverValueToStringName)(g.I("value"))),
	)
}
func (g *generatorState) renderValuer(typ types.Type, what, where ast.Expr) ast.Stmt {
	switch v := typ.(type) {
	case *types.Pointer:
		return g.IfElse(g.Eq(what, g.Nil))(
			g.Assign(where)(g.AsLit("")),
		)(
			g.renderValuer(v.Elem(), g.Star(what), where),
		)
	case *types.Slice:
		return g.renderValuerRangeable(v.Elem(), what, where)
	case *types.Array:
		return g.renderValuerRangeable(v.Elem(), what, where)
	case *types.Named:
		if _, ok := g.generatedPackages[v.Obj().Pkg()]; ok {
			// For packages that we do actually care about we may generate
			// a scanner implementation ourselves; we queue for generation (if needed)
			// and then just generate the call
			if impl(v, valuerIface) {
				for i := range v.Methods() {
					if i.Name() == "Value" {
						var file = g.fset.File(i.Pos())
						if !strings.HasSuffix(file.Name(), "zz_scannervaluer.generated.go") {
							return g.renderValuableValuerForField(what, where)
						}
						break
					}
				}
			}
			g.queueObjForValuer(v.Obj())
			return g.renderValuableValuerForField(what, where)
		} else if impl(v, valuerIface) {
			// For external types implementing the driver.Valuer interface
			return g.renderValuableValuerForField(what, where)
		} else {
			panic("TODO: return error probably or something")
		}
	case *types.Basic:
		return g.renderValuerBasic(v, what, where)
	case *types.Alias:
		return g.renderValuer(v.Underlying(), what, where)
	case *types.Map:
		panic("TODO: implement maps as an array of two-field structs: key and value")
	case *types.Struct:
		panic("TODO: implement anonymous structs (using the same heuristics)")
	case *types.Interface:
		// TODO: support generic types where the thing implements the
		//       needed interface
		// case
		// 	*types.Interface,
		// if impl(v, scannerIface) {
		// 	return g.renderScannableScannerForField(place, from)
		// }
		panic("TODO: return an error on interface/type param not implementing the needed interface")
	case *types.TypeParam:
		panic("TODO: generation of type param-based scanner")
	case *types.Chan:
		panic("The channel type is never going to be supported, probably. How am i supposed to read a CHANNEL from a postgresql database? More on postgresql (version 17) types can be read here: https://www.postgresql.org/docs/17/datatype.html (note the absense of golang channel type in the list)")
	// These should probably never happen:
	case
		*types.Signature,
		*types.Tuple,
		*types.Union:
		panic(fmt.Sprintf(
			"The %T type is not implemented, and, honestly, it should have never been possible for you to trick me into dealing with this kind of a type. Whatever you are doing, stop.", v))
	default:
		panic(fmt.Sprintf("unexpected type kind(%T): %#v", v, v))
	}
	return g.Block()
}

func (g *generatorState) renderValuerRangeable(elem types.Type, what ast.Expr, where ast.Expr) ast.Stmt {
	var stmts = make([]ast.Stmt, 0)
	stmts = append(stmts, g.VarDeclType("value2", g.String))
	stmts = append(stmts, g.VarDeclType("value2Sb", g.ImportAndUse("strings", "Builder")))
	var value2 = g.I("value2")
	var wvalue = g.Stmt(g.MethodCall(g.I("value2Sb"), "WriteString")(value2))
	var wcommaIf = g.If(g.Neq(g.I("i"), g.SubConst(g.Len(what), 1)))(
		g.Stmt(g.MethodCall(g.I("value2Sb"), "WriteByte")(g.AsLitChar(","))),
	)
	stmts = append(stmts, g.Stmt(g.MethodCall(g.I("value2Sb"), "WriteByte")(g.AsLitChar("{"))))
	var val = g.I("val")
	stmts = append(stmts, g.Range2Def(g.I("i"), val, what)(
		g.renderValuer(elem, val, value2),
		wvalue,
		wcommaIf,
	))
	stmts = append(stmts, g.Stmt(g.MethodCall(g.I("value2Sb"), "WriteByte")(g.AsLitChar("}"))))
	stmts = append(stmts,
		g.Assign(where)(
			g.MethodCall(g.I("value2Sb"), "String")()))
	return g.Block(
		stmts...)
}
func (g *generatorState) renderScanner(typ types.Type, place, from ast.Expr) ast.Stmt {
	switch v := typ.(type) {
	case *types.Pointer:
		// should somehow deblock this as there is no need for scoping
		// TODO: denest this parser in the caller, since having {{{{ scopes like that }}}}
		//       is bad.
		//
		// TODO: make pointer-typed fields probably optional???
		//       That is postgres nulls should just translate into a nil here.
		//       I dont know.
		return g.Block(
			g.Assign(place)(g.New(g.ValueTypeExpr(v.Elem()))),
			g.renderScanner(v.Elem(), g.Star(place), from),
		)
	case *types.Slice:
		return g.renderSliceScannerFor(place, from, v.Elem())
	// TODO: use the same logic like we employ in the slice parser
	//       to parse the array thing
	case *types.Array:
		var arrLen = g.AsLitInt(int(v.Len()))
		var sliceAsArray = g.I("sliceAsArray")
		var sliceLen = g.Len(sliceAsArray)
		return g.Block(
			g.VarDecl2(g.ValueSpec("sliceAsArray")(g.SliceType(g.ValueTypeExpr(v.Elem())))),
			g.renderSliceScannerFor(sliceAsArray, from, v.Elem()),
			g.If(g.Neq(sliceLen, arrLen))(
				g.Return(
					g.Errorf(fmt.Sprintf(
						"bad parsed array element count: got %%d, expected %d", v.Len(),
					), sliceLen)),
			),
			g.Stmt(g.FuncCall("copy")(
				g.Slice2(place, 0, int(v.Len())), sliceAsArray,
			)),
		)
		// panic("TODO: array parsing")
	case *types.Named:
		if _, ok := g.generatedPackages[v.Obj().Pkg()]; ok {
			// For packages that we do actually care about we may generate
			// a scanner implementation ourselves; we queue for generation (if needed)
			// and then just generate the call
			if impl(v, scannerIface) {
				for i := range v.Methods() {
					if i.Name() == "Scan" {
						var file = g.fset.File(i.Pos())
						if !strings.HasSuffix(file.Name(), "zz_scannervaluer.generated.go") {
							return g.renderScannableScannerForField(place, from)
						}
						break
					}
				}
			}
			g.queueObjForScanner(v.Obj())
			return g.renderScannableScannerForField(place, from)
		} else if impl(v, scannerIface) {
			// For external types implementing the driver.Valuer interface
			return g.renderScannableScannerForField(place, from)
		} else {
			panic("TODO: return error probably or something")
		}
	case *types.Basic:
		return g.renderBasicScan(v, place, from)
	case *types.Alias:
		return g.renderScanner(v.Underlying(), place, from)
	case *types.Map:
		panic("TODO: implement maps as an array of two-field structs: key and value")
	case *types.Struct:
		panic("TODO: implement anonymous structs (using the same heuristics)")
	case *types.Interface:
		// TODO: support generic types where the thing implements the
		//       needed interface
		// case
		// 	*types.Interface,
		if impl(v, scannerIface) {
			return g.renderScannableScannerForField(place, from)
		}
		panic("TODO: return an error on interface/type param not implementing the needed interface")
	case *types.TypeParam:
		panic("TODO: generation of type param-based scanner")
	case *types.Chan:
		panic("The channel type is never going to be supported, probably. How am i supposed to read a CHANNEL from a postgresql database? More on postgresql (version 17) types can be read here: https://www.postgresql.org/docs/17/datatype.html (note the absense of golang channel type in the list)")
	// These should probably never happen:
	case
		*types.Signature,
		*types.Tuple,
		*types.Union:
		panic(fmt.Sprintf(
			"The %T type is not implemented, and, honestly, it should have never been possible for you to trick me into dealing with this kind of a type. Whatever you are doing, stop.", v))
	default:
		panic(fmt.Sprintf("unexpected type kind(%T): %#v", v, v))
	}
}

func (g *generatorState) renderBasicScan(v *types.Basic, place ast.Expr, from ast.Expr) ast.Stmt {
	switch name := v.Name(); name {
	case "float32":
		return g.renderFLoatScannerForField(place, from, g.Float32, 32)
	case "float64":
		return g.renderFLoatScannerForField(place, from, g.Float64, 64)
	case "bool":
		return g.renderBooleanScannerForField(place, from)
	case "string":
		return g.renderStringScannerForField(place, from)
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"uintptr":
		return g.renderIntScannerForField(place, from, g.IntFromString(name))
	default:
		panic("unreachable")
	}
}

func (g *generatorState) renderScannableScannerForField(place, from ast.Expr) *ast.IfStmt {
	return g.IfElse2(
		g.AssignStmt("next", "n", "e")(
			g.FuncCall("__intrinsic_readString")(
				from,
				g.AsLitChar(")"))),
		g.Eq(g.I("n"), g.AsLitInt(0)),
	)()(
		g.IfElse(
			g.Neq(g.I("e"), g.Nil),
		)(
			g.Return(g.I("e")),
		)(
			g.Assign(g.I("cur"))(
				g.Add(g.I("cur"), g.I("n"))),
			g.If2(
				g.AssignStmt("e")(
					g.MethodCall(place, "Scan")(
						g.I("next"))),
				g.Neq(g.I("e"), g.Nil),
			)(g.Return(g.I("e"))),
		),
	)
}
func (g *generatorState) renderIntScannerForField(place, from ast.Expr, castTo ast.Expr) *ast.IfStmt {
	return g.IfElse2(
		g.AssignStmt("next", "n", "e")(
			g.FuncCall("__intrinsic_readString")(
				from,
				g.AsLitChar(")"))),
		g.Neq(g.I("e"), g.Nil),
	)(
		g.Return(g.I("e")),
	)(
		g.IfElse2(
			g.AssignStmt("answer", "e")(
				g.ImportAndCall("strconv", "ParseInt")(
					g.I("next"),
					g.AsLitInt(10),
					g.AsLitInt(64))),
			g.Neq(g.I("e"), g.Nil),
		)(
			g.Return(g.I("e")),
		)(
			g.Block(
				g.Assign(g.I("cur"))(
					g.Add(g.I("cur"), g.I("n"))),
				g.Assign(place)(
					g.Cast(g.I("answer"), castTo)))))
}

func (g *generatorState) renderFLoatScannerForField(place, from ast.Expr, flavour *ast.Ident, bits int) *ast.IfStmt {
	return g.IfElse2(
		g.AssignStmt("next", "n", "e")(
			g.FuncCall("__intrinsic_readString")(
				from,
				g.AsLitChar(")"))),
		g.Neq(g.I("e"), g.Nil),
	)(
		g.Return(g.I("e")),
	)(
		g.IfElse2(
			g.AssignStmt("vfloat", "e")(
				g.ImportAndCall("strconv", "ParseFloat")(
					g.I("next"),
					g.AsLitInt(bits))),
			g.Neq(g.I("e"), g.Nil),
		)(
			g.Return(g.I("e")),
		)(
			g.Block(
				g.Assign(g.I("cur"))(
					g.Add(g.I("cur"), g.I("n"))),
				g.Assign(place)(
					g.Cast(g.I("vfloat"), flavour)))))
}
func (g *generatorState) renderBooleanScannerForField(place, from ast.Expr) *ast.IfStmt {
	return g.If3(
		g.AssignStmt("next", "n", "e")(
			g.FuncCall("__intrinsic_readString")(
				from,
				g.AsLitChar(")"))),
		g.Neq(g.I("e"), g.Nil),
		g.Return(g.I("e")),
	)(
		g.Assign(g.I("cur"))(g.Add(g.I("cur"), g.I("n"))),
		g.Switch(g.I("next"))(
			g.Case(g.AsLit("t"))(
				g.Assign(place)(g.True)),
			g.Case(g.AsLit("f"))(
				g.Assign(place)(g.False)),
			g.Case()(
				g.Stmt(g.FuncCall("panic")(
					g.Add(
						g.AsLit("bad bool string from postgres: "),
						g.I("next")))))))
}

// TODO: currently compiled packages should not import themselves.
func (g *generatorState) renderSliceScannerFor(place, from ast.Expr, elem types.Type) *ast.BlockStmt {
	return g.Block(
		g.Assign(
			place)(
			g.Make2(
				g.ValueTypeExpr(elem),
				g.AsLitInt(0))),
		g.If3(
			g.AssignStmt("next", "n", "e")(
				g.FuncCall("__intrinsic_readString")(
					from,
					g.AsLitChar(")"))),
			g.Neq(g.I("e"), g.Nil),
			g.Return(g.I("e")),
		)(
			g.VarDeclInit("sourceLen")(g.Len(g.I("next"))),
			g.Assign(
				g.I("cur"))(
				g.Add(
					g.I("cur"),
					g.I("n"),
					g.AsLitInt(1))),
			g.For2(
				g.AssignStmt("cur")(g.AsLitInt(1)),
				g.Lt(g.I("cur"), g.SubConst(g.I("sourceLen"), 1)),
			)(
				g.Append(place, g.ZeroValue(elem)),
				g.renderScanner(
					elem,
					g.Index(place, g.SubConst(g.Len(place), 1)),
					g.SliceExpr1(g.unsafeBytesRender(g.I("next")), g.I("cur"))),
				g.Assign(
					g.I("cur"))(
					g.Add(
						g.I("cur"),
						g.AsLitInt(1))))))
}

func (g *generatorState) renderStringScannerForField(place, from ast.Expr) *ast.IfStmt {
	return g.If3(
		g.AssignStmt("next", "n", "e")(
			g.FuncCall("__intrinsic_readString")(
				from,
				g.AsLitChar(")"))),
		g.Neq(g.I("e"), g.Nil),
		g.Return(g.I("e")),
	)(
		g.Block(
			g.Assign(g.I("cur"))(
				g.Add(g.I("cur"), g.I("n"))),
			g.Assign(place)(
				g.I("next"))))
}

type importerFunc func(path string) (*types.Package, error)

func (f importerFunc) Import(path string) (*types.Package, error) { return f(path) }

type Error struct {
	Pos  string
	Msg  string
	Kind ErrorKind
}

// ErrorKind describes the source of the error, allowing the user to
// differentiate between errors generated by the driver, the parser, or the
// type-checker.
type ErrorKind int

const (
	UnknownError ErrorKind = iota
	ListError
	ParseError
	TypeError
)

func (err Error) Error() string {
	pos := err.Pos
	if pos == "" {
		pos = "-" // like token.Position{}.String()
	}
	return pos + ": " + err.Msg
}

type Package struct {
	Name            string
	ID              string
	PkgPath         string
	Dir             string
	Target          string
	ExportFile      string
	GoFiles         []string
	CompiledGoFiles []string
	OtherFiles      []string
	EmbedFiles      []string
	EmbedPatterns   []string
	IgnoredFiles    []string
	ForTest         string
	Module          *Module
	Imports         map[string]*Package
	Errors          []Error

	Syntax    []*ast.File
	TypesInfo types.Info
	Types     *types.Package
}

type TypecheckingFlags struct {
	TypecheckCgo bool
}

const (
	marker_generateValuer  = "#[generate(Valuer)]"
	marker_generateScanner = "#[generate(Scanner)]"
	marker_excludeValuer   = "#[exclude(Valuer)]"
	marker_excludeScanner  = "#[exclude(Scanner)]"
)

// This is a fucking shame, my brother
//
// This iterator iterates over two maps, map a and map b.
// when iterating over current g.typeNeeds (map a), it substitues g.typeNeeds to map b
// to avoid concurrent iteration panika.
//
// then it swaps them, and so on until new types stop coming.
// It could be done better, i believe
//
// TODO: discover a better approach
func (g *generatorState) typeNeedsIterator() iter.Seq2[*types.Package, map[string]needs] {
	var a = g.typeNeeds
	var b = make(map[*types.Package]map[string]needs, len(g.typeNeeds))
	g.typeNeeds = b
	return func(yield func(*types.Package, map[string]needs) bool) {
		for {
			for pkg, sub := range a {
				if !yield(pkg, sub) {
					return
				}
			}
			for _, sub := range a {
				clear(sub)
			}
			var needMore int
			for _, sub := range b {
				needMore += len(sub)
			}
			if needMore == 0 {
				return
			}
			g.typeNeeds = a
			a = b
		}
	}
}
func (tm typeNeedsMap) getFromSeenMap(pkg *types.Package, obj types.Object) needs {
	var m = tm[pkg]
	if m == nil {
		return needs{}
	}
	var id = obj.Name()
	return m[id]
}
func (tm typeNeedsMap) insertToSeenMap(pkg *types.Package, obj types.Object, n needs) {
	var m = tm[pkg]
	if m == nil {
		m = make(map[string]needs, 10)
		tm[pkg] = m
	}
	m[obj.Name()] = n
}
func (g *generatorState) queueObjForScanner(obj types.Object) {
	var old = g.seenTypes.getFromSeenMap(obj.Pkg(), obj)
	if old.scanner {
		return
	} else {
		old.scanner = true
		g.seenTypes.insertToSeenMap(obj.Pkg(), obj, old)
	}
	var n = g.typeNeeds.getFromSeenMap(obj.Pkg(), obj)
	n.scanner = true
	g.typeNeeds.insertToSeenMap(obj.Pkg(), obj, n)
}
func (g *generatorState) queueObjForValuer(obj types.Object) {
	var old = g.seenTypes.getFromSeenMap(obj.Pkg(), obj)
	if old.valuer {
		return
	} else {
		old.valuer = true
		g.seenTypes.insertToSeenMap(obj.Pkg(), obj, old)
	}

	var n = g.typeNeeds.getFromSeenMap(obj.Pkg(), obj)
	n.valuer = true
	g.typeNeeds.insertToSeenMap(obj.Pkg(), obj, n)
}

// The semantic commentaries:
// - #[generate(Valuer)]
// - #[generate(Scanner)]
// - #[exclude(Valuer)]
// - #[exclude(Scanner)]
func (g *generatorState) queueByMacros(pkg *types.Package, decl ast.Decl) {
	if _, ok := g.generatedPackages[pkg]; !ok {
		// TODO: maybe issue a warning here? idk
		return
	}
	switch gd := decl.(type) {
	case *ast.GenDecl:
		if gd.Tok == token.TYPE {
			var scanner = false
			var valuer = false
			if gd.Doc != nil {
				for _, k := range gd.Doc.List {
					var komment = strings.TrimSpace(strings.TrimPrefix(k.Text, "//"))
					switch komment {
					case marker_excludeValuer:
						panic("TODO: error handling here; unneeded generate directives")
					case marker_excludeScanner:
						panic("TODO: error handling here; unneeded generate directives")
					case marker_generateScanner:
						if scanner {
							panic("TODO: error handling here; duplicate generate directives")
						} else {
							scanner = true
						}
					case marker_generateValuer:
						if valuer {
							panic("TODO: error handling here; duplicate generate directives")
						} else {
							valuer = true
						}
					}
				}
			}
			for _, spec := range gd.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Doc == nil {
						continue
					}
					var locscanner = scanner
					var locvaluer = valuer
					for _, k := range s.Doc.List {
						var komment = strings.TrimSpace(strings.TrimPrefix(k.Text, "//"))
						switch komment {
						case marker_generateScanner:
							if locscanner {
								panic("TODO: error handling here; duplicate generate directives")
							} else {
								locscanner = true
							}
						case marker_generateValuer:
							if locvaluer {
								panic("TODO: error handling here; duplicate generate directives")
							} else {
								locvaluer = true
							}
						case marker_excludeValuer:
							if !locvaluer {
								panic("TODO: error handling here; unneeded generate directives")
							} else {
								locvaluer = false
							}
						case marker_excludeScanner:
							if !locscanner {
								panic("TODO: error handling here; unneeded generate directives")
							} else {
								locscanner = false
							}
						}
					}
					var scope = pkg.Scope()
					if locscanner {
						g.queueObjForScanner(scope.Lookup(s.Name.Name))
					}
					if locvaluer {
						g.queueObjForValuer(scope.Lookup(s.Name.Name))
					}
				}
			}
		}

	}
}

func (p *Package) ParseSyntaxSync(g *generatorState) error {
	if p.Syntax != nil {
		return nil
	}
	var out = make([]*ast.File, len(p.CompiledGoFiles))
	var wg sync.WaitGroup
	wg.Add(len(p.CompiledGoFiles))
	// TODO: error handling
	for i, filename := range p.CompiledGoFiles {
		go func(i int, filename string) {
			var src, err = os.ReadFile(filename)
			if err != nil {
				wg.Done()
				panic(err)
			}
			if out[i], err = parser.ParseFile(
				g.fset, filename, src,
				parser.SkipObjectResolution|parser.ParseComments,
			); err != nil {
				wg.Done()
				panic(err)
			}
			wg.Done()
		}(i, filename)
	}
	wg.Wait()
	p.Syntax = out
	return nil
}

// the package should have the syntax parsed
func (g *generatorState) addPackage(pkg *Package) (err error) {
	if _, checked := g.checked[pkg]; checked {
		return
	}

	// TODO: do error reporting
	// visit all the children first, you know.
	var visit2 func(*Package)
	visit2 = func(lpkg *Package) {
		// get this shit out of here
		if lpkg.PkgPath == "unsafe" {
			return
		}
		var packages = slices.Collect(maps.Keys(lpkg.Imports))
		// parsePackages is an EXTREMELY expensive call - it shells away into other executable,
		// a go list thing. we should batch this as much as we can, and we can batch it more as
		// leaf nodes with different parents CAN be go list'ed simultaneously.
		//
		// TODO: do proper batching
		lpkg.Imports, _ = g.parsePackages(TypecheckingFlags{false}, packages...)
		for _, impp := range lpkg.Imports {
			visit2(impp)
		}
		// we dont really want to parse this syncronously, but i am too fucked
		// to do this properly too this time. TODO: do better, moron
		if err := lpkg.ParseSyntaxSync(g); err != nil {
			panic(err)
		}
		// we are traversing the graph depth-first, so in theory the leaves should be processed
		// before the non-leaves
		if err := g.typecheck(lpkg); err != nil {
			panic(err)
		}
	}
	visit2(pkg)
	return g.typecheck(pkg)
}

func (g *generatorState) typecheck(pkg *Package) error {
	if pkg.Types != nil {
		return nil
	}
	pkg.TypesInfo = types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
	pkg.Types = types.NewPackage(pkg.PkgPath, pkg.Name)

	var tc = types.Config{
		Importer:         pkg.importer(),
		IgnoreFuncBodies: true,
		Error: func(err error) {
			fmt.Printf("ERROR: %s\n", err.Error())
		},
	}

	return types.NewChecker(&tc, g.fset, pkg.Types, &pkg.TypesInfo).Files(pkg.Syntax)
}

func (p *Package) importer() importerFunc {
	return importerFunc(func(path string) (*types.Package, error) {
		if path == "unsafe" {
			return types.Unsafe, nil
		}
		var ipkg = p.Imports[path]
		if ipkg == nil {
			return nil, fmt.Errorf("no metadata for %s", path)
		}

		if ipkg.Types != nil && ipkg.Types.Complete() {
			return ipkg.Types, nil
		}
		fmt.Printf("FATAL: internal error: package %q without types was imported from %q", path, p.ID)
		panic("unreachable")
	})
}

func (g *generatorState) generateValuerForType(t types.Type, name string, prologue []ast.Stmt) error {
	switch v := t.(type) {
	case *types.Struct:
		var body = prologue

		var nf = v.NumFields()
		// TODO: move these to generatorState structure and cache to avoid excessive allocations
		//       (even arena memory needs to be used with caution, frater meus)
		var rec = g.I("v")
		var wcomma = g.Stmt(g.MethodCall(g.I("b"), "WriteByte")(g.AsLitChar(",")))
		var val = g.I("value")
		var wvalue = g.Stmt(g.MethodCall(g.I("b"), "WriteString")(val))
		body = append(body, g.VarDeclType("value", g.String))
		for i := range nf {
			var field = v.Field(i)

			body = append(body,
				g.renderValuer(
					field.Type(),
					g.Selector(rec, field.Name()),
					val))
			body = append(body, wvalue)
			if nf != i+1 {
				body = append(body, wcomma)
			}

		}
		body = append(body,
			g.Stmt(g.MethodCall(g.I("b"), "WriteByte")(g.AsLitChar(")"))),
			g.Return(g.MethodCall(g.I("b"), "String")(), g.Nil))

		g.CreateMethod("Value", "v", name, true)()(
			g.Param("t", g.ImportAndUse2("database/sql/driver", "driver", "Value")),
			g.Param("err", g.I("error")),
		)(body...)
	default:
		panic(fmt.Sprintf("unexpected types.Type: %#v", v))
	}
	return nil
}
func (g *generatorState) generateScanForType(t types.Type, name string, prologue []ast.Stmt) error {
	switch v := t.(type) {
	case *types.Slice:
		// TODO: check if the generated slice parser actually works
		var checkSrc = g.If(g.Or(
			g.Neq(g.IndexInt(g.I("sourceBytes"), 0), g.AsLitChar("(")),
			g.Neq(g.Index(g.I("sourceBytes"), g.SubConst(g.Len(g.I("source")), 1)), g.AsLitChar(")")),
		))(
			g.Return(
				g.ImportAndCall("fmt", "Errorf")(
					g.AsLit("bad source string: %q"),
					g.I("source"))))
		var declCur = g.VarDeclInit("cur")(g.AsLitInt(1))

		prologue = append(
			prologue, checkSrc, declCur,
			g.renderScanner(
				v,
				g.Star(g.I("v")),
				g.SliceExpr1(g.I("sourceBytes"), g.I("cur"))),
			g.Assign(g.Blank)(g.I("cur")),
			g.ReturnNil)

		g.CreateMethod("Scan", "v", name, true)(
			g.Param("thing", g.Any),
		)(g.Param("err", g.I("error")))(prologue...)
	case *types.Struct:
		var body, err = g.GenerateScanForStruct(name, v, prologue)
		if err != nil {
			return err
		}
		g.CreateMethod("Scan", "v", name, true)(
			g.Param("thing", g.Any),
		)(g.Param("err", g.I("error")))(body...)
	default:
		panic(fmt.Sprintf("unexpected types.Type: %#v", v))
	}
	return nil
}
func (g *generatorState) ParseModuleAndGenerate(f TypecheckingFlags, ppath string) (err error) {
	var packages map[string]*Package
	if packages, err = g.parsePackages(f, ppath); err != nil {
		return err
	}
	// for _, p := range packages {
	var p = packages[ppath]
	g.addPackage(p)
	g.generatedPackages[p.Types] = struct{}{}
	for _, file := range p.Syntax {
		for _, decl := range file.Decls {
			g.queueByMacros(p.Types, decl)
		}
	}
	// }
	var declVars = g.VarDeclInit("source", "ok")(g.TypeAssert("thing")(g.I("string")))
	var errorf = g.ImportAndCall("fmt", "Errorf")(g.AsLit("incompatible type: %+v"), g.I("thing"))
	var ifNotOkRet = g.If(g.Nok)(g.Return(errorf))
	var ifNotQuote = g.If(
		g.Eq(g.IndexInt(g.I("source"), 0), g.AsLitChar(`"`)),
	)(g.If2(
		g.Assign(
			g.I("source"), g.Blank, g.Err,
		)(g.FuncCall("__intrinsic_readString")(
			g.Slice1(g.unsafeBytesRender(g.I("source")), 1),
			g.AsLitInt(0))),
		g.ErrNotNil,
	)(g.ReturnErr))
	var scannerprologue = []ast.Stmt{
		declVars, ifNotOkRet, ifNotQuote,
		g.VarDeclInit("sourceBytes")(g.unsafeBytesRender(g.I("source"))),
	}
	var valuerprologue = []ast.Stmt{
		g.VarDeclType("b", g.ImportAndUse("strings", "Builder")),
		g.Stmt(g.MethodCall(g.I("b"), "WriteByte")(g.AsLitChar("("))),
	}

	var needsMap map[string]needs
	for g.nowPkg, needsMap = range g.typeNeedsIterator() {
		var pkgScope = g.nowPkg.Scope()
		for typename, need := range needsMap {
			var typ = pkgScope.Lookup(typename)
			if need.scanner {
				if err = g.generateScanForType(
					typ.(*types.TypeName).Type().Underlying(),
					typename, scannerprologue,
				); err != nil {
					return err
				}
			}
			if need.valuer {
				if err = g.generateValuerForType(
					typ.(*types.TypeName).Type().Underlying(),
					typename, valuerprologue,
				); err != nil {
					return err
				}
			}
		}
	}

	g.Import("database/sql/driver")
	g.Import("bytes")
	g.Import("time")
	g.Import("encoding/hex")
	g.Import("strconv")
	var actualFile = g.AsFile(p.Name)
	var out *os.File
	var outname = p.Dir + "/zz_scannervaluer.generated.go"
	if out, err = os.Create(outname); err != nil {
		panic(err.Error())
	}
	var config = printer.Config{Tabwidth: 4}
	if err := config.Fprint(out, g.fset, actualFile); err != nil {
		panic(err.Error())
	}
	out.WriteString(readString)
	out.WriteString(driverValueToString)
	return nil
}

func (g *generatorState) parsePackages(f TypecheckingFlags, ppath ...string) (map[string]*Package, error) {

	var resp = make(map[string]*Package)
	var toList = make([]string, 0, len(ppath))
	for _, name := range ppath {
		if p, ok := g.pkgs[name]; ok {
			resp[name] = p
		} else {
			toList = append(toList, name)
		}
	}
	if len(toList) == 0 {
		return resp, nil
	}
	for p, err := range runGoToolAndIterateResults(toList...) {
		if err != nil {
			return nil, err
		}
		if p.ImportPath == "" {
			return nil, fmt.Errorf("empty import path for %+v", p)
		}
		if filepath.IsAbs(p.ImportPath) && p.Error != nil {
			panic("TODO: implement this logic from packages")
		}

		g.seen[p.ImportPath] = p

		var pkg = Package{
			Name:            p.Name,
			ID:              p.ImportPath,
			Dir:             p.Dir,
			Target:          p.Target,
			GoFiles:         absolutize(p.Dir, p.GoFiles, p.CgoFiles),
			CompiledGoFiles: absolutize(p.Dir, p.CompiledGoFiles),
			OtherFiles:      absolutize(p.Dir, p.CFiles, p.CXXFiles, p.MFiles, p.HFiles, p.FFiles, p.SFiles, p.SwigFiles, p.SwigCXXFiles, p.SysoFiles),
			EmbedFiles:      absolutize(p.Dir, p.EmbedFiles),
			EmbedPatterns:   absolutize(p.Dir, p.EmbedPatterns),
			IgnoredFiles:    absolutize(p.Dir, p.IgnoredGoFiles, p.IgnoredOtherFiles),
			ForTest:         p.ForTest,
			Module:          p.Module,
		}
		if f.TypecheckCgo && len(pkg.CompiledGoFiles) != 0 {
			panic("implement cgo typechecking!")
		}

		if len(pkg.CompiledGoFiles) != 0 {
			var out = pkg.CompiledGoFiles[:0]
			for _, f := range pkg.CompiledGoFiles {
				// trash non-go and cached cgo files
				if ext := filepath.Ext(f); ext != ".go" && ext != "" {
					continue
				}
				out = append(out, f)
			}
			pkg.CompiledGoFiles = out
		}

		if i := strings.IndexByte(pkg.ID, ' '); i >= 0 {
			pkg.PkgPath = pkg.ID[:i]
		} else {
			pkg.PkgPath = pkg.ID
		}

		if pkg.PkgPath == "unsafe" {
			pkg.CompiledGoFiles = nil // ignore unsafe
		}

		// Assume go list emits only absolute paths for Dir.
		if p.Dir != "" && !filepath.IsAbs(p.Dir) {
			return nil, fmt.Errorf("relative Package.Dir %q received from go list", p.Dir)
		}

		if p.Export != "" && !filepath.IsAbs(p.Export) {
			pkg.ExportFile = filepath.Join(p.Dir, p.Export)
		} else {
			pkg.ExportFile = p.Export
		}

		var ids = make(map[string]bool)
		for _, id := range p.Imports {
			ids[id] = true
		}

		pkg.Imports = make(map[string]*Package)
		for path, id := range p.ImportMap {
			if id == "C" {
				delete(ids, id)
				continue
			}
			// non-identity import
			pkg.Imports[path] = &Package{ID: id}
			delete(ids, id)
		}

		// identity import
		for id := range ids {
			pkg.Imports[id] = &Package{ID: id}
		}

		// if !p.DepOnly {
		// 	roots = append(roots, pkg.ID)
		// }

		// some error thing
		if err := p.Error; err != nil && shouldAddFilenameFromError {
			fmt.Println(err)
		}

		if p.Error != nil {
			msg := strings.TrimSpace(p.Error.Err)
			if msg == "import cycle not allowed" && len(p.Error.ImportStack) != 0 {
				msg += fmt.Sprintf(": import stack: %v", p.Error.ImportStack)
			}
			pkg.Errors = append(pkg.Errors, Error{
				Pos:  p.Error.Pos,
				Msg:  msg,
				Kind: ListError,
			})
		}

		var pkgp = &pkg
		resp[pkgp.PkgPath] = pkgp
		g.pkgs[pkg.ID] = pkgp
	}

	return resp, nil
}

// TODO: deduce the output file from the passed packages;
// TODO: have a mechanism to compile several packages at the same time.
// TODO: have a mechanism to regenerate things into existing files.
//
//	That is, an autogenerated file should be parsed and then we should generate
//	only those things that are absent in the parsed implementation.
//	This way we may change the generated implementation and the generator will account
//	for that, not touching those functions that were generated already.
//	Moreover, a flag shall be introduced to force-regenerate things partly or fully.
//
// TODO: take all the ast-centric things out into a library and make it more general than
// whatever is needed here.
func main() {
	var g generatorState
	g.generatedPackages = make(map[*types.Package]struct{})
	g.typeNeeds = make(typeNeedsMap)
	g.seenTypes = make(typeNeedsMap)
	g.seen = make(map[string]*jsonPackage)
	g.pkgs = make(map[string]*Package)
	g.fset = token.NewFileSet()
	pggen.InitAstAcc(&g.AstAcc)
	var pathname = os.Args[1]
	var err error
	if err = g.ParseModuleAndGenerate(TypecheckingFlags{false}, pathname); err != nil {
		panic(err.Error())
	}
}

/*
go/types package integration layer with go/ast package.
Various functions to interact with/genererate AST for types.
*/

// returns a provided type as an ast.Expr usable, e.g., in a new, make, etc call.
// some heuristic is used here
func (g *generatorState) ValueTypeExpr(typ types.Type) ast.Expr {
	switch v := typ.(type) {
	case *types.Alias:
		return g.I(v.Obj().Name())
	case *types.Array:
		return g.ArrayType(g.ValueTypeExpr(v.Elem()), int(v.Len()))
	case *types.Basic:
		switch v.Kind() {
		case types.UntypedString:
			return g.String
		case types.Bool, types.UntypedBool:
			return g.Bool
		case types.Complex128, types.UntypedComplex:
			return g.Complex128
		case types.Complex64:
			return g.Complex64
		case types.Float32:
			return g.Float32
		case types.Float64, types.UntypedFloat:
			return g.Float64
		case types.Int, types.UntypedInt:
			return g.Int
		case types.Int16:
			return g.Int16
		case types.Int32, types.UntypedRune:
			return g.Int32
		case types.Int64:
			return g.Int64
		case types.Int8:
			return g.Int8
		case types.String:
			return g.String
		case types.Uint:
			return g.Uint
		case types.Uint8:
			return g.Uint8
		case types.Uint16:
			return g.Uint16
		case types.Uint32:
			return g.Uint32
		case types.Uint64:
			return g.Uint64
		case types.Uintptr:
			return g.Uintptr
		}
	case *types.Map:
		return g.MapType(g.ValueTypeExpr(v.Key()), g.ValueTypeExpr(v.Elem()))
	case *types.Named:
		var pkg = v.Obj().Pkg()
		if pkg == g.nowPkg {
			return g.I(v.Obj().Name())
		} else {
			return g.Selector(g.I(v.Obj().Pkg().Name()), v.Obj().Name())
		}
	case *types.Pointer:
		return g.Star(g.ValueTypeExpr(v.Elem()))
	case *types.Slice:
		return g.SliceType(g.ValueTypeExpr(v.Elem()))
	case *types.Struct:
		var fields = make([]*ast.Field, v.NumFields())
		var counter = 0
		for i := range v.Fields() {
			fields[counter] = g.Field(i.Name(), v.Tag(counter), g.ValueTypeExpr(i.Type()))
			counter++
		}
		return g.StructType(fields...)
	}
	panic(fmt.Sprintf("unreachable: unexpected types.Type: %#v", typ))
}

// renders a zero value expression literal for a given type.
func (g *generatorState) ZeroValue(typ types.Type) ast.Expr {
	switch v := typ.(type) {
	case *types.Alias:
		return g.ZeroValue(v.Origin())
	case *types.Basic:
		switch v.Kind() {
		case types.UntypedString, types.String:
			return g.EmptyString
		case types.Bool, types.UntypedBool:
			return g.False
		case types.Complex64, types.Complex128, types.UntypedComplex:
			return g.Cast(g.Complex(0, 0), g.ValueTypeExpr(v))
		case types.Float32, types.Float64, types.UntypedFloat:
			return g.Cast(g.AsFloatLit(0), g.ValueTypeExpr(v))
		case types.Int, types.UntypedInt, types.Int16,
			types.Int32, types.UntypedRune, types.Int64,
			types.Int8, types.Uint, types.Uint8,
			types.Uint16, types.Uint32, types.Uint64,
			types.Uintptr:
			return g.Cast(g.AsLitInt(0), g.ValueTypeExpr(v))
		case types.UnsafePointer:
			return g.Cast(g.Cast(g.AsLitInt(0), g.Uintptr), g.Selector(g.I("unsafe"), "Pointer"))
		default:
			panic("unreachable")
		}
	case *types.Named:
		var pkg = v.Obj().Pkg()
		if pkg != g.nowPkg {
			g.Import(pkg.Path())
		}
		return g.Cast(g.ZeroValue(v.Underlying()), g.ValueTypeExpr(v))
	case *types.Chan, *types.Interface,
		*types.Map, *types.Pointer, *types.Signature,
		*types.Slice:
		return g.Nil
	case *types.Struct, *types.Array:
		return g.CompositeLiteral(g.ValueTypeExpr(v))()
	default:
		panic(fmt.Sprintf("unexpected types.Type: %#v", v))
	}
}
