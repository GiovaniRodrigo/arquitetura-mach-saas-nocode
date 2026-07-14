// Package app monta o roteador HTTP do Gateway, compartilhado entre o binário e
// os testes de integração.
package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	deployv1 "github.com/machv4/platform/gen/go/construtor/deploy/v1"
	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	exportv1 "github.com/machv4/platform/gen/go/construtor/export/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/gateway/internal/middleware"
	"github.com/machv4/platform/gateway/internal/routes"
)

// NewRouter compõe a cadeia de middlewares e as rotas REST→gRPC.
//
// Ordem: Tracing (span raiz) é o mais externo; dentro do grupo autenticado,
// Auth valida o JWT e injeta o TenantContext, e só então o RateLimiter aplica a
// cota por tenant. /health fica fora da autenticação.
func NewRouter(iam iamv1.IAMServiceClient, design designv1.DesignEngineServiceClient, logic logicv1.LogicEngineServiceClient, deploy deployv1.DeployEngineServiceClient, export exportv1.ExportEngineServiceClient, rl *middleware.RateLimiter, oauth *routes.OAuthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Tracing)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Login social — rotas públicas (não autenticadas), fora do grupo com Auth.
	if oauth != nil {
		oauth.Registrar(r)
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(iam))
		r.Use(rl.Handler)

		r.Get("/api/v1/permissoes", routes.Permissoes(iam))

		r.Post("/api/v1/designs", routes.CriarDesign(design))
		r.Get("/api/v1/designs/{id}", routes.ObterDesign(design))
		r.Put("/api/v1/designs/{id}", routes.AtualizarDesign(design))
		r.Delete("/api/v1/designs/{id}", routes.RemoverDesign(design))

		r.Post("/api/v1/regras", routes.CriarRegra(logic))
		r.Get("/api/v1/regras/{id}", routes.ObterRegra(logic))
		r.Put("/api/v1/regras/{id}", routes.AtualizarRegra(logic))
		r.Delete("/api/v1/regras/{id}", routes.RemoverRegra(logic))

		r.Post("/api/v1/formularios", routes.SalvarFormulario(logic))

		r.Post("/api/v1/sistemas/{sistema_id}/publicar", routes.Publicar(deploy))
		r.Post("/api/v1/sistemas/{sistema_id}/rollback", routes.Rollback(deploy))
		r.Get("/api/v1/sistemas/{sistema_id}/versao-ativa", routes.VersaoAtiva(deploy))

		r.Post("/api/v1/exportacoes", routes.CriarExportacao(export))
		r.Get("/api/v1/exportacoes/{job_id}", routes.ObterExportacao(export))
	})

	return r
}
