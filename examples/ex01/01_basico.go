package main

import "fmt"

func main() {
	// Teste de variáveis e tipos primitivos
	var idade int = 30
	var nome string = "Go Transpiler"
	ativo := true

	if ativo {
		fmt.Println(nome)
		fmt.Println(idade)
	}
}