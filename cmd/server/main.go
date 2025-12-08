package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"net/http"
	"go2kotlin/pkg/transpiler"
)

type RequestBody struct {
	GoCode string `json:"code"`
}

type ResponseBody struct {
	KotlinCode string `json:"kotlin"`
	Error      string `json:"error,omitempty"`
}

func main() {
	// 1. Rota para servir a página HTML (Frontend)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	// 2. Rota da API que faz a conversão
	http.HandleFunc("/transpile", handleTranspile)

	// 3. Inicia o servidor
	fmt.Println("🚀 Servidor rodando em: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleTranspile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao ler JSON", http.StatusBadRequest)
		return
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "editor.go", req.GoCode, parser.ParseComments)
	
	response := ResponseBody{}

	if err != nil {
		response.Error = fmt.Sprintf("Erro de Sintaxe Go: %v", err)
	} else {
		tr := transpiler.NewTranspiler()
		if err := tr.Transpile(node); err != nil {
			response.Error = fmt.Sprintf("Erro na Conversão: %v", err)
		} else {
			response.KotlinCode = tr.GetOutput()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}