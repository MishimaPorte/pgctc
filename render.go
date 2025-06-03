package pggen

import (
	"bytes"
	"fmt"
	"go/ast"
)

func setHas(s map[string]struct{}, name string) bool {
	var _, ok = s[name]
	return ok
}
func setPut(s map[string]struct{}, name string) {
	s[name] = struct{}{}
}

type rendererState struct {
	out bytes.Buffer

	imports       []string
	renderedTypes map[string]struct{}
	ErrorSet      ErrorSet
}

func (r *rendererState) render(t *ast.TypeSpec) {
	if setHas(r.renderedTypes, t.Name.Name) {
		return
	}

	switch v := t.Type.(type) {
	case *ast.ArrayType:
	case *ast.MapType:
	case *ast.StructType:
		r.renderStruct(v)
	default:
		r.ErrorSet.Errorf("unsupported type kind: %T", v)
	}
}

func (r *rendererState) renderStruct(t *ast.StructType) {
}

func (r *rendererState) fprintf(format string, a ...any) {
	fmt.Fprintf(&r.out, format, a...)
}
func (r *rendererState) addImport(a ...string) {
	r.imports = append(r.imports, a...)
}
