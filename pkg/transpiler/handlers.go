package transpiler

import (
	"go/ast"
	"go/token"
	"strings"
)

// registerHandlers mapeia todos os nós suportados
func (t *Transpiler) registerHandlers() {
	// Arquivo e Declarações
	t.register(&ast.File{}, t.handleFile)
	t.register(&ast.GenDecl{}, t.handleGenDecl)
	t.register(&ast.FuncDecl{}, t.handleFuncDecl)
	
	// Statements (Comandos)
	t.register(&ast.BlockStmt{}, t.handleBlockStmt)
	t.register(&ast.AssignStmt{}, t.handleAssignStmt)
	t.register(&ast.ExprStmt{}, t.handleExprStmt)
	t.register(&ast.ReturnStmt{}, t.handleReturnStmt)
	t.register(&ast.IfStmt{}, t.handleIfStmt)
	t.register(&ast.ForStmt{}, t.handleForStmt)
	t.register(&ast.RangeStmt{}, t.handleRangeStmt)
	t.register(&ast.BranchStmt{}, t.handleBranchStmt)
	t.register(&ast.SwitchStmt{}, t.handleSwitchStmt)
	t.register(&ast.CaseClause{}, t.handleCaseClause)
	t.register(&ast.IncDecStmt{}, t.handleIncDecStmt)
	t.register(&ast.GoStmt{}, t.handleGoStmt)
	t.register(&ast.SendStmt{}, t.handleSendStmt)
	t.register(&ast.DeclStmt{}, t.handleDeclStmt)

	// Expressions (Expressões)
	t.register(&ast.CallExpr{}, t.handleCallExpr)
	t.register(&ast.BinaryExpr{}, t.handleBinaryExpr)
	t.register(&ast.UnaryExpr{}, t.handleUnaryExpr)
	t.register(&ast.ParenExpr{}, t.handleParenExpr)
	t.register(&ast.IndexExpr{}, t.handleIndexExpr)
	t.register(&ast.StarExpr{}, t.handleStarExpr)
	t.register(&ast.KeyValueExpr{}, t.handleKeyValueExpr)
	t.register(&ast.SelectorExpr{}, t.handleSelectorExpr)
	t.register(&ast.CompositeLit{}, t.handleCompositeLit)
	t.register(&ast.Ident{}, t.handleIdent)
	t.register(&ast.BasicLit{}, t.handleBasicLit)
	t.register(&ast.FuncLit{}, t.handleFuncLit)
	t.register(&ast.TypeAssertExpr{}, t.handleTypeAssertExpr)
}

// --- IMPLEMENTAÇÃO DAS ESTRATÉGIAS ---

func (t *Transpiler) handleFile(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.File)
	
	// Pre-analysis pass
	t.analyzeFeatures(n)

	t.writeLine("package " + n.Name.Name)
	t.write("\n")

	if t.usesCoroutines {
		t.writeLine("import kotlinx.coroutines.*")
	}
	if t.usesChannels {
		t.writeLine("import kotlinx.coroutines.channels.Channel")
	}

	if len(n.Imports) > 0 {
		for _, imp := range n.Imports {
			path := strings.Trim(imp.Path.Value, "\"")
			if path != "fmt" && path != "time" {
				t.writeLine("import " + path)
			}
		}
		t.write("\n")
	}

	for _, decl := range n.Decls {
		t.Transpile(decl)
		t.write("\n\n")
	}
	return nil
}

func (t *Transpiler) handleGenDecl(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.GenDecl)
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
	return nil
}

func (t *Transpiler) handleFuncDecl(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.FuncDecl)
	t.writeIndent()
	if ast.IsExported(n.Name.Name) { t.write("public ") } else { t.write("internal ") }
	t.write("fun ")

	if n.Type.TypeParams != nil {
		t.write("<")
		for i, field := range n.Type.TypeParams.List {
			if i > 0 { t.write(", ") }
			for j, name := range field.Names {
				if j > 0 { t.write(", ") }
				t.write(name.Name)
			}
		}
		t.write("> ")
	}

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

	if n.Name.Name == "main" {
		if t.usesCoroutines {
			t.write(" = runBlocking ")
		} else {
			t.write(" ")
		}
	} else {
		t.write(" ")
	}

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
	return nil
}

func (t *Transpiler) handleBlockStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.BlockStmt)
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
	return nil
}

func (t *Transpiler) handleAssignStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.AssignStmt)
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
	return nil
}

