// Package rag implementa a base de conhecimento e a recuperação (retrieval)
// usadas pelo Assistente de Design (spec chat-ia-rag): um corpus curado de
// boas práticas de arquitetura/design de sistemas, embutido no binário do
// Gateway via go:embed, recuperado por similaridade de palavras-chave e
// injetado no prompt enviado à Anthropic — sem depender de um vector DB
// externo, adequado ao tamanho do corpus (dezenas de documentos curados).
package rag

import (
	"embed"
	"path"
	"sort"
	"strings"
)

//go:embed docs/*.md
var arquivosDocs embed.FS

// Documento é um item da base de conhecimento: um markdown curado sobre um
// tópico de design/arquitetura de sistemas.
type Documento struct {
	ID       string
	Titulo   string
	Conteudo string
}

// corpus é carregado uma vez no import do pacote — o conteúdo é estático
// (embutido no binário), não há custo em relê-lo por requisição.
var corpus = carregarCorpus()

func carregarCorpus() []Documento {
	entradas, err := arquivosDocs.ReadDir("docs")
	if err != nil {
		// Não deveria acontecer (diretório embutido em tempo de build), mas
		// não derruba o processo por isso — o retriever devolve vazio.
		return nil
	}
	docs := make([]Documento, 0, len(entradas))
	for _, e := range entradas {
		if e.IsDir() {
			continue
		}
		bytes, err := arquivosDocs.ReadFile(path.Join("docs", e.Name()))
		if err != nil {
			continue
		}
		conteudo := string(bytes)
		docs = append(docs, Documento{
			ID:       strings.TrimSuffix(e.Name(), ".md"),
			Titulo:   tituloDoMarkdown(conteudo),
			Conteudo: conteudo,
		})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	return docs
}

// tituloDoMarkdown extrai o "# Título" da primeira linha do documento, para
// exibição/legibilidade — cai no ID do arquivo se o markdown não tiver H1.
func tituloDoMarkdown(conteudo string) string {
	primeiraLinha, _, _ := strings.Cut(conteudo, "\n")
	if t, ok := strings.CutPrefix(strings.TrimSpace(primeiraLinha), "# "); ok {
		return t
	}
	return primeiraLinha
}
