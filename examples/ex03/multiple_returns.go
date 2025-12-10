package main

import "fmt"

func divideAndRemainder(a, b int) (quotient, remainder int) {
    quotient = a / b
    remainder = a % b
    return
}

func main() {
    q, r := divideAndRemainder(11, 3)
    fmt.Println("q:", q, "r:", r)
}
