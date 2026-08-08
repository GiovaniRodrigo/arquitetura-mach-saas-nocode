// Package web centraliza a escrita de respostas HTTP do Gateway. Erros seguem o
// formato comum de api.md e nunca expõem nomes reais de campos (RNF08).
package web

import (
	"encoding/json"
	"net/http"
)

// ErrorBody é o corpo padrão de erro.
type ErrorBody struct {
	Codigo   string `json:"codigo"`
	Mensagem string `json:"mensagem"`
}

// JSON escreve v como resposta JSON com o status informado.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error escreve um erro no formato comum.
func Error(w http.ResponseWriter, status int, codigo, mensagem string) {
	JSON(w, status, ErrorBody{Codigo: codigo, Mensagem: mensagem})
}
