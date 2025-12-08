package main

import (
	"bytes"
	"go/token"
	"strings"
)

// StructDef armazena metadados sobre as structs encontradas
type StructDef struct {
	Fields map[string]bool
	Embeds []string
}

// Transpiler mantém o estado da geração de código.
type Transpiler struct {
	fset        *token.FileSet
	output      bytes.Buffer
	indentLevel int
	structs map[string]StructDef 
	vars    map[string]string    
}

// NewTranspiler inicializa os mapas
func NewTranspiler() *Transpiler {
	return &Transpiler{
		fset:    token.NewFileSet(),
		structs: make(map[string]StructDef),
		vars:    make(map[string]string),
	}
}

// GetOutput retorna o código Kotlin final gerado
func (t *Transpiler) GetOutput() string {
	return t.output.String()
}

// --- Métodos Auxiliares de Escrita ---

func (t *Transpiler) indent() {
	t.indentLevel++
}

func (t *Transpiler) unindent() {
	if t.indentLevel > 0 {
		t.indentLevel--
	}
}

func (t *Transpiler) write(s string) {
	t.output.WriteString(s)
}

func (t *Transpiler) writeLine(s string) {
	t.output.WriteString(strings.Repeat("    ", t.indentLevel))
	t.output.WriteString(s)
	t.output.WriteString("\n")
}

func (t *Transpiler) writeIndent() {
	t.output.WriteString(strings.Repeat("    ", t.indentLevel))
}