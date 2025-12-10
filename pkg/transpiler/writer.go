package transpiler

import (
	"bytes"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
)

// HandlerFunc define a assinatura da estratégia
type HandlerFunc func(t *Transpiler, node ast.Node) error

// StructDef armazena metadados sobre as structs
type StructDef struct {
	Fields map[string]bool
	Embeds []string
}

// Transpiler agora possui um mapa de estratégias (handlers)
type Transpiler struct {
	fset           *token.FileSet
	output         bytes.Buffer
	indentLevel    int
	structs        map[string]StructDef
	vars           map[string]string
	
	// Mapa de Estratégias (Tipo do Nó -> Função de Tratamento)
	handlers       map[string]HandlerFunc

	// Flags de Análise
	usesCoroutines bool
	usesChannels   bool
}

// NewTranspiler inicializa e REGISTRA as estratégias
func NewTranspiler() *Transpiler {
	t := &Transpiler{
		fset:           token.NewFileSet(),
		structs:        make(map[string]StructDef),
		vars:           make(map[string]string),
		handlers:       make(map[string]HandlerFunc),
		usesCoroutines: false,
		usesChannels:   false,
	}
	
	// Inicializa o mapa de handlers
	t.registerHandlers()
	
	return t
}

// GetOutput retorna o código gerado
func (t *Transpiler) GetOutput() string {
	return t.output.String()
}

// --- Helper Methods (Indentação e Escrita) ---

func (t *Transpiler) indent() { t.indentLevel++ }
func (t *Transpiler) unindent() { if t.indentLevel > 0 { t.indentLevel-- } }
func (t *Transpiler) write(s string) { t.output.WriteString(s) }
func (t *Transpiler) writeLine(s string) {
	t.output.WriteString(strings.Repeat("    ", t.indentLevel))
	t.output.WriteString(s)
	t.output.WriteString("\n")
}
func (t *Transpiler) writeIndent() { t.output.WriteString(strings.Repeat("    ", t.indentLevel)) }

// register associa um tipo AST a uma função
func (t *Transpiler) register(nodeType interface{}, handler HandlerFunc) {
	// Usa Reflection para pegar o nome exato do tipo (ex: "*ast.IfStmt")
	typeName := reflect.TypeOf(nodeType).String()
	t.handlers[typeName] = handler
}