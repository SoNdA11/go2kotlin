package ex04

import "fmt"

// 1. Função com parâmetros (Requisito PDF)
func Calcular(a int, b int) int {
	// 2. Expressões Aritméticas e Parênteses (Requisito PDF)
	return (a + b) * 2
}

func main() {
	// 3. Declaração e Atribuição
	var entrada string = ""
	var num1 int = 10
	var num2 int = 5
	
	// 4. Saída de Dados
	fmt.Println("Digite seu nome:")
	
	// 5. Entrada de Dados
	fmt.Scanln(&entrada)
	
	fmt.Print("Ola ")
	fmt.Println(entrada)

	// 6. Lógica E, OU, NÃO
	ativo := true
	bloqueado := false

	if ativo && !bloqueado {
		fmt.Println("Sistema Online")
		
		resultado := Calcular(num1, num2)
		fmt.Println("Resultado do calculo (10+5)*2:")
		fmt.Println(resultado)
	}

	// 8. Repetição
	fmt.Println("Contagem Regressiva:")
	i := 5
	for i > 0 {
		fmt.Println(i)
		i--
	}
}