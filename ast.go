package pggen

import (
	"go/ast"
	"go/token"
	"slices"
	"strconv"
)

// A table of frequently-used identifiers and other ast nodes.
type Freqtable struct {
	Err, Ok, Nil                  *ast.Ident
	True, False                   *ast.Ident
	Any                           *ast.Ident
	String                        *ast.Ident
	EmptyString                   *ast.BasicLit
	Bool                          *ast.Ident
	Int                           *ast.Ident
	Int8, Int16, Int32, Int64     *ast.Ident
	Uint                          *ast.Ident
	Uint8, Uint16, Uint32, Uint64 *ast.Ident
	Uintptr                       *ast.Ident
	Float32, Float64              *ast.Ident
	Complex64, Complex128         *ast.Ident
	Byte                          *ast.Ident
	BytesType                     *ast.ArrayType
	AppendI                       *ast.Ident
	Blank                         *ast.Ident
	ReturnErr                     *ast.ReturnStmt
	ReturnNil                     *ast.ReturnStmt

	Nok       *ast.UnaryExpr
	ErrNotNil *ast.BinaryExpr
	ErrIsNil  *ast.BinaryExpr
}

func InitAstAcc(a *AstAcc) {
	a.Err = a.I("err")
	a.Ok = a.I("ok")
	a.Nil = a.I("nil")
	a.False = a.I("false")
	a.Bool = a.I("bool")
	a.String = a.I("string")
	a.EmptyString = a.AsLit("")
	a.True = a.I("true")
	a.Any = a.I("any")
	a.Int = a.I("int")
	a.Int8 = a.I("int8")
	a.Int16 = a.I("int16")
	a.Int32 = a.I("int32")
	a.Int64 = a.I("int64")
	a.Uint = a.I("uint")
	a.Uint8 = a.I("uint8")
	a.Uint16 = a.I("uint16")
	a.Uint32 = a.I("uint32")
	a.Uint64 = a.I("uint64")
	a.Uintptr = a.I("uintptr")
	a.Float32 = a.I("float32")
	a.Float64 = a.I("float64")
	a.Complex64 = a.I("complex64")
	a.Complex128 = a.I("complex128")
	a.Byte = a.I("byte")
	a.Blank = a.I("_")
	a.AppendI = a.I("append")
	a.BytesType = a.SliceType(a.Byte)
	a.Nok = a.Not(a.Ok)
	a.ErrNotNil = a.Neq(a.Err, a.Nil)
	a.ErrIsNil = a.Eq(a.Err, a.Nil)
	a.ReturnErr = a.Return(a.Err)
	a.ReturnNil = a.Return(a.Nil)
}

type AstAcc struct {
	a Arena

	funcs   []*ast.FuncDecl
	imports []string
	file    *ast.File

	Freqtable
}

func (a *AstAcc) Field(name, tag string, t ast.Expr) *ast.Field {
	var ret = Alloc[ast.Field](&a.a)
	var nameI = a.I(name)
	var nameIPtr = Alloc[*ast.Ident](&a.a)
	*nameIPtr = nameI
	ret.Names = AsSlice(nameIPtr)
	ret.Type = t
	ret.Tag = a.AsLit(tag)
	return ret
}
func (a *AstAcc) StructType(fields ...*ast.Field) *ast.StructType {
	var ret = Alloc[ast.StructType](&a.a)
	var flist = Alloc[ast.FieldList](&a.a)
	ret.Fields = flist
	flist.List = Clone(&a.a, fields).Slice()
	return ret
}

func (a *AstAcc) MapType(k, v ast.Expr) *ast.MapType {
	var ret = Alloc[ast.MapType](&a.a)
	ret.Key = k
	ret.Value = v
	return ret
}

