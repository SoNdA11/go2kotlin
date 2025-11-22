package main

import (
	"go/ast"
)

// Tabela 1 do PDF: Mapeamento direto de tipos primitivos
var typeMapping = map[string]string{
	"int":     "Int",
	"int8":    "Byte",
	"int16":   "Short",
	"int32":   "Int",
	"int64":   "Long",
	"uint":    "UInt", // Nota do PDF: uint é problemático, mas mapeamos direto por enquanto
	"float32": "Float",
	"float64": "Double",
	"bool":    "Boolean",
	"string":  "String",
}

// resolveType implementa a lógica de conversão de tipos complexos
// Baseado na seção "Tipos (Types)" do documento
func (t *Transpiler) resolveType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		// Verifica se é um tipo primitivo mapeado
		if val, ok := typeMapping[e.Name]; ok {
			return val
		}
		// Se não, assume que é uma Struct ou Interface customizada
		return e.Name 

	case *ast.ArrayType:
		// PDF Tabela 1: Slices ([]T) -> MutableList<T>
		// Arrays ([N]T) -> Array<T> (simplificado aqui para MutableList para generalizar)
		inner := t.resolveType(e.Elt)
		return "MutableList<" + inner + ">"

	case *ast.StarExpr:
		// PDF Seção "Tipos Especiais": *T -> T? (Nullable)
		// Decisão de design: Ponteiros viram nulos para evitar NullPointerExceptions diretas
		return t.resolveType(e.X) + "?"

	case *ast.SelectorExpr:
		// Caso para tipos importados (ex: time.Time)
		return t.resolveType(e.X) + "." + e.Sel.Name

	default:
		return "Any" // Fallback para interface{} ou tipos desconhecidos
	}
}