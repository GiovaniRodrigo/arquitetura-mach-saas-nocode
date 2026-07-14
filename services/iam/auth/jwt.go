// Package auth emite e valida os JWT da plataforma (RF03).
//
// Os tokens são assinados com RS256 (chave assimétrica): apenas o IAM Service
// detém a chave privada para emitir; qualquer serviço pode validar com a chave
// pública. Os claims carregam a identidade que depois vira TenantContext no
// Gateway: tenant_id, sub (user id) e tipo (dono|parceiro|cliente).
package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrTokenInvalido cobre assinatura inválida, expiração e formato malformado.
var ErrTokenInvalido = errors.New("auth: token inválido")

// Claims são os dados de identidade transportados no JWT. Subject (sub) é o user id.
type Claims struct {
	TenantID string `json:"tenant_id"`
	Tipo     string `json:"tipo"`
	jwt.RegisteredClaims
}

// GenerateKey gera um par de chaves RSA para desenvolvimento/testes.
func GenerateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// Issuer emite tokens assinados. Vive apenas no IAM Service.
type Issuer struct {
	priv *rsa.PrivateKey
	ttl  time.Duration
	now  func() time.Time
}

// NewIssuer cria um emissor com a chave privada e o tempo de vida dos tokens.
func NewIssuer(priv *rsa.PrivateKey, ttl time.Duration) *Issuer {
	return &Issuer{priv: priv, ttl: ttl, now: time.Now}
}

// Issue assina um JWT RS256 para o utilizador. tipo deve ser dono|parceiro|cliente.
func (i *Issuer) Issue(userID, tenantID, tipo string) (string, error) {
	if userID == "" || tenantID == "" {
		return "", fmt.Errorf("auth: userID e tenantID obrigatórios")
	}
	now := i.now()
	claims := Claims{
		TenantID: tenantID,
		Tipo:     tipo,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(i.priv)
}

// Validator verifica tokens com a chave pública. Vive em qualquer serviço.
type Validator struct {
	pub *rsa.PublicKey
}

// NewValidator cria um validador a partir da chave pública correspondente.
func NewValidator(pub *rsa.PublicKey) *Validator {
	return &Validator{pub: pub}
}

// Validate verifica assinatura e expiração e devolve os claims. Recusa qualquer
// método de assinatura diferente de RS256 (defesa contra o ataque alg=none/HS256).
func (v *Validator) Validate(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: método de assinatura inesperado %q", ErrTokenInvalido, t.Header["alg"])
		}
		return v.pub, nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrTokenInvalido
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || claims.TenantID == "" || claims.Subject == "" {
		return nil, ErrTokenInvalido
	}
	return claims, nil
}
