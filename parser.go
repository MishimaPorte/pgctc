package pggen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func ParseModule(dir string) error {
	var pkgs, err = parser.ParseDir(
		token.NewFileSet(),
		dir,
		nil,
		0)
	if err != nil {
		return err
	}

	for k, p := range pkgs {
		fmt.Printf("package %q parsed\n", k)
		for filename, f := range p.Files {
			fmt.Printf("filename %q parsed\n", filename)
			for _, decl := range f.Decls {
				switch v := decl.(type) {
				case *ast.GenDecl:
					if v.Tok == token.TYPE {
						for _, spec := range v.Specs {
							switch s := spec.(type) {
							case *ast.TypeSpec:
								var name = s.Name
								fmt.Printf("type %s parsed\n", name)
							}
						}
					}
				}
			}
		}
	}
	return nil
}
