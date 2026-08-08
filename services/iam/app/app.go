// Package app monta o servidor IAM já cabeado com validador de JWT e store,
// para uso pelo binário e por testes de integração de outros módulos (o pacote é
// público, ao contrário de internal/*).
package app

import (
	"crypto/rsa"
	"time"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/services/iam/auth"
	"github.com/machv4/platform/services/iam/internal/server"
	"github.com/machv4/platform/services/iam/internal/store"
)

// NewServer devolve o IAMServiceServer que valida e emite JWT com a chave privada
// priv (a pública é derivada dela) e persiste identidades/permissões no pool
// informado (satisfaz store.DB — ex.: *pgxpool.Pool). ttl é o tempo de vida do JWT.
//
// Não liga a chave de cifra do MFA (spec 004, RF15) — as rotas de Ativar/
// ConfirmarMfa operariam com uma chave zero-value. Mantido para não quebrar os
// chamadores existentes que não exercitam MFA; use NewServerComMfa em produção
// e em qualquer teste que precise validar o fluxo de MFA de ponta a ponta.
func NewServer(priv *rsa.PrivateKey, ttl time.Duration, pool store.DB) iamv1.IAMServiceServer {
	return NewServerComMfa(priv, ttl, pool, [32]byte{})
}

// NewServerComMfa é NewServer mais a chave AES-256-GCM usada para cifrar/
// decifrar segredos TOTP em repouso (spec 004, RF15; services/iam/auth/mfa.go).
// O binário (services/iam/cmd/main.go) sempre chama esta função — nunca
// NewServer — e recusa subir sem IAM_MFA_ENCRYPTION_KEY (sem fallback
// efêmero: um segredo TOTP recriado a cada restart invalidaria os apps
// autenticadores de todos os usuários, pior que travar o boot).
func NewServerComMfa(priv *rsa.PrivateKey, ttl time.Duration, pool store.DB, mfaKey [32]byte) iamv1.IAMServiceServer {
	st := store.New(pool)
	srv := server.New(auth.NewValidator(&priv.PublicKey), auth.NewIssuer(priv, ttl), st)
	return srv.ComChaveMfa(mfaKey)
}