func (a *AstAcc) SliceType(sub ast.Expr) *ast.ArrayType {
	var ret = Alloc[ast.ArrayType](&a.a)
	ret.Elt = sub
	return ret
}
func (a *AstAcc) ArrayType(sub ast.Expr, length int) *ast.ArrayType {
	var ret = Alloc[ast.ArrayType](&a.a)
	ret.Elt = sub
	ret.Len = a.AsLitInt(length)
	return ret
}
func (a *AstAcc) asIdents(s string) []*ast.Ident {
	var id = a.I(s)
	var ids = Alloc[*ast.Ident](&a.a)
	*ids = id
	return AsSlice(ids)
}
func (a *AstAcc) AsFloatLit(s float64) *ast.BasicLit {
	var id = Alloc[ast.BasicLit](&a.a)
	id.Value = a.a.Strclone(strconv.FormatFloat(s, 'f', -1, 64))
	id.Kind = token.FLOAT
	return id
}
func (a *AstAcc) Complex(r, i float64) *ast.CallExpr {
	return a.FuncCall("complex")(a.AsFloatLit(r), a.AsFloatLit(i))
}
func (a *AstAcc) AsLitInt(s int) *ast.BasicLit {
	var id = Alloc[ast.BasicLit](&a.a)
	id.Value = a.a.Strclone(strconv.Itoa(s))
	id.Kind = token.INT
	return id
}
func (a *AstAcc) AsLitChar(s string) *ast.BasicLit {
	var id = Alloc[ast.BasicLit](&a.a)
	id.Value = a.a.Strclone(strconv.QuoteRune(rune(s[0])))
	id.Kind = token.CHAR
	return id
}
func (a *AstAcc) AsLit(s string) *ast.BasicLit {
	var id = Alloc[ast.BasicLit](&a.a)
	id.Value = a.a.Strclone(strconv.Quote(s))
	id.Kind = token.STRING
	return id
}
func (a *AstAcc) I(s string) *ast.Ident {
	var id = Alloc[ast.Ident](&a.a)
	id.Name = a.a.Strclone(s)
	return id
}
func (a *AstAcc) AsFile(packageName string) *ast.File {
	if a.file == nil {
		a.file = Alloc[ast.File](&a.a)
		a.file.Name = a.I(packageName)
	}
	var decls = Allocn[ast.Decl](&a.a, uintptr(len(a.funcs)+1))
	var decl = Alloc[ast.GenDecl](&a.a)
	decl.Tok = token.IMPORT
	// This abhorrent shit is needed to make format.Node sort imports.
	// It for some reason unknown to men cannot just take len(decl.Specs)
	// and this way reason about whether imports are to be sorted; it cannot for
	// some convoluted criteria just assume that imports ALWAYS need to be sorted
	// (they are). It takes decl.Lparen.IsValid() and if it is valid it sorts
	// the imports. It does not use the value for anything; it is bad.
	decl.Lparen = 1
	*decls.RefAt(0) = decl
	var specs = Allocn[ast.Spec](&a.a, uintptr(len(a.imports)))
	decl.Specs = specs.Slice()
	var impSpecPtrs = Allocn[*ast.ImportSpec](&a.a, uintptr(len(a.imports)))
	var impSpecs = Allocn[ast.ImportSpec](&a.a, uintptr(len(a.imports)))
	for i, imp := range a.imports {
		*impSpecPtrs.RefAt(i) = impSpecs.RefAt(i)
		impSpecs.RefAt(i).Path = a.AsLit(imp)
		*specs.RefAt(i) = impSpecs.RefAt(i)
	}
	a.file.Imports = impSpecPtrs.Slice()
	for i, fuc := range a.funcs {
		*decls.RefAt(i + 1) = fuc
	}
	a.file.Decls = decls.Slice()
	return a.file
}

type MethDecl struct {
	D *ast.FuncDecl
}

func (a *AstAcc) Param(name string, t ast.Expr) *ast.Field {
	var f = Alloc[ast.Field](&a.a)
	var i = a.I(name)
	var pI = Alloc[*ast.Ident](&a.a)
	*pI = i
	f.Names = AsSlice(pI)
	f.Type = t
	return f
}

