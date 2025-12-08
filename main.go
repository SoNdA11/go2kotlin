package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"net/http"
)

// Estrutura para receber os dados do navegador (JSON)
type RequestBody struct {
	GoCode string `json:"code"`
}

// Estrutura para responder ao navegador
type ResponseBody struct {
	KotlinCode string `json:"kotlin"`
	Error      string `json:"error,omitempty"`
}

func main() {
	// 1. Rota para servir a página HTML (Frontend)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// 2. Rota da API que faz a conversão
	http.HandleFunc("/transpile", handleTranspile)

	// 3. Inicia o servidor
	fmt.Println("🚀 Servidor rodando em: http://localhost:8080")
	fmt.Println("Cole seu código Go no navegador e veja a mágica acontecer!")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleTranspile(w http.ResponseWriter, r *http.Request) {
	// Apenas aceita POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Ler o JSON enviado pelo navegador
	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao ler JSON", http.StatusBadRequest)
		return
	}

	
	// 1. Parse do código recebido (string)
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "editor.go", req.GoCode, parser.ParseComments)
	
	response := ResponseBody{}

	if err != nil {
		response.Error = fmt.Sprintf("Erro de Sintaxe Go: %v", err)
	} else {
		// 2. Chama o Transpilador
		tr := NewTranspiler()
		if err := tr.Transpile(node); err != nil {
			response.Error = fmt.Sprintf("Erro na Conversão: %v", err)
		} else {
			response.KotlinCode = tr.GetOutput()
		}
	}

	// 3. Envia a resposta JSON de volta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}