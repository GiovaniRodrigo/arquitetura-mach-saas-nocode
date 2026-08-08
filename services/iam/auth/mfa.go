package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

// ErrSegredoMfaInvalido cobre segredo cifrado corrompido/truncado ou decifrado
// com a chave errada — GCM falha a autenticação e não distinguimos as causas
// (nenhuma delas deve ser tratada de forma diferente pelo chamador).
var ErrSegredoMfaInvalido = errors.New("auth: segredo mfa inválido")

// CifrarSegredo cifra o segredo TOTP em claro com AES-256-GCM antes de
// persistir em users.mfa_segredo_cifrado (nunca em texto claro, RNF01). O
// nonce aleatório vai prefixado ao ciphertext (nonce || ciphertext || tag), o
// que dispensa uma coluna separada para guardá-lo.
func CifrarSegredo(chave [32]byte, segredoClaro string) ([]byte, error) {
	gcm, err := novoGCM(chave)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("auth: gerar nonce mfa: %w", err)
	}
	return gcm.Seal(nonce, nonce, []byte(segredoClaro), nil), nil
}

// DecifrarSegredo reverte CifrarSegredo. Qualquer falha (chave errada, dado
// corrompido, tamanho inválido) devolve ErrSegredoMfaInvalido — nunca o erro
// de baixo nível do pacote crypto, para não vazar detalhes de implementação.
func DecifrarSegredo(chave [32]byte, cifrado []byte) (string, error) {
	gcm, err := novoGCM(chave)
	if err != nil {
		return "", err
	}
	if len(cifrado) < gcm.NonceSize() {
		return "", ErrSegredoMfaInvalido
	}
	nonce, texto := cifrado[:gcm.NonceSize()], cifrado[gcm.NonceSize():]
	claro, err := gcm.Open(nil, nonce, texto, nil)
	if err != nil {
		return "", ErrSegredoMfaInvalido
	}
	return string(claro), nil
}

func novoGCM(chave [32]byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(chave[:])
	if err != nil {
		return nil, fmt.Errorf("auth: criar cipher aes-256 para mfa: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("auth: criar gcm para mfa: %w", err)
	}
	return gcm, nil
}
