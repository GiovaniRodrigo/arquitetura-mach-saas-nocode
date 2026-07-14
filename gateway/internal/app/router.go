// Package app monta o roteador HTTP do Gateway, compartilhado entre o binário e
// os testes de integração.
package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/gateway/internal/middleware"
	"github.com/machv4/platform/gateway/internal/routes"
)

// NewRouter compõe a cadeia de middlewares e as rotas REST→gRPC.
//
// Ordem: Tracing (span raiz) é o mais externo; dentro do grupo autenticado,
// Auth valida o JWT e injeta o TenantContext, e só então o RateLimiter aplica a
// cota por tenant. /health fica fora da autenticação.
func NewRouter(iam iamv1.IAMServiceClient, design designv1.DesignEngineServiceClient, rl *middleware.RateLimiter) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Tracing)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(iam))
		r.Use(rl.Handler)

		r.Get("/api/v1/permissoes", routes.Permissoes(iam))

		r.Post("/api/v1/designs", routes.CriarDesign(design))
		r.Get("/api/v1/designs/{id}", routes.ObterDesign(design))
		r.Put("/api/v1/designs/{id}", routes.AtualizarDesign(design))
		r.Delete("/api/v1/designs/{id}", routes.RemoverDesign(design))
	})

	return r
}
