# Go2Kotlin Transpiler

<p align="center"> <img src="web/static/print-go2kt.png" alt="Go2Kotlin Demo" width="650"> </p> <p align="center"> <img src="https://img.shields.io/badge/Language-Go-blue?logo=go" /> <img src="https://img.shields.io/badge/Target-Kotlin-purple?logo=kotlin" /> <img src="https://img.shields.io/badge/Status-In%20Development-yellow" /> <img src="https://img.shields.io/badge/Build-Source%20to%20Source-4caf50" /></p>

## Projeto Acadêmico

Este projeto faz parte da disciplina **Compiladores e Paradigmas de Programação**, ministrada pelo professor **Sebastião Filho**, na **Universidade do Estado do Rio Grande do Norte (UERN)**.

---

## Grupo de Desenvolvimento

* **Paulo Sérgio**
* **Eduardo Marinho**
* **Marlos Emanuel**
* **Vinicius Eduardo**
* **Luiz Henrique**

---

## Sobre o Projeto

O **Go2Kotlin Transpiler** é um transpilador *source-to-source* que converte código **Golang → Kotlin**, mantendo a estrutura e a lógica do código original.

---

## Arquitetura e Evolução

Este projeto **não** cria binários executáveis; ele traduz a **Árvore Sintática Abstrata (AST)** de Go para Kotlin.
Abaixo está um resumo da evolução arquitetural do projeto.

---

### 1. Histórico de Evolução (Refatoração)

#### Como era antes — *Abordagem Monolítica*

* Toda a lógica de tradução estava centralizada em `visitor.go`.
* Um grande `switch case` controlava todos os tipos de nós.
* **Problemas:**

  * Crescimento infinito do arquivo.
  * Violação do princípio **OCP (Open/Closed Principle)**.
  * Dificuldade para testar e manter.

#### Como é agora — *Abordagem Strategy/Handler*

* O padrão de projeto **Strategy** foi adotado.
* O `visitor.go` virou apenas um **dispatcher**.
* Cada tipo de nó da AST agora possui um **handler especializado**.

**Benefícios:**

* Código modular.
* Fácil de estender sem tocar no núcleo.
* Testabilidade muito superior.

---

### 2. Fluxo de Dados (Pipeline Atual)

1. **Input**
   O servidor recebe o código Go via HTTP (JSON).

2. **Parsing**
   O pacote `go/parser` gera a AST.

3. **Traversal**
   O módulo `visitor` percorre cada nó.

4. **Strategy**
   O visitor consulta o mapa de handlers e delega o nó.

5. **Generation**
   O módulo `writer` produz a saída Kotlin formatada.

---

## Responsabilidade dos Módulos

O projeto segue o **Standard Go Project Layout**.

```
cmd/server/main.go
    → Entrada da aplicação, servidor HTTP, integração com o web editor

pkg/transpiler/
    visitor.go   → Dispatcher da AST
    handlers.go  → Estratégias de tradução (por tipo de nó)
    writer.go    → Formatação, indentação e estado
    types.go     → Tabela de conversão de tipos Go → Kotlin
```

---

## Status da Implementação

### Implementado

#### **Tipos Primitivos**

* `int → Int`
* `string → String`
* `bool → Boolean`
* `float64 → Double`

#### **Controle de Fluxo**

* `if/else`
  Com conversão obrigatória para parênteses Kotlin.

#### **Loops**

* `for` do Go convertido para:

  ```kotlin
  run {
      while (...) {
      }
  }
  ```

#### **Funções**

* Conversão Go → Kotlin:

  ```
  nome tipo → nome: Tipo
  ```
* Converte visibilidade:

  * `Func` → `public`
  * `func` → `internal`
* Mantém alinhamento visual com o código original ("visual mirroring").

---

### Limitações Atuais (Roadmap)

* **Goroutines / Channels**
  Suporte experimental → mapeado para Coroutines.

* **Structs / Methods**
  Receivers viram *extension functions*.

* **Tratamento de Erros**
  Retornos múltiplos ainda não mapeados para `Result<T>` ou similar.

---

## Como Rodar

### Acesse Online(Deploy na Vercel)

[Go2Kotlin Link](https://go2kotlin.vercel.app/)

### Pré-requisitos

* **Go 1.20+**

---

### Linux / macOS

#### Modo Dev

```bash
go run cmd/server/main.go
```

#### Produção

```bash
go build -o server cmd/server/main.go
./server
```

---

### Windows

#### Modo Dev

```powershell
go run cmd\server\main.go
```

#### Produção

```powershell
go build -o server.exe cmd\server\main.go
.\server.exe
```

---

### Acesse no navegador

```
http://localhost:8080
```

---

## Estrutura de Pastas (Standard Go Layout)

```
/
├── cmd/
│   └── server/
│       └── main.go          # Entry point do servidor
│
├── pkg/
│   └── transpiler/          # Lógica CORE da transpilação
│       ├── visitor.go       # Dispatcher (Visitor)
│       ├── handlers.go      # Estratégias de tradução (Strategy)
│       ├── writer.go        # Estado e formatação
│       └── types.go         # Mapeamento de tipos Go → Kotlin
│
├── web/
│   ├── templates/
│   │   └── index.html       # Editor web
│   └── static/
│       └── img/             # Assets
│
├── examples/                # Código Go para testes
└── README.md                # Documentação do projeto
```

## Pasta `examples/` — Casos de Testes Reais

A pasta **`examples/`** contém diversos códigos Go utilizados para testar o transpilador.
Eles abrangem desde exemplos simples até casos avançados que ainda não possuem suporte completo.

### O que você encontrará lá:

### Exemplos que funcionam corretamente

Códigos Go totalmente suportados pelo transpilador.
Servem para validar a conversão de estruturas como funções, condicionais, loops, tipos primitivos etc.

---

### Exemplos que **ainda não funcionam**

Alguns arquivos demonstram recursos da linguagem Go que o transpilador **ainda não suporta**.
Cada um desses exemplos contém um **comentário explicando o motivo**, como:

* uso de *goroutines* e *channels* complexos
* retornos múltiplos avançados
* ponteiros
* interfaces
* struct e métodos específicos
* closures mais sofisticadas

---

### Exemplos com **TODOs**

Há arquivos que **funcionam parcialmente**, mas possuem trechos com a anotação:

```go
// TODO: implementar suporte para ...
```

Esses casos representam funcionalidades que estão no **roadmap oficial do projeto**, como:

* mapeamento completo de coroutines
* extension functions para receivers de structs
* tradução de interfaces para sealed classes
* tratamento de erros e Result<T>
* suporte ao pacote `go/types` mais profundo

---

### Finalidade da pasta `examples/`

Ela existe para:

* demonstrar ao usuário o estágio atual do transpiler
* auxiliar no desenvolvimento e debug
* servir como benchmark de regressão
* guiar contribuições futuras (PRs / melhorias)

---