func (a *AstAcc) VarDeclType(name string, typeast ast.Expr) ast.Stmt {
	var vard = Alloc[ast.GenDecl](&a.a)
	vard.Tok = token.VAR
	var varspec = Alloc[ast.ValueSpec](&a.a)
	varspec.Names = a.asIdents(name)
	varspec.Type = typeast
	var specs = Alloc[ast.Spec](&a.a)
	*specs = varspec
	vard.Specs = AsSlice(specs)
	var spec = Alloc[ast.DeclStmt](&a.a)
	spec.Decl = vard
	return spec
}

// slice[x:y:z]
func (a *AstAcc) Slice3(e ast.Expr, x, y, z int) *ast.SliceExpr {
	var node = Alloc[ast.SliceExpr](&a.a)
	node.X = e
	node.Low = a.AsLitInt(x)
	node.High = a.AsLitInt(y)
	node.Max = a.AsLitInt(z)
	return node
}

// slice[x:y]
func (a *AstAcc) Slice2(e ast.Expr, x, y int) *ast.SliceExpr {
	var node = Alloc[ast.SliceExpr](&a.a)
	node.X = e
	node.Low = a.AsLitInt(x)
	node.High = a.AsLitInt(y)
	return node
}

// slice[x:]
func (a *AstAcc) SliceExpr1(e, x ast.Expr) *ast.SliceExpr {
	var node = Alloc[ast.SliceExpr](&a.a)
	node.X = e
	node.Low = x
	return node
}

// slice[x:]
func (a *AstAcc) Slice1(e ast.Expr, x int) *ast.SliceExpr {
	var node = Alloc[ast.SliceExpr](&a.a)
	node.X = e
	node.Low = a.AsLitInt(x)
	return node
}
func (a *AstAcc) Index(e, index ast.Expr) *ast.IndexExpr {
	var node = Alloc[ast.IndexExpr](&a.a)
	node.X = e
	node.Index = index
	return node
}
func (a *AstAcc) IndexInt(e ast.Expr, index int) *ast.IndexExpr {
	var node = Alloc[ast.IndexExpr](&a.a)
	node.X = e
	node.Index = a.AsLitInt(index)
	return node
}
func (a *AstAcc) TypeAssert(varname string) func(t ast.Expr) *ast.TypeAssertExpr {
	var ta = Alloc[ast.TypeAssertExpr](&a.a)
	ta.X = a.I(varname)
	return func(t ast.Expr) *ast.TypeAssertExpr {
		ta.Type = t
		return ta
	}
}

