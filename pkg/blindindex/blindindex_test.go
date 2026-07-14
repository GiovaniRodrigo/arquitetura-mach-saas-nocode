package blindindex

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestGenerate_KnownVector(t *testing.T) {
	key := []byte("chave-do-tenant-A")
	got, err := Generate(key, "email")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("email"))
	want := hex.EncodeToString(mac.Sum(nil))

	if got != want {
		t.Fatalf("Blind Index divergente:\n got=%s\nwant=%s", got, want)
	}
	if len(got) != 64 {
		t.Fatalf("Blind Index deve ter 64 chars (cabe em varchar(64)); got=%d", len(got))
	}
}

func TestGenerate_IsolamentoPorTenant(t *testing.T) {
	// O mesmo nome em claro sob chaves de tenants diferentes NÃO pode colidir (RN02).
	a, _ := Generate([]byte("chave-tenant-A"), "cpf")
	b, _ := Generate([]byte("chave-tenant-B"), "cpf")
	if a == b {
		t.Fatal("Blind Index igual entre tenants distintos — correlação cruzada possível")
	}
}

func TestGenerate_EmptyKey(t *testing.T) {
	if _, err := Generate(nil, "qualquer"); err != ErrEmptyKey {
		t.Fatalf("esperava ErrEmptyKey; got=%v", err)
	}
}

func TestVerify(t *testing.T) {
	key := []byte("chave")
	bi, _ := Generate(key, "telefone")

	ok, err := Verify(key, "telefone", bi)
	if err != nil || !ok {
		t.Fatalf("Verify deveria validar o próprio índice; ok=%v err=%v", ok, err)
	}
	if ok, _ := Verify(key, "outro-campo", bi); ok {
		t.Fatal("Verify validou plaintext incorreto")
	}
	if ok, _ := Verify([]byte("chave-errada"), "telefone", bi); ok {
		t.Fatal("Verify validou sob chave de outro tenant")
	}
}