func (t *Transpiler) handleExprStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.ExprStmt)
	t.Transpile(n.X)
	return nil
}

func (t *Transpiler) handleCallExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.CallExpr)
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
				if ch, ok := n.Args[0].(*ast.ChanType); ok {
					innerType := t.resolveType(ch.Value)
					t.write("Channel<" + innerType + ">()")
					return nil
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
		if x, ok := sel.X.(*ast.Ident); ok {
			if x.Name == "time" && sel.Sel.Name == "Sleep" {
				t.write("delay")
				isHandled = true
			}
			if x.Name == "fmt" {
				if sel.Sel.Name == "Printf" {
					t.write("System.out.printf(")
					for i, arg := range n.Args {
						if i > 0 { t.write(", ") }
						t.Transpile(arg)
					}
					t.write(")")
					return nil
				}

				if strings.HasPrefix(sel.Sel.Name, "Print") {
					cmd := "print"
					if sel.Sel.Name == "Println" { cmd = "println" }
					if len(n.Args) > 1 {
						t.write(cmd + "(\"")
						for i, arg := range n.Args {
							if i > 0 { t.write(" ") }
							t.write("${")
							t.Transpile(arg)
							t.write("}")
						}
						t.write("\")")
						return nil
					} else {
						t.write(cmd)
						isHandled = true
					}
				}

				if strings.HasPrefix(sel.Sel.Name, "Scan") {
					if len(n.Args) > 0 {
						t.Transpile(n.Args[0])
						t.write(" = readln()")
						t.write(" // !! Converter tipo se necessario")
						return nil
					}
				}
			}
		}
	}
	if !isHandled {
		if _, isFuncLit := n.Fun.(*ast.FuncLit); isFuncLit {
			t.write("(")
			t.Transpile(n.Fun)
			t.write(")")
		} else {
			t.Transpile(n.Fun)
		}
	}

	t.write("(")
	for i, arg := range n.Args {
		if i > 0 { t.write(", ") }
		t.Transpile(arg)
	}
	t.write(")")
	return nil
}

func (t *Transpiler) handleIfStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.IfStmt)
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
	return nil
}

func (t *Transpiler) handleForStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.ForStmt)
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
	return nil
}

func (t *Transpiler) handleReturnStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.ReturnStmt)
	t.write("return")
	if len(n.Results) > 0 {
		t.write(" ")
		for i, res := range n.Results {
			if i > 0 { t.write(", ") }
			t.Transpile(res)
		}
	}
	return nil
}

func (t *Transpiler) handleIncDecStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.IncDecStmt)
	t.Transpile(n.X)
	t.write(n.Tok.String())
	return nil
}

func (t *Transpiler) handleBinaryExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.BinaryExpr)
	if t.hasComplex(n) {
		t.write("/* Complex Logic */ null")
	} else {
		t.Transpile(n.X)
		t.write(" " + n.Op.String() + " ")
		t.Transpile(n.Y)
	}
	return nil
}

func (t *Transpiler) handleParenExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.ParenExpr)
	t.write("(")
	t.Transpile(n.X)
	t.write(")")
	return nil
}

func (t *Transpiler) handleIdent(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.Ident)
	t.write(n.Name)
	return nil
}

func (t *Transpiler) handleBasicLit(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.BasicLit)
	t.write(n.Value)
	return nil
}

func (t *Transpiler) handleFuncLit(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.FuncLit)
	t.write("fun(")
	t.writeParams(n.Type.Params)
	t.write(")")
	if n.Type.Results != nil && len(n.Type.Results.List) > 0 {
		retType := t.resolveType(n.Type.Results.List[0].Type)
		t.write(": " + retType)
	}
	t.write(" ")
	t.Transpile(n.Body)
	return nil
}

func (t *Transpiler) handleSendStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.SendStmt)
	t.Transpile(n.Chan)
	t.write(".send(")
	t.Transpile(n.Value)
	t.write(")")
	return nil
}

func (t *Transpiler) handleUnaryExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.UnaryExpr)
	switch n.Op.String() {
	case "<-":
		t.Transpile(n.X)
		t.write(".receive()")
	case "&":
		t.Transpile(n.X)
	default:
		t.write(n.Op.String())
		t.Transpile(n.X)
	}
	return nil
}

