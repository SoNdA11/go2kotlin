<p align="center"> <img src="print-go2kt.png" alt="Go2Kotlin Demo" width="650"> </p> <p align="center"> <img src="https://img.shields.io/badge/Language-Go-blue?logo=go" /> <img src="https://img.shields.io/badge/Target-Kotlin-purple?logo=kotlin" /> <img src="https://img.shields.io/badge/Status-In%20Development-yellow" /> <img src="https://img.shields.io/badge/Build-Source%20to%20Source-4caf50" /></p>

# Projeto Acadêmico

Este projeto faz parte da disciplina Compiladores e Paradigmas de Programação, ministrada pelo professor Sebastião Filho, na Universidade do Estado do Rio Grande do Norte (UERN).

# Grupo de Desenvolvimento

- Paulo Sérgio

- Eduardo Marinho

- Marlos Emanuel

- Vinicius Eduardo

- Luiz Henrique

# Sobre o Projeto

O Go2Kotlin Transpiler é um transpilador source-to-source que converte código Golang → Kotlin, mantendo a lógica, estrutura do código original.

## Arquitetura do Projeto

Este projeto não cria binários executáveis; ele traduz a **Árvore Sintática Abstrata (AST)** de uma linguagem para outra.
Abaixo está a visão geral de como o sistema funciona internamente.

## 1. Fluxo de Dados (Pipeline)

O processo segue uma linha de montagem linear:

1. **Input (String)**
   O servidor recebe o código Go via requisição HTTP.

2. **Parsing (Análise)**
   Utiliza o pacote nativo `go/parser` para gerar a AST.

   > A AST é uma árvore onde o código deixa de ser texto e vira nós lógicos
   > (ex.: “Declaração de Função”, “Bloco If”).

3. **Traversal (Visita)**
   O arquivo `visitor.go` percorre cada nó da árvore recursivamente.

4. **Generation (Síntese)**
   O `writer.go` reconstrói a lógica usando a sintaxe do Kotlin.

## 2. Responsabilidade dos Módulos

A comunicação entre os arquivos funciona assim:

### **main.go**

* Porta de entrada da aplicação.
* Gerencia o servidor e serve a interface Web.
* Recebe o JSON do frontend, executa o Parser e entrega a AST ao Transpilador.

### **visitor.go**

* O “cérebro” do sistema.
* Contém um grande `switch` que analisa os nós da AST.
* Converte:

  * `ast.IfStmt` → `if`
  * `ast.ForStmt` → `while`

### **types.go**

* Implementa a tabela de conversão de tipos.
* Exemplos:

  * `int` → `Int`
  * `float64` → `Double`
  * Ponteiros viram tipos *nullable* (`?`).

### **writer.go**

* Responsável pela formatação e indentação.
* Garante que o código Kotlin gerado respeite blocos `{ }` e alinhamento visual.

## 3. Status da Implementação

O projeto implementa o **Subset Inicial** da linguagem.

### Implementado (Funcional)

* **Tipos Primitivos:** `int`, `string`, `bool`, `float`.
* **Controle de Fluxo:** `if/else` com parênteses obrigatórios do Kotlin.
* **Loops Seguros:**
  `for` (Go) → bloco `run { while(...) }` para preservar escopo.
* **Funções:**

  * Conversão de sintaxe: `nome tipo` → `nome: Tipo`
  * Conversão de visibilidade: `Func` → `public`, `func` → `internal`
* **Visual Mirroring:**
  O código Kotlin mantém alinhamento vertical do código Go original.

---

## Limitações Atuais (Roadmap)

* **Goroutines/Channels:**
  Ainda não converte `go func()` para Coroutines.
* **Structs / OO:**
  Métodos com receiver viram funções soltas.
* **Tratamento de Erros:**
  Retornos múltiplos `(Data, error)` ainda não viram `Result<Data>`.

## 🚀 Como Rodar

### Pré-requisitos

* Go **1.20+**

### Execução

Clone o repositório e, na raiz, execute:

```sh
go run .
```

Acesse no navegador:

```
http://localhost:8080
```

---

## 📂 Estrutura de Pastas

```
/
├── main.go         # Servidor Web
├── visitor.go      # Lógica de navegação na AST
├── writer.go       # Utilitários de escrita e indentação
├── types.go        # Tabela de conversão de tipos
├── index.html      # Frontend (Editor com Syntax Highlighting)
├── print-go2kt.png # Screenshot do projeto
└── README.md       # Documentação
```