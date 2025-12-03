package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func (t *Transpiler) Transpile(node ast.Node) error {
	switch n := node.(type) {
	
	// --- Estrutura de Arquivo ---
	case *ast.File:
		t.writeLine("package " + n.Name.Name)
		t.write("\n")
		
		if len(n.Imports) > 0 {
			for _, imp := range n.Imports {
				path := strings.Trim(imp.Path.Value, "\"")
				if path != "fmt" { 
					t.writeLine("import " + path)
				}
			}
			t.write("\n")
		}
		
		// Declarações Globais (Funções, Variáveis Globais)
		for _, decl := range n.Decls {
			t.Transpile(decl)
			t.write("\n\n") 
		}

	// --- Funções ---
	case *ast.FuncDecl:
		t.writeIndent()
		
		// Visibilidade
		if ast.IsExported(n.Name.Name) {
			t.write("public ")
		} else {
			t.write("internal ")
		}
		
		t.write("fun " + n.Name.Name + "(")
		
		// Parâmetros
		if n.Type.Params != nil {
			for i, field := range n.Type.Params.List {
				if i > 0 { t.write(", ") }
				typeName := t.resolveType(field.Type)
				for j, name := range field.Names {
					if j > 0 { t.write(", ") }
					t.write(name.Name + ": " + typeName)
				}
			}
		}
		t.write(")")

		// Retorno
		if n.Type.Results != nil && len(n.Type.Results.List) > 0 {
			retType := t.resolveType(n.Type.Results.List[0].Type)
			t.write(": " + retType)
		}
		
		t.write(" ")
		t.Transpile(n.Body)

	// --- Blocos e Declarações ---
	case *ast.BlockStmt:
		t.write("{\n")
		t.indent()
		for _, stmt := range n.List {
			t.writeIndent()
			t.Transpile(stmt)
			t.write("\n")
		}
		t.unindent()
		t.writeIndent()
		t.write("}")

	case *ast.GenDecl: // var ou const
		if n.Tok == token.VAR || n.Tok == token.CONST {
			keyword := "var"
			if n.Tok == token.CONST { keyword = "val" }
			
			for _, spec := range n.Specs {
				vspec := spec.(*ast.ValueSpec)
				typeName := ""
				if vspec.Type != nil {
					typeName = ": " + t.resolveType(vspec.Type)
				}
				for i, name := range vspec.Names {
					t.write(keyword + " " + name.Name + typeName)
					if i < len(vspec.Values) {
						t.write(" = ")
						t.Transpile(vspec.Values[i])
					}
				}
			}
		}

	// --- Atribuição ---
	case *ast.AssignStmt:
		for i, lhs := range n.Lhs {
			if n.Tok == token.DEFINE {
				t.write("var ") 
			}
			t.Transpile(lhs)
			if i < len(n.Rhs) {
				t.write(" = ")
				t.Transpile(n.Rhs[i])
			}
		}

	case *ast.ExprStmt:
		t.Transpile(n.X)

	// --- Chamadas de Função (I/O) ---
	case *ast.CallExpr:
		isHandled := false
		if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "fmt" {
				
				// 1. Saída (Print)
				if strings.HasPrefix(sel.Sel.Name, "Print") {
					cmd := "print"
					if sel.Sel.Name == "Println" { cmd = "println" }
					t.write(cmd)
					isHandled = true
				} 
				
				// 2. Entrada (Scan)
				if strings.HasPrefix(sel.Sel.Name, "Scan") {
					if len(n.Args) > 0 {
						t.Transpile(n.Args[0]) 
						t.write(" = readln()") 
						t.write(" // !! Converter tipo se necessario")
						isHandled = true
						return nil 
					}
				}
			}
		}

		if !isHandled {
			t.Transpile(n.Fun)
		}

		t.write("(")
		for i, arg := range n.Args {
			if i > 0 { t.write(", ") }
			t.Transpile(arg)
		}
		t.write(")")

	// --- Controle de Fluxo (IF) ---
	case *ast.IfStmt:
		if n.Init != nil {
			t.write("run {\n")
			t.indent()
			t.writeIndent()
			t.Transpile(n.Init)
			t.write("\n")
			t.writeIndent()
		}

		t.write("if (")
		t.Transpile(n.Cond)
		t.write(") ")
		t.Transpile(n.Body)

		if n.Else != nil {
			t.write(" else ")
			t.Transpile(n.Else)
		}

		if n.Init != nil {
			t.unindent()
			t.write("\n")
			t.writeIndent()
			t.write("}")
		}

	// --- Controle de Fluxo (FOR -> WHILE) ---
	case *ast.ForStmt:
		t.write("run {\n")
		t.indent()
		
		if n.Init != nil {
			t.writeIndent()
			t.Transpile(n.Init)
			t.write("\n")
		}

		t.writeIndent()
		t.write("while (")
		if n.Cond != nil {
			t.Transpile(n.Cond)
		} else {
			t.write("true")
		}
		t.write(") {\n")
		
		t.indent()
		for _, stmt := range n.Body.List {
			t.writeIndent()
			t.Transpile(stmt)
			t.write("\n")
		}
		
		if n.Post != nil {
			t.writeIndent()
			t.Transpile(n.Post)
			t.write("\n")
		}
		
		t.unindent()
		t.writeIndent()
		t.write("}\n") 
		
		t.unindent()
		t.writeIndent()
		t.write("}")

	case *ast.ReturnStmt:
		t.write("return")
		if len(n.Results) > 0 {
			t.write(" ")
			for i, res := range n.Results {
				if i > 0 { t.write(", ") }
				t.Transpile(res)
			}
		}

	case *ast.IncDecStmt:
		t.Transpile(n.X)
		t.write(n.Tok.String())

	case *ast.BinaryExpr:
		t.Transpile(n.X)
		t.write(" " + n.Op.String() + " ")
		t.Transpile(n.Y)

	case *ast.UnaryExpr:
		op := n.Op.String()
		if op == "&" {
			t.Transpile(n.X)
		} else {
			t.write(op)
			t.Transpile(n.X)
		}

	case *ast.ParenExpr:
		t.write("(")
		t.Transpile(n.X)
		t.write(")")

	case *ast.Ident:
		t.write(n.Name)

	case *ast.BasicLit:
		t.write(n.Value)

	case *ast.DeclStmt:
		t.Transpile(n.Decl)

	default:
		t.write(fmt.Sprintf("/* TODO: Implementar nó %T */", n))
	}
	return nil
}