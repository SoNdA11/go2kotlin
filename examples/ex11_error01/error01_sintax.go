package main

import "fmt"

// --- EXEMPLO DE ERRO DE SINTAXE ---
// Este código falhará na etapa de Parsing (Análise Sintática).
// O transpilador nem chegará a tentar converter para Kotlin.

func main() {
	// ERRO 1: Go não permite ponto e vírgula no início de bloco assim
	// ERRO 2: A palavra 'var' está escrita errada como 'variable'
	variable x int = 10;

	// ERRO 3: if sem chaves (Go obriga o uso de chaves {})
	if x > 5 
		fmt.Println("Maior que 5")
	
	// ERRO 4: Tentativa de usar sintaxe de outras linguagens
	// Go não tem 'while', apenas 'for'
	while (x > 0) {
		x--
	}
}