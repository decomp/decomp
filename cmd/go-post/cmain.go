package main

import (
	"go/ast"
	"go/token"
)

func init() {
	register(cmainFix)
}

var cmainFix = fix{
	name:     "cmain",
	date:     "2019-10-31",
	f:        cmain,
	desc:     "Wrap C main function to handle return statement status codes.",
	disabled: false,
}

func cmain(file *ast.File) bool {
	fixed := false

	// Locate the "main" function.
	cMainFunc, ok := findCMainFunc(file)
	if !ok {
		return false
	}

	// rename "main" to "c_main".
	cMainFunc.Name.Name = "c_main"
	fixed = true

	// Add "os" import if needed.
	addImport(file, "os")

	// add main wrapper function.
	mainFunc := &ast.FuncDecl{
		Name: ast.NewIdent("main"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// ret := int(c_main(0, nil))
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("ret")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun:  ast.NewIdent("int"),
						Args: []ast.Expr{
							&ast.CallExpr{
								Fun: ast.NewIdent("c_main"),
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.INT, Value: "0"},
									ast.NewIdent("nil"),
								},
							},
						},
					}},
				},
				// os.Exit(ret)
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: ast.NewIdent("os"),
							Sel: ast.NewIdent("Exit"),
						},
						Args: []ast.Expr{ast.NewIdent("ret")},
					},
				},
			},
		},
	}
	file.Decls = append(file.Decls, mainFunc)

	return fixed
}

// findCMainFunc attempts to locate the "main" function of the provided file.
// The boolean value is true if successful, and false otherwise.
func findCMainFunc(file *ast.File) (f *ast.FuncDecl, ok bool) {
	for _, f := range file.Decls {
		switch f := f.(type) {
		case *ast.FuncDecl:
			if f.Name.Name != "main" {
				continue
			}
			// Check that C main function has two (or more) parameters and a return
			// type.
			//
			//    int main(int argc, char **argv)
			//    int main(int argc, char **argv, char **envp)
			if len(f.Type.Params.List) >= 2 && f.Type.Results != nil {
				return f, true
			}
		}
	}
	return nil, false
}
