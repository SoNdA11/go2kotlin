package handler

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"net/http"
	"go2kotlin/pkg/transpiler"
)

// Estruturas de Dados (DTOs)
type RequestBody struct {
	GoCode string `json:"code"`
}

type ResponseBody struct {
	KotlinCode string `json:"kotlin"`
	Error      string `json:"error,omitempty"`
}

// Handler é a função exportada que a Vercel executa
func Handler(w http.ResponseWriter, r *http.Request) {
	// Configuração de CORS para permitir requisições do frontend
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Responde a requisições pre-flight (OPTIONS)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Valida se é POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodifica o JSON
	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao ler JSON", http.StatusBadRequest)
		return
	}

	// Executa a lógica do transpilador
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "editor.go", req.GoCode, parser.ParseComments)
	
	response := ResponseBody{}

	if err != nil {
		response.Error = fmt.Sprintf("Erro de Sintaxe Go: %v", err)
	} else {
		// Chama o Transpilador do seu pacote pkg
		tr := transpiler.NewTranspiler()
		if err := tr.Transpile(node); err != nil {
			response.Error = fmt.Sprintf("Erro na Conversão: %v", err)
		} else {
			response.KotlinCode = tr.GetOutput()
		}
	}

	// Retorna a resposta JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}