// items = append(items, xs)
func (a *AstAcc) Append(items ast.Expr, xs ...ast.Expr) *ast.AssignStmt {
	var fc = Alloc[ast.CallExpr](&a.a)
	fc.Fun = a.AppendI
	var args = Allocn[ast.Expr](&a.a, uintptr(len(xs)+1))
	copy(args.Slice()[1:], xs)
	*args.RefAt(0) = items
	fc.Args = args.Slice()
	return a.Assign(items)(fc)
}
func (a *AstAcc) New(x ast.Expr) *ast.CallExpr {
	return a.FuncCall("new")(x)
}
func (a *AstAcc) Len(e ast.Expr) *ast.CallExpr {
	return a.FuncCall("len")(e)
}
func (a *AstAcc) Make2(elem, length ast.Expr) *ast.CallExpr {
	return a.FuncCall("make")(a.SliceType(elem), length)
}
func (a *AstAcc) Make3(elem, length, capacity ast.Expr) *ast.CallExpr {
	return a.FuncCall("make")(a.SliceType(elem), length, capacity)
}
func (a *AstAcc) Selector(e ast.Expr, s ...string) *ast.SelectorExpr {
	var now = Alloc[ast.SelectorExpr](&a.a)
	now.X = e
	for i, x := range s {
		now.Sel = a.I(x)
		if i == len(s)-1 {
			return now
		}
		var newNow = Alloc[ast.SelectorExpr](&a.a)
		newNow.X = now
		now = newNow
	}
	panic("unreachable")
}
func (a *AstAcc) ImportAndUse2(path, pkgname, id string) *ast.SelectorExpr {
	a.Import(path)
	var selector = Alloc[ast.SelectorExpr](&a.a)
	selector.X = a.I(pkgname)
	selector.Sel = a.I(id)
	return selector
}
func (a *AstAcc) ImportAndUse(r, id string) *ast.SelectorExpr {
	a.Import(r)
	var selector = Alloc[ast.SelectorExpr](&a.a)
	selector.X = a.I(r)
	selector.Sel = a.I(id)
	return selector
}
func (a *AstAcc) ImportAndCall(r, funcname string) func(args ...ast.Expr) *ast.CallExpr {
	var fc = Alloc[ast.CallExpr](&a.a)
	fc.Fun = a.ImportAndUse(r, funcname)
	return func(args ...ast.Expr) *ast.CallExpr {
		var exprs = Allocn[ast.Expr](&a.a, uintptr(len(args)))
		copy(exprs.Slice(), args)
		fc.Args = exprs.Slice()
		return fc
	}
}
func (a *AstAcc) MethodCall(r ast.Expr, funcname string) func(args ...ast.Expr) *ast.CallExpr {
	var fc = Alloc[ast.CallExpr](&a.a)
	var selector = Alloc[ast.SelectorExpr](&a.a)
	selector.X = r
	selector.Sel = a.I(funcname)
	fc.Fun = selector
	return func(args ...ast.Expr) *ast.CallExpr {
		var exprs = Allocn[ast.Expr](&a.a, uintptr(len(args)))
		copy(exprs.Slice(), args)
		fc.Args = exprs.Slice()
		return fc
	}
}
func (a *AstAcc) FuncCall(name string) func(args ...ast.Expr) *ast.CallExpr {
	var fc = Alloc[ast.CallExpr](&a.a)
	fc.Fun = a.I(name)
	return func(args ...ast.Expr) *ast.CallExpr {
		var exprs = Allocn[ast.Expr](&a.a, uintptr(len(args)))
		copy(exprs.Slice(), args)
		fc.Args = exprs.Slice()
		return fc
	}
}
func (a *AstAcc) Return(exprs ...ast.Expr) *ast.ReturnStmt {
	var ret = Alloc[ast.ReturnStmt](&a.a)
	ret.Results = Clone(&a.a, exprs).Slice()
	return ret
}
func (a *AstAcc) Add(xs ...ast.Expr) *ast.BinaryExpr {
	var now = Alloc[ast.BinaryExpr](&a.a)
	now.X = xs[0]
	for i, x := range xs[1:] {
		now.Op = token.ADD
		now.Y = x
		if i == len(xs)-2 {
			return now
		}
		var newNow = Alloc[ast.BinaryExpr](&a.a)
		newNow.X = now
		now = newNow
	}
	panic("unreachable")
}
func (a *AstAcc) Star(x ast.Expr) *ast.StarExpr {
	var ret = Alloc[ast.StarExpr](&a.a)
	ret.X = x
	return ret
}
func (a *AstAcc) SubConst(x ast.Expr, y int) *ast.BinaryExpr {
	var unary = Alloc[ast.BinaryExpr](&a.a)
	unary.X = x
	unary.Op = token.SUB
	unary.Y = a.AsLitInt(y)
	return unary
}
func (a *AstAcc) And(x, y ast.Expr) *ast.BinaryExpr {
	var unary = Alloc[ast.BinaryExpr](&a.a)
	unary.X = x
	unary.Op = token.AND
	unary.Y = y
	return unary
}
func (a *AstAcc) Or(x, y ast.Expr) *ast.BinaryExpr {
	var unary = Alloc[ast.BinaryExpr](&a.a)
	unary.X = x
	unary.Op = token.LOR
	unary.Y = y
	return unary
}
func (a *AstAcc) Lt(x, y ast.Expr) *ast.BinaryExpr {
	var unary = Alloc[ast.BinaryExpr](&a.a)
	unary.X = x
	unary.Op = token.LSS
	unary.Y = y
	return unary
}
func (a *AstAcc) Neq(x, y ast.Expr) *ast.BinaryExpr {
	var unary = Alloc[ast.BinaryExpr](&a.a)
	unary.X = x
	unary.Op = token.NEQ
	unary.Y = y
	return unary
}
func (a *AstAcc) Eq(x, y ast.Expr) *ast.BinaryExpr {
	var unary = Alloc[ast.BinaryExpr](&a.a)
	unary.X = x
	unary.Op = token.EQL
	unary.Y = y
	return unary
}
func (a *AstAcc) Not(expr ast.Expr) *ast.UnaryExpr {
	var unary = Alloc[ast.UnaryExpr](&a.a)
	unary.Op = token.NOT
	unary.X = expr
	return unary
}
func (a *AstAcc) For1(cond ast.Expr) func(...ast.Stmt) *ast.ForStmt {
	var fors = Alloc[ast.ForStmt](&a.a)
	fors.Cond = cond
	return func(s ...ast.Stmt) *ast.ForStmt {
		fors.Body = a.Block(s...)
		return fors
	}
}
func (a *AstAcc) For2(init ast.Stmt, cond ast.Expr) func(...ast.Stmt) *ast.ForStmt {
	var fors = Alloc[ast.ForStmt](&a.a)
	fors.Init = init
	fors.Cond = cond
	return func(s ...ast.Stmt) *ast.ForStmt {
		fors.Body = Alloc[ast.BlockStmt](&a.a)
		fors.Body.List = Clone(&a.a, s).Slice()
		return fors
	}
}
func (a *AstAcc) For3(init, post ast.Stmt, cond ast.Expr) func(...ast.Stmt) *ast.ForStmt {
	var fors = Alloc[ast.ForStmt](&a.a)
	fors.Init = init
	fors.Post = post
	fors.Cond = cond
	return func(s ...ast.Stmt) *ast.ForStmt {
		fors.Body = Alloc[ast.BlockStmt](&a.a)
		fors.Body.List = Clone(&a.a, s).Slice()
		return fors
	}
}
func (a *AstAcc) Range2Def(k, v, val ast.Expr) func(...ast.Stmt) *ast.RangeStmt {
	var fors = Alloc[ast.RangeStmt](&a.a)
	fors.Key = k
	fors.Value = v
	fors.X = val
	fors.Tok = token.DEFINE
	return func(s ...ast.Stmt) *ast.RangeStmt {
		fors.Body = Alloc[ast.BlockStmt](&a.a)
		fors.Body.List = Clone(&a.a, s).Slice()
		return fors
	}
}
func (a *AstAcc) CompositeLiteral(t ast.Expr) func(...ast.Expr) *ast.CompositeLit {
	var cl = Alloc[ast.CompositeLit](&a.a)
	cl.Type = t
	return func(e ...ast.Expr) *ast.CompositeLit {
		if len(e) == 0 {
			return cl
		}
		cl.Elts = Clone(&a.a, e).Slice()
		return cl
	}
}
func (a *AstAcc) Stmt(e ast.Expr) *ast.ExprStmt {
	var ret = Alloc[ast.ExprStmt](&a.a)
	ret.X = e
	return ret
}
func (a *AstAcc) If(cond ast.Expr) func(then ...ast.Stmt) *ast.IfStmt {
	var node = Alloc[ast.IfStmt](&a.a)
	node.Cond = cond
	return func(then ...ast.Stmt) *ast.IfStmt {
		var block = Alloc[ast.BlockStmt](&a.a)
		block.List = Clone(&a.a, then).Slice()
		node.Body = block
		return node
	}
}
func (a *AstAcc) IfElse2(init ast.Stmt, cond ast.Expr) func(then ...ast.Stmt) func(elses ...ast.Stmt) *ast.IfStmt {
	var node = Alloc[ast.IfStmt](&a.a)
	node.Init = init
	node.Cond = cond
	return func(then ...ast.Stmt) func(elses ...ast.Stmt) *ast.IfStmt {
		var block = Alloc[ast.BlockStmt](&a.a)
		block.List = Clone(&a.a, then).Slice()
		node.Body = block
		return func(elses ...ast.Stmt) *ast.IfStmt {
			if len(elses) == 1 {
				node.Else = elses[0]
			} else {
				var block = Alloc[ast.BlockStmt](&a.a)
				block.List = Clone(&a.a, elses).Slice()
				node.Else = block
			}
			return node
		}
	}
}
func (a *AstAcc) IfElse(cond ast.Expr) func(then ...ast.Stmt) func(elses ...ast.Stmt) *ast.IfStmt {
	var node = Alloc[ast.IfStmt](&a.a)
	node.Cond = cond
	return func(then ...ast.Stmt) func(elses ...ast.Stmt) *ast.IfStmt {
		var block = Alloc[ast.BlockStmt](&a.a)
		block.List = Clone(&a.a, then).Slice()
		node.Body = block
		return func(elses ...ast.Stmt) *ast.IfStmt {
			if len(elses) == 1 {
				node.Else = elses[0]
			} else {
				var block = Alloc[ast.BlockStmt](&a.a)
				block.List = Clone(&a.a, elses).Slice()
				node.Else = block
			}
			return node
		}
	}
}
func (a *AstAcc) If2(init ast.Stmt, cond ast.Expr) func(...ast.Stmt) *ast.IfStmt {
	var node = Alloc[ast.IfStmt](&a.a)
	node.Cond = cond
	node.Init = init
	return func(then ...ast.Stmt) *ast.IfStmt {
		var block = Alloc[ast.BlockStmt](&a.a)
		block.List = Clone(&a.a, then).Slice()
		node.Body = block
		return node
	}
}

