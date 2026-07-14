// Package blindindex gera e verifica Blind Indexes (RN02, RNF08).
//
// Um Blind Index é o HMAC-SHA256 do identificador em claro de um componente
// (nome real de campo/tabela) usando a chave HMAC específica do tenant. É ele —
// e nunca o nome real — que trafega em payloads, logs e traces, e que indexa
// campos_definicao. Como a chave é por tenant, o mesmo nome em claro produz
// Blind Indexes distintos entre tenants, impedindo correlação cruzada.
package blindindex

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// ErrEmptyKey é retornado quando a chave HMAC do tenant está vazia. Gerar um
// Blind Index sem chave por tenant violaria o isolamento (RN02), por isso falha.
var ErrEmptyKey = errors.New("blindindex: chave HMAC do tenant vazia")

// Generate produz o Blind Index de plaintext sob tenantKey.
//
// O resultado é o HMAC-SHA256 (32 bytes) codificado em hexadecimal — 64 caracteres,
// exatamente o tamanho de campos_definicao.blind_index (varchar(64)).
func Generate(tenantKey []byte, plaintext string) (string, error) {
	if len(tenantKey) == 0 {
		return "", ErrEmptyKey
	}
	mac := hmac.New(sha256.New, tenantKey)
	mac.Write([]byte(plaintext))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// Verify compara, em tempo constante, um Blind Index candidato com o esperado
// para plaintext sob tenantKey. Retorna erro apenas se a chave for inválida.
func Verify(tenantKey []byte, plaintext, candidate string) (bool, error) {
	expected, err := Generate(tenantKey, plaintext)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(expected), []byte(candidate)), nil
}
