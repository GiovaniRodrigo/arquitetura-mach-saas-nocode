package rag

import (
	"math"
	"strings"
	"unicode"
)

// indiceFrequencia guarda, para cada termo do corpus, em quantos documentos
// ele aparece — usado para pesar termos raros (mais informativos) acima de
// termos comuns, sem precisar de um índice externo.
var indiceFrequencia = construirIndice(corpus)

func construirIndice(docs []Documento) map[string]int {
	idx := make(map[string]int)
	for _, d := range docs {
		vistos := make(map[string]bool)
		for _, termo := range tokenizar(d.Conteudo) {
			if !vistos[termo] {
				vistos[termo] = true
				idx[termo]++
			}
		}
	}
	return idx
}

// tokenizar reduz um texto a palavras minúsculas de ao menos 3 caracteres,
// descartando pontuação — suficiente para um corpus pequeno e curado em
// português; não precisa de stemming para o volume de documentos envolvido.
func tokenizar(texto string) []string {
	campos := strings.FieldsFunc(strings.ToLower(texto), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	termos := make([]string, 0, len(campos))
	for _, c := range campos {
		if len([]rune(c)) >= 3 {
			termos = append(termos, c)
		}
	}
	return termos
}

// Recuperar devolve os `limite` documentos mais relevantes para `consulta`
// (a pergunta do usuário, opcionalmente concatenada ao nome/descrição do
// sistema em foco), pontuados por TF·IDF simples. Nunca devolve lista vazia
// enquanto houver corpus: sem nenhum termo em comum, cai para os primeiros
// documentos do corpus — o assistente sempre recebe algum contexto de
// design de sistemas, mesmo numa pergunta genérica ("por onde eu começo?").
func Recuperar(consulta string, limite int) []Documento {
	if limite <= 0 || len(corpus) == 0 {
		return nil
	}

	termosConsulta := tokenizar(consulta)
	pontuacoes := make([]float64, len(corpus))
	algumaPontuacao := false

	for i, doc := range corpus {
		freqDoc := make(map[string]int)
		for _, t := range tokenizar(doc.Conteudo) {
			freqDoc[t]++
		}
		var pontos float64
		for _, termo := range termosConsulta {
			tf := freqDoc[termo]
			if tf == 0 {
				continue
			}
			// idf: quanto mais raro o termo no corpus, maior o peso.
			idf := math.Log(1 + float64(len(corpus))/float64(indiceFrequencia[termo]))
			pontos += float64(tf) * idf
		}
		pontuacoes[i] = pontos
		if pontos > 0 {
			algumaPontuacao = true
		}
	}

	indices := make([]int, len(corpus))
	for i := range indices {
		indices[i] = i
	}
	if algumaPontuacao {
		// Ordena por pontuação desc.; empate mantém a ordem estável do corpus
		// (já ordenado por ID) para resultados determinísticos.
		for i := 1; i < len(indices); i++ {
			for j := i; j > 0 && pontuacoes[indices[j]] > pontuacoes[indices[j-1]]; j-- {
				indices[j], indices[j-1] = indices[j-1], indices[j]
			}
		}
	}

	if limite > len(indices) {
		limite = len(indices)
	}
	resultado := make([]Documento, 0, limite)
	for _, i := range indices[:limite] {
		if algumaPontuacao && pontuacoes[i] == 0 {
			break // não força documentos irrelevantes quando já há sinal
		}
		resultado = append(resultado, corpus[i])
	}
	return resultado
}
