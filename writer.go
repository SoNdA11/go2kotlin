package main

import (
	"bytes"
	"go/token"
	"strings"
)

// Transpiler mantém o estado da geração de código.
// Ele encapsula o buffer de saída e o nível de indentação atual.
type Transpiler struct {
	fset        *token.FileSet // Usado para rastrear posições no arquivo original
	output      bytes.Buffer   // Onde o código Kotlin é acumulado
	indentLevel int            // Nível atual de indentação (tabs/espaços)
}

// NewTranspiler cria uma nova instância pronta para uso
func NewTranspiler() *Transpiler {
	return &Transpiler{
		fset: token.NewFileSet(),
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

// Escreve string crua no buffer
func (t *Transpiler) write(s string) {
	t.output.WriteString(s)
}

// Escreve uma linha completa com a indentação correta e quebra de linha
func (t *Transpiler) writeLine(s string) {
	t.output.WriteString(strings.Repeat("    ", t.indentLevel))
	t.output.WriteString(s)
	t.output.WriteString("\n")
}

// Apenas insere os espaços de indentação (útil para inícios de blocos)
func (t *Transpiler) writeIndent() {
	t.output.WriteString(strings.Repeat("    ", t.indentLevel))
}