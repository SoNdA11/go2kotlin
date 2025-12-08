package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func (t *Transpiler) Transpile(node ast.Node) error {
	switch n := node.(type) {
	
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
		
		for _, decl := range n.Decls {
			t.Transpile(decl)
			t.write("\n\n") 
		}

	case *ast.GenDecl:
		// STRUCTS
		if n.Tok == token.TYPE {
			for _, spec := range n.Specs {
				ts := spec.(*ast.TypeSpec)
				if st, ok := ts.Type.(*ast.StructType); ok {
					def := StructDef{Fields: make(map[string]bool), Embeds: []string{}}
					t.write("data class " + ts.Name.Name + "(")
					if st.Fields != nil {
						for i, field := range st.Fields.List {
							if i > 0 { t.write(", ") }
							typeStr := t.resolveType(field.Type)
							if len(field.Names) == 0 {
								fieldName := typeStr
								if idx := strings.LastIndex(fieldName, "."); idx != -1 {
									fieldName = fieldName[idx+1:]
								}
								t.write("var " + fieldName + ": " + typeStr)
								def.Embeds = append(def.Embeds, fieldName)
							} else {
								for j, name := range field.Names {
									if j > 0 { t.write(", ") }
									t.write("var " + name.Name + ": " + typeStr)
									def.Fields[name.Name] = true
								}
							}
						}
					}
					t.write(")")
					t.structs[ts.Name.Name] = def
				} else {
					t.write("typealias " + ts.Name.Name + " = " + t.resolveType(ts.Type))
				}
			}
			return nil
		}
		
		// VARIAVEIS E CONSTANTES
		if n.Tok == token.VAR || n.Tok == token.CONST {
			keyword := "var"
			if n.Tok == token.CONST { keyword = "val" }
			for _, spec := range n.Specs {
				vspec := spec.(*ast.ValueSpec)
				typeName := ""
				
				if vspec.Type != nil {
					typeName = t.resolveType(vspec.Type) 
				}

				for i, name := range vspec.Names {
					if len(vspec.Values) > i {
						if comp, ok := vspec.Values[i].(*ast.CompositeLit); ok {
							if ident, ok := comp.Type.(*ast.Ident); ok {
								t.vars[name.Name] = ident.Name
							}
						}
					}
					t.write(keyword + " " + name.Name)
					
					if typeName != "" {
						t.write(": " + typeName)
					}

					if i < len(vspec.Values) {
						t.write(" = ")
						t.transpileTypedValue(vspec.Values[i], typeName)
					}
				}
			}
		}

	case *ast.FuncDecl:
		t.writeIndent()
		if ast.IsExported(n.Name.Name) { t.write("public ") } else { t.write("internal ") }
		t.write("fun ")

		recvParamName := ""
		recvTypeName := ""
		if n.Recv != nil && len(n.Recv.List) > 0 {
			recvTypeName = t.resolveType(n.Recv.List[0].Type)
			t.write(recvTypeName + ".")
			if len(n.Recv.List[0].Names) > 0 {
				recvParamName = n.Recv.List[0].Names[0].Name
				t.vars[recvParamName] = recvTypeName
			}
		}

		t.write(n.Name.Name + "(")
		t.writeParams(n.Type.Params)
		t.write(")")

		if n.Type.Results != nil && len(n.Type.Results.List) > 0 {
			retType := t.resolveType(n.Type.Results.List[0].Type)
			t.write(": " + retType)
		}
		t.write(" ")
		
		if recvParamName != "" && recvParamName != "_" {
			t.write("{\n")
			t.indent()
			t.writeIndent()
			t.write("val " + recvParamName + " = this\n")
			for _, stmt := range n.Body.List {
				t.writeIndent()
				t.Transpile(stmt)
				t.write("\n")
			}
			t.unindent()
			t.writeIndent()
			t.write("}")
		} else {
			t.Transpile(n.Body)
		}

	case *ast.FuncLit:
		t.write("fun(")
		t.writeParams(n.Type.Params)
		t.write(")")
		if n.Type.Results != nil && len(n.Type.Results.List) > 0 {
			retType := t.resolveType(n.Type.Results.List[0].Type)
			t.write(": " + retType)
		}
		t.write(" ")
		t.Transpile(n.Body)

	// --- SWITCH / WHEN ---
	case *ast.SwitchStmt:
		// Se tiver inicialização (switch x := 10; x {...}), isola o escopo
		if n.Init != nil {
			t.write("run {\n")
			t.indent()
			t.writeIndent()
			t.Transpile(n.Init)
			t.write("\n")
			t.writeIndent()
		}

		t.write("when")
		// Se tiver tag (switch x {...}), gera 'when(x)'. Se não (switch {...}), gera 'when'
		if n.Tag != nil {
			t.write(" (")
			t.Transpile(n.Tag)
			t.write(")")
		}
		t.write(" ")
		
		t.Transpile(n.Body)

		if n.Init != nil {
			t.unindent()
			t.write("\n")
			t.writeIndent()
			t.write("}")
		}

	// --- CASE ---
	case *ast.CaseClause:
		if n.List == nil {
			// default:
			t.write("else -> ")
		} else {
			// case "A", "B":
			for i, expr := range n.List {
				if i > 0 { t.write(", ") }
				t.Transpile(expr)
			}
			t.write(" -> ")
		}
		
		// Gera bloco para o corpo do case
		t.write("{\n")
		t.indent()
		for _, stmt := range n.Body {
			t.writeIndent()
			t.Transpile(stmt)
			t.write("\n")
		}
		t.unindent()
		t.writeIndent()
		t.write("}")

	case *ast.IndexExpr:
		t.Transpile(n.X)
		t.write("[")
		t.Transpile(n.Index)
		t.write("]")

	case *ast.StarExpr:
		t.Transpile(n.X)

	case *ast.CompositeLit:
		switch n.Type.(type) {
		case *ast.ArrayType:
			t.write("mutableListOf")
		case *ast.MapType:
			t.write("mutableMapOf")
		case nil:
			// Tipo implícito
		default:
			t.Transpile(n.Type)
		}
		
		t.write("(")
		for i, elt := range n.Elts {
			if i > 0 { t.write(", ") }
			t.Transpile(elt)
		}
		t.write(")")
	
	case *ast.KeyValueExpr:
		t.Transpile(n.Key)
		t.write(" to ")
		t.Transpile(n.Value)

	case *ast.SelectorExpr:
		varVarName := ""
		if ident, ok := n.X.(*ast.Ident); ok {
			varVarName = ident.Name
		}
		t.Transpile(n.X)
		injectedEmbed := ""
		if varType, ok := t.vars[varVarName]; ok {
			if structInfo, ok := t.structs[varType]; ok {
				fieldName := n.Sel.Name
				if !structInfo.Fields[fieldName] {
					for _, embedName := range structInfo.Embeds {
						injectedEmbed = "." + embedName
						break 
					}
				}
			}
		}
		t.write(injectedEmbed) 
		t.write(".")
		t.write(n.Sel.Name)

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

	case *ast.AssignStmt:
		for i, lhs := range n.Lhs {
			if n.Tok == token.DEFINE {
				t.write("var ") 
				if ident, ok := lhs.(*ast.Ident); ok {
					if len(n.Rhs) > i {
						if comp, ok := n.Rhs[i].(*ast.CompositeLit); ok {
							if typeIdent, ok := comp.Type.(*ast.Ident); ok {
								t.vars[ident.Name] = typeIdent.Name
							}
						}
						if rhsIdent, ok := n.Rhs[i].(*ast.Ident); ok {
							if existingType, exists := t.vars[rhsIdent.Name]; exists {
								t.vars[ident.Name] = existingType
							}
						}
					}
				}
			}
			t.Transpile(lhs)
			if i < len(n.Rhs) {
				t.write(" = ")
				t.Transpile(n.Rhs[i])
			}
		}

	case *ast.ExprStmt:
		t.Transpile(n.X)

	case *ast.CallExpr:
		if ident, ok := n.Fun.(*ast.Ident); ok {
			if ident.Name == "make" {
				if len(n.Args) > 0 {
					if _, ok := n.Args[0].(*ast.MapType); ok {
						ktType := t.resolveType(n.Args[0])
						if strings.HasPrefix(ktType, "MutableMap") {
							genericPart := strings.TrimPrefix(ktType, "MutableMap")
							t.write("mutableMapOf" + genericPart + "()")
							return nil
						}
					}
				}
			}
			if ident.Name == "append" && len(n.Args) >= 2 {
				t.write("(")
				t.Transpile(n.Args[0])
				t.write(" + ")
				t.Transpile(n.Args[1])
				t.write(")")
				return nil
			}
		}

		isHandled := false
		if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "fmt" {
				// Printf
				if sel.Sel.Name == "Printf" {
					t.write("System.out.printf(") 
					for i, arg := range n.Args {
						if i > 0 { t.write(", ") }
						t.Transpile(arg)
					}
					t.write(")")
					isHandled = true
					return nil 
				}

				// Print / Println
				if strings.HasPrefix(sel.Sel.Name, "Print") {
					cmd := "print"
					if sel.Sel.Name == "Println" { cmd = "println" }
					if len(n.Args) > 1 {
						t.write(cmd + "(\"\" + ") 
						for i, arg := range n.Args {
							if i > 0 { t.write(" + \" \" + ") }
							t.Transpile(arg)
						}
						t.write(")")
						isHandled = true
						return nil 
					} else {
						t.write(cmd)
						isHandled = true
					}
				} 
				
				// Scan
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
		if t.hasComplex(n) {
			t.write("/* Complex Logic */ null")
		} else {
			t.Transpile(n.X)
			t.write(" " + n.Op.String() + " ")
			t.Transpile(n.Y)
		}

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

func (t *Transpiler) writeParams(fields *ast.FieldList) {
	if fields != nil {
		for i, field := range fields.List {
			if i > 0 { t.write(", ") }
			typeName := t.resolveType(field.Type)
			for j, name := range field.Names {
				if j > 0 { t.write(", ") }
				t.write(name.Name + ": " + typeName)
			}
		}
	}
}

func (t *Transpiler) hasComplex(node ast.Node) bool {
	has := false
	ast.Inspect(node, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.IMAG {
			has = true
			return false
		}
		return true
	})
	return has
}

func (t *Transpiler) transpileTypedValue(expr ast.Expr, targetType string) {
	if lit, ok := expr.(*ast.BasicLit); ok {
		val := lit.Value
		if lit.Kind == token.IMAG {
			t.write("/* Complex not supported */ null")
			return
		}
		if targetType == "Float" && !strings.HasSuffix(val, "f") {
			t.write(val + "f")
			return
		}
		if strings.HasPrefix(targetType, "U") { 
			if strings.HasPrefix(val, "'") {
				t.write(val + ".code.toUByte()")
				return
			}
			if !strings.HasSuffix(val, "u") {
				if targetType == "ULong" {
					t.write(val + "uL")
				} else {
					t.write(val + "u")
				}
				return
			}
		}
		if targetType == "Long" && !strings.HasSuffix(val, "L") {
			t.write(val + "L")
			return
		}
		t.write(val)
		return
	}
	if t.hasComplex(expr) {
		t.write("/* Complex Expression */ null")
	} else {
		t.Transpile(expr)
	}
}