func (a *AstAcc) Cast(e ast.Expr, to ast.Expr) *ast.CallExpr {
	var fc = Alloc[ast.CallExpr](&a.a)
	fc.Fun = to
	var exprs = Alloc[ast.Expr](&a.a)
	*exprs = e
	fc.Args = AsSlice(exprs)
	return fc
}
func (a *AstAcc) If3(init ast.Stmt, cond ast.Expr, then ...ast.Stmt) func(elses ...ast.Stmt) *ast.IfStmt {
	var node = Alloc[ast.IfStmt](&a.a)
	node.Cond = cond
	node.Init = init
	var block = Alloc[ast.BlockStmt](&a.a)
	block.List = Clone(&a.a, then).Slice()
	node.Body = block
	return func(elses ...ast.Stmt) *ast.IfStmt {
		if len(elses) == 1 {
			node.Else = elses[0]
		} else {
			var block = Alloc[ast.BlockStmt](&a.a)
			block.List = Clone(&a.a, elses).Slice()
			node.Else = block
		}
		return node
	}
}
func (a *AstAcc) Assign(lhs ...ast.Expr) func(rhs ...ast.Expr) *ast.AssignStmt {
	var ass = Alloc[ast.AssignStmt](&a.a)
	ass.Tok = token.ASSIGN
	var lhses = Allocn[ast.Expr](&a.a, uintptr(len(lhs)))
	copy(lhses.Slice(), lhs)
	ass.Lhs = lhses.Slice()
	return func(rhs ...ast.Expr) *ast.AssignStmt {
		var rhses = Allocn[ast.Expr](&a.a, uintptr(len(rhs)))
		copy(rhses.Slice(), rhs)
		ass.Rhs = rhses.Slice()
		return ass
	}
}
func (a *AstAcc) AssignStmt(names ...string) func(...ast.Expr) *ast.AssignStmt {
	var lhs = Allocn[ast.Expr](&a.a, uintptr(len(names)))
	for i := range names {
		var ident = a.I(names[i])
		*lhs.RefAt(i) = ident
	}
	var spec = Alloc[ast.AssignStmt](&a.a)
	spec.Lhs = lhs.Slice()
	spec.Tok = token.DEFINE
	return func(e ...ast.Expr) *ast.AssignStmt {
		spec.Rhs = Clone(&a.a, e).Slice()
		return spec
	}
}
func (a *AstAcc) Block(stmts ...ast.Stmt) *ast.BlockStmt {
	var block = Alloc[ast.BlockStmt](&a.a)
	block.List = Clone(&a.a, stmts).Slice()
	return block
}
func (a *AstAcc) VarDeclInit(names ...string) func(...ast.Expr) *ast.DeclStmt {
	var vard = Alloc[ast.GenDecl](&a.a)
	vard.Tok = token.VAR
	var aspecs = Alloc[ast.Spec](&a.a)
	vard.Specs = AsSlice(aspecs)
	var varspec = Alloc[ast.ValueSpec](&a.a)
	*aspecs = varspec
	var idents = Allocn[*ast.Ident](&a.a, uintptr(len(names)))
	varspec.Names = idents.Slice()
	for i := range names {
		var ident = a.I(names[i])
		*idents.RefAt(i) = ident
	}
	var spec = Alloc[ast.DeclStmt](&a.a)
	spec.Decl = vard
	return func(e ...ast.Expr) *ast.DeclStmt {
		varspec.Values = Clone(&a.a, e).Slice()
		return spec
	}
}

