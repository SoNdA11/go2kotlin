package main

import (
	"go/ast"
	"strings"
)

var typeMapping = map[string]string{
	"int":     "Int",
	"int8":    "Byte",
	"int16":   "Short",
	"int32":   "Int",
	"int64":   "Long",
	"uint":    "UInt",
	"uint8":   "UByte",
	"uint16":  "UShort",
	"uint32":  "UInt",
	"uint64":  "ULong",
	"float32": "Float",
	"float64": "Double",
	"complex64":  "Any /* Complex */",
	"complex128": "Any /* Complex */",
	"byte":    "UByte",
	"rune":    "Char",
	"bool":    "Boolean",
	"string":  "String",
	"uintptr": "Long",
	"any":     "Any",
}

func (t *Transpiler) resolveType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		if val, ok := typeMapping[e.Name]; ok {
			return val
		}
		return e.Name

	case *ast.ArrayType:
		inner := t.resolveType(e.Elt)
		return "MutableList<" + inner + ">"

	case *ast.MapType:
		key := t.resolveType(e.Key)
		val := t.resolveType(e.Value)
		return "MutableMap<" + key + ", " + val + ">"

	case *ast.StarExpr:
		return t.resolveType(e.X) + "?"

	case *ast.SelectorExpr:
		return t.resolveType(e.X) + "." + e.Sel.Name

	case *ast.InterfaceType:
		if e.Methods == nil || len(e.Methods.List) == 0 {
			return "Any"
		}
		return "Any"

	case *ast.FuncType:
		var params []string
		if e.Params != nil {
			for _, field := range e.Params.List {
				pType := t.resolveType(field.Type)
				count := len(field.Names)
				if count == 0 { count = 1 }
				for k := 0; k < count; k++ {
					params = append(params, pType)
				}
			}
		}
		
		ret := "Unit"
		if e.Results != nil && len(e.Results.List) > 0 {
			ret = t.resolveType(e.Results.List[0].Type)
		}
		
		return "(" + strings.Join(params, ", ") + ") -> " + ret

	default:
		return "Any"
	}
}