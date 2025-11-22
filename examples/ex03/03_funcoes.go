package main

import "fmt"

// Teste de função pública (Capitalizada -> Public)
func Multiplicar(a int, b int) int {
	return a * b
}

// Teste de função interna (Minúscula -> Internal)
func log(msg string) {
	fmt.Println(msg)
}

func main() {
	x := 10
	y := 5
	
	// Chamando função que retorna valor
	resultado := Multiplicar(x, y)
	
	// Chamando função void (sem retorno)
	log("O resultado da multiplicação é:")
	fmt.Println(resultado)
}