func (a *AstAcc) ValueSpec(names ...string) func(ast.Expr) *ast.ValueSpec {
	var spec = Alloc[ast.ValueSpec](&a.a)
	var nameIds = Allocn[*ast.Ident](&a.a, uintptr(len(names)))
	for i := range nameIds.Length {
		*nameIds.RefAt(i) = a.I(names[i])
	}
	spec.Names = nameIds.Slice()
	return func(t ast.Expr) *ast.ValueSpec {
		spec.Type = t
		return spec
	}
}

func (a *AstAcc) VarDecl2(specs ...*ast.ValueSpec) *ast.DeclStmt {
	var vard = Alloc[ast.GenDecl](&a.a)
	vard.Tok = token.VAR
	var aspecs = Allocn[ast.Spec](&a.a, uintptr(len(specs)))
	vard.Specs = aspecs.Slice()
	for i := range specs {
		*aspecs.RefAt(i) = specs[i]
	}
	var spec = Alloc[ast.DeclStmt](&a.a)
	spec.Decl = vard
	return spec
}

func (a *AstAcc) Case(exprs ...ast.Expr) func(...ast.Stmt) *ast.CaseClause {
	var c = Alloc[ast.CaseClause](&a.a)
	if len(exprs) != 0 {
		c.List = Clone(&a.a, exprs).Slice()
	}
	return func(stmts ...ast.Stmt) *ast.CaseClause {
		c.Body = Clone(&a.a, stmts).Slice()
		return c
	}
}
func (a *AstAcc) Switch(value ast.Expr) func(...*ast.CaseClause) *ast.SwitchStmt {
	var swi = Alloc[ast.SwitchStmt](&a.a)
	swi.Tag = value
	return func(cc ...*ast.CaseClause) *ast.SwitchStmt {
		swi.Body = Alloc[ast.BlockStmt](&a.a)
		var list = Allocn[ast.Stmt](&a.a, uintptr(len(cc)))
		for i, item := range cc {
			*list.RefAt(i) = item
		}
		swi.Body.List = list.Slice()
		return swi
	}
}
func (a *AstAcc) VarDecl(name, typename string, isPtr bool) *ast.DeclStmt {
	var vard = Alloc[ast.GenDecl](&a.a)
	vard.Tok = token.VAR
	var varspec = Alloc[ast.ValueSpec](&a.a)
	varspec.Names = a.asIdents(name)
	var vartype = a.I(typename)
	if isPtr {
		var vartypeStar = Alloc[ast.StarExpr](&a.a)
		vartypeStar.X = vartype
		varspec.Type = vartypeStar
	} else {
		varspec.Type = vartype
	}
	var specs = Alloc[ast.Spec](&a.a)
	*specs = varspec
	vard.Specs = AsSlice(specs)
	var spec = Alloc[ast.DeclStmt](&a.a)
	spec.Decl = vard
	return spec
}

