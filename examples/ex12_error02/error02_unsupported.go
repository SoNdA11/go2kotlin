package main

import "fmt"

// --- EXEMPLO DE LIMITAÇÃO DE SUPORTE ---
// Este código é Go válido e compila normalmente com 'go build'.
// Porém, nosso transpilador (v1.0) pode não saber converter tudo.

func main() {
	// 1. DEFER: O transpilador atual ignora ou não converte 'defer' corretamente
	// Em Kotlin, isso seria um 'try/finally', mas é complexo de mapear.
	defer fmt.Println("Isso roda no final")

	fmt.Println("Isso roda primeiro")

	// 2. INTERFACE VAZIA E TYPE SWITCH
	// Recursos avançados de tipagem dinâmica do Go.
	var x interface{} = "teste"

	switch v := x.(type) {
	case int:
		fmt.Println("É inteiro:", v)
	case string:
		fmt.Println("É string:", v)
	default:
		fmt.Println("Não sei o tipo")
	}

	// 3. PANIC
	// Go usa panic/recover. Kotlin usa throw/try-catch.
	// O transpilador pode apenas copiar a chamada de função ou gerar erro.
	panic("Erro fatal simulado")
}

/* RESULTADO ESPERADO NO KOTLIN (Parcial):
O código pode ser gerado com comentários como:
"/* TODO: Implementar nó *ast.DeferStmt e gerar código Kotlin que não compila (ex: chamando panic() que não existe igual). 
*/