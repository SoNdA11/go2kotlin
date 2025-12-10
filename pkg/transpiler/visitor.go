package transpiler

import (
	"fmt"
	"go/ast"
	"reflect"
)

func (t *Transpiler) Transpile(node ast.Node) error {
	if node == nil {
		return nil
	}

	// 1. Descobre o tipo do nó (ex: "*ast.IfStmt")
	nodeType := reflect.TypeOf(node).String()

	// 2. Busca a estratégia no mapa
	handler, found := t.handlers[nodeType]

	if found {
		// 3. Executa a estratégia
		return handler(t, node)
	}

	// Fallback para nós não implementados
	t.write(fmt.Sprintf("/* TODO: Implementar nó %s */", nodeType))
	return nil
}