func (a *AstAcc) Import(pname string) {
	// maybe employ a hashset if the thing is too slow in O(n)
	// though it should not be since n is small at any time
	if !slices.Contains(a.imports, pname) {
		a.imports = append(a.imports, pname)
	}
}

func (a *AstAcc) CreateMethod(
	mName string,
	rName string, rType string, rIsPtr bool,
) func(...*ast.Field) func(...*ast.Field) func(...ast.Stmt) *ast.FuncDecl {
	var decl = Alloc[ast.FuncDecl](&a.a)
	decl.Name = a.I(mName)

	decl.Recv = Alloc[ast.FieldList](&a.a)
	var recv = Alloc[ast.Field](&a.a)
	var pRecv = Alloc[*ast.Field](&a.a)
	*pRecv = recv
	var nameP = Alloc[*ast.Ident](&a.a)
	*nameP = a.I(rName)
	recv.Names = AsSlice(nameP)
	var typ ast.Expr
	if rIsPtr {
		var star = Alloc[ast.StarExpr](&a.a)
		star.X = a.I(rType)
		typ = star
	} else {
		typ = a.I(rType)
	}
	recv.Type = typ
	decl.Recv = Alloc[ast.FieldList](&a.a)
	*decl.Recv = ast.FieldList{
		List: AsSlice(pRecv),
	}
	decl.Type = Alloc[ast.FuncType](&a.a)
	return func(params ...*ast.Field) func(...*ast.Field) func(...ast.Stmt) *ast.FuncDecl {
		if len(params) != 0 {
			decl.Type.Params = Alloc[ast.FieldList](&a.a)
			decl.Type.Params.List = Clone(&a.a, params).Slice()
		}
		return func(returns ...*ast.Field) func(...ast.Stmt) *ast.FuncDecl {
			if len(returns) != 0 {
				decl.Type.Results = Alloc[ast.FieldList](&a.a)
				decl.Type.Results.List = Clone(&a.a, returns).Slice()
			}
			return func(stmts ...ast.Stmt) *ast.FuncDecl {
				if len(stmts) != 0 {
					decl.Body = Alloc[ast.BlockStmt](&a.a)
					decl.Body.List = Clone(&a.a, stmts).Slice()
				}
				a.funcs = append(a.funcs, decl)
				return decl
			}
		}
	}
}

func (a *AstAcc) IfErrNotNilReturnErr(val ast.Expr, as ...ast.Expr) *ast.IfStmt {
	return a.If2(
		a.Assign(a.Err)(val),
		a.ErrNotNil,
	)(a.ReturnErr)
}
func (a *AstAcc) IntFromString(s string) ast.Expr {
	switch s {
	case "int":
		return a.Int
	case "int8":
		return a.Int8
	case "int16":
		return a.Int16
	case "int32":
		return a.Int32
	case "int64":
		return a.Int64
	case "uint":
		return a.Uint
	case "uint8":
		return a.Uint8
	case "uint16":
		return a.Uint16
	case "uint32":
		return a.Uint32
	case "uint64":
		return a.Uint64
	case "uintptr":
		return a.Uintptr
	default:
		panic("bad int in IntFromString")
	}
}
