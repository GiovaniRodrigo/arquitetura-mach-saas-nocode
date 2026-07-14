package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
)

type fakeIAM struct{}

func (fakeIAM) AutenticarThirdParty(context.Context, *iamv1.AutenticarThirdPartyRequest, ...grpc.CallOption) (*iamv1.AutenticarThirdPartyResponse, error) {
	return &iamv1.AutenticarThirdPartyResponse{Jwt: "jwt-x"}, nil
}

func handler(t *testing.T) *OAuthHandler {
	t.Helper()
	h := NewOAuthHandler(fakeIAM{}, "https://gfcode.com.br",
		"gid", "gsecret", "hid", "hsecret", "https://gfcode.com.br")
	if h == nil {
		t.Fatal("esperava handler não-nil com credenciais configuradas")
	}
	return h
}

func TestNewOAuthHandler_DesligadoSemCredenciais(t *testing.T) {
	if NewOAuthHandler(fakeIAM{}, "https://x", "", "", "", "", "https://x") != nil {
		t.Fatal("sem credenciais o handler deve ser nil (feature desligada)")
	}
}

func TestLogin_RedirecionaAoProvedor(t *testing.T) {
	r := chi.NewRouter()
	handler(t).Registrar(r)

	req := httptest.NewRequest(http.MethodGet, "/auth/google?redirect_uri=https://gfcode.com.br/app", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "accounts.google.com") || !strings.Contains(loc, "state=r_") {
		t.Fatalf("Location inesperado: %s", loc)
	}
}

func TestLogin_ProvedorDesconhecido(t *testing.T) {
	r := chi.NewRouter()
	handler(t).Registrar(r)
	req := httptest.NewRequest(http.MethodGet, "/auth/facebook", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("provedor desconhecido: status=%d", rec.Code)
	}
}

func TestValidarRedirect_Allowlist(t *testing.T) {
	h := handler(t)
	if got := h.validarRedirect("https://gfcode.com.br/app"); got == "" {
		t.Fatal("origem permitida deveria passar")
	}
	if got := h.validarRedirect("https://evil.example/app"); got != "" {
		t.Fatalf("origem não permitida deveria ser bloqueada; got=%q", got)
	}
}

func TestDestino_FallbackSemStateValido(t *testing.T) {
	h := handler(t)
	if got := h.destino("mach-state"); got != "https://gfcode.com.br/app" {
		t.Fatalf("fallback esperado; got=%q", got)
	}
	if got := h.destino(codificarState("https://gfcode.com.br/app")); got != "https://gfcode.com.br/app" {
		t.Fatalf("state válido deveria voltar a origem permitida; got=%q", got)
	}
	// state carregando origem fora da allowlist → fallback
	if got := h.destino(codificarState("https://evil.example/x")); got != "https://gfcode.com.br/app" {
		t.Fatalf("origem não permitida deveria cair no fallback; got=%q", got)
	}
}
