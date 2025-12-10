package main

import (
    "errors"
    "fmt"
)

func mightFail(shouldFail bool) (string, error) {
    if shouldFail {
        return "", errors.New("algo deu errado")
    }
    return "sucesso", nil
}

func main() {
    if msg, err := mightFail(true); err != nil {
        fmt.Println("Erro:", err)
    } else {
        fmt.Println("OK:", msg)
    }
}