func (t *Transpiler) handleGoStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.GoStmt)
	t.write("launch {\n")
	t.indent()
	if call, ok := n.Call.Fun.(*ast.FuncLit); ok {
		for _, stmt := range call.Body.List {
			t.writeIndent()
			t.Transpile(stmt)
			t.write("\n")
		}
	} else {
		t.writeIndent()
		t.Transpile(n.Call)
		t.write("\n")
	}
	t.unindent()
	t.writeIndent()
	t.write("}")
	return nil
}

func (t *Transpiler) handleTypeAssertExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.TypeAssertExpr)
	t.Transpile(n.X)
	t.write(" as ")
	t.write(t.resolveType(n.Type))
	return nil
}

func (t *Transpiler) handleRangeStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.RangeStmt)
	t.write("for (")
	key := "_"
	if n.Key != nil {
		if id, ok := n.Key.(*ast.Ident); ok {
			key = id.Name
		}
	}
	val := ""
	if n.Value != nil {
		if id, ok := n.Value.(*ast.Ident); ok {
			val = id.Name
		}
	}
	if key != "_" && val != "" {
		t.write("(" + key + ", " + val + ") in ")
		t.Transpile(n.X)
		t.write(".withIndex()")
	} else if key == "_" && val != "" {
		t.write(val + " in ")
		t.Transpile(n.X)
	} else {
		t.write(key + " in ")
		t.Transpile(n.X)
		t.write(".indices")
	}
	t.write(") ")
	t.Transpile(n.Body)
	return nil
}

func (t *Transpiler) handleBranchStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.BranchStmt)
	t.write(n.Tok.String())
	if n.Label != nil {
		t.write("@" + n.Label.Name)
	}
	return nil
}

func (t *Transpiler) handleSwitchStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.SwitchStmt)
	if n.Init != nil {
		t.write("run {\n")
		t.indent()
		t.writeIndent()
		t.Transpile(n.Init)
		t.write("\n")
		t.writeIndent()
	}
	t.write("when")
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
	return nil
}

func (t *Transpiler) handleCaseClause(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.CaseClause)
	if n.List == nil {
		t.write("else -> ")
	} else {
		for i, expr := range n.List {
			if i > 0 { t.write(", ") }
			t.Transpile(expr)
		}
		t.write(" -> ")
	}
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
	return nil
}

func (t *Transpiler) handleIndexExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.IndexExpr)
	t.Transpile(n.X)
	t.write("[")
	t.Transpile(n.Index)
	t.write("]")
	return nil
}

func (t *Transpiler) handleStarExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.StarExpr)
	t.Transpile(n.X)
	return nil
}

func (t *Transpiler) handleCompositeLit(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.CompositeLit)
	switch n.Type.(type) {
	case *ast.ArrayType:
		t.write("mutableListOf")
	case *ast.MapType:
		t.write("mutableMapOf")
	case nil:
	default:
		t.Transpile(n.Type)
	}
	t.write("(")
	for i, elt := range n.Elts {
		if i > 0 { t.write(", ") }
		t.Transpile(elt)
	}
	t.write(")")
	return nil
}

func (t *Transpiler) handleKeyValueExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.KeyValueExpr)
	t.Transpile(n.Key)
	t.write(" to ")
	t.Transpile(n.Value)
	return nil
}

func (t *Transpiler) handleSelectorExpr(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.SelectorExpr)
	if x, ok := n.X.(*ast.Ident); ok && x.Name == "time" {
		switch n.Sel.Name {
		case "Second":
			t.write("1000L")
			return nil
		case "Millisecond":
			t.write("1L")
			return nil
		}
	}
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
	return nil
}

func (t *Transpiler) handleDeclStmt(tr *Transpiler, node ast.Node) error {
	n := node.(*ast.DeclStmt)
	t.Transpile(n.Decl)
	return nil
}

// --- Métodos Auxiliares Necessários nos Handlers ---

func (t *Transpiler) analyzeFeatures(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GoStmt:
			t.usesCoroutines = true
		case *ast.ChanType:
			t.usesChannels = true
			t.usesCoroutines = true
		case *ast.SendStmt:
			t.usesChannels = true
			t.usesCoroutines = true
		case *ast.UnaryExpr:
			if x.Op == token.ARROW {
				t.usesChannels = true
				t.usesCoroutines = true
			}
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == "time" && sel.Sel.Name == "Sleep" {
					t.usesCoroutines = true
				}
			}
			if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "make" && len(x.Args) > 0 {
				if _, ok := x.Args[0].(*ast.ChanType); ok {
					t.usesChannels = true
					t.usesCoroutines = true
				}
			}
		}
		return true
	})
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