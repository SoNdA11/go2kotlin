package main

import "fmt"

// Teste de função pública (Capitalizada)
func Multiplicar(a int, b int) int {
	return a * b
}

// Teste de função interna (Minúscula)
func log(msg string) {
	fmt.Println(msg)
}

func main() {
	x := 10
	y := 5
	
	resultado := Multiplicar(x, y)
	log("O resultado é:")
	fmt.Println(resultado)
}