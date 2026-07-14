package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueValidate_RoundTrip(t *testing.T) {
	priv, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	iss := NewIssuer(priv, time.Hour)
	val := NewValidator(&priv.PublicKey)

	token, err := iss.Issue("user-1", "tenant-A", "parceiro")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	claims, err := val.Validate(token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.Subject != "user-1" || claims.TenantID != "tenant-A" || claims.Tipo != "parceiro" {
		t.Fatalf("claims incorretos: %+v", claims)
	}
}

func TestValidate_ChaveDeOutroEmissor(t *testing.T) {
	priv, _ := GenerateKey()
	outra, _ := GenerateKey()

	token, _ := NewIssuer(priv, time.Hour).Issue("u", "t", "dono")

	if _, err := NewValidator(&outra.PublicKey).Validate(token); err != ErrTokenInvalido {
		t.Fatalf("token de outro emissor deveria ser rejeitado; err=%v", err)
	}
}

func TestValidate_Expirado(t *testing.T) {
	priv, _ := GenerateKey()
	iss := NewIssuer(priv, time.Hour)
	iss.now = func() time.Time { return time.Now().Add(-2 * time.Hour) } // emitido no passado

	token, _ := iss.Issue("u", "t", "cliente")
	if _, err := NewValidator(&priv.PublicKey).Validate(token); err != ErrTokenInvalido {
		t.Fatalf("token expirado deveria ser rejeitado; err=%v", err)
	}
}

// Um token forjado com alg=HS256 usando a chave pública como segredo NÃO pode
// ser aceito (ataque clássico de confusão de algoritmo).
func TestValidate_RejeitaAlgHS256(t *testing.T) {
	priv, _ := GenerateKey()
	pub := &priv.PublicKey
	pubBytes := pub.N.Bytes()

	forjado := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		TenantID: "t", RegisteredClaims: jwt.RegisteredClaims{Subject: "u"},
	})
	s, err := forjado.SignedString(pubBytes)
	if err != nil {
		t.Skipf("não foi possível forjar token: %v", err)
	}
	if _, err := NewValidator(pub).Validate(s); err != ErrTokenInvalido {
		t.Fatalf("token HS256 forjado deveria ser rejeitado; err=%v", err)
	}
}

func TestValidate_Malformado(t *testing.T) {
	priv, _ := GenerateKey()
	if _, err := NewValidator(&priv.PublicKey).Validate("nao.e.um.jwt"); err != ErrTokenInvalido {
		t.Fatalf("esperava ErrTokenInvalido; got=%v", err)
	}
}
