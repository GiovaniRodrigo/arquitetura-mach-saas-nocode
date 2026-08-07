package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/gateway/internal/web"
)

// TenantsCliente é o subconjunto do IAMServiceClient usado por estas rotas.
type TenantsCliente interface {
	ListarTenants(ctx context.Context, in *iamv1.ListarTenantsRequest, opts ...grpc.CallOption) (*iamv1.ListarTenantsResponse, error)
	CriarTenant(ctx context.Context, in *iamv1.CriarTenantRequest, opts ...grpc.CallOption) (*iamv1.Tenant, error)
}

// tenantResp espelha um tenant no corpo JSON.
type tenantResp struct {
	ID   string `json:"id"`
	Nome string `json:"nome"`
}

// criarTenantReq é o corpo de POST /api/v1/tenants.
type criarTenantReq struct {
	Nome string `json:"nome"`
}

// ListarTenants serve GET /api/v1/tenants (spec 004, RF07): devolve os tenants
// (clientes/negócios) filhos do tenant do contexto. O tenant vem do
// TenantContext (posto pelo Auth) e é propagado ao IAM via Metadata gRPC.
func ListarTenants(iam TenantsCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := iam.ListarTenants(r.Context(), &iamv1.ListarTenantsRequest{})
		if err != nil {
			writeTenantError(w, err)
			return
		}
		lista := make([]tenantResp, 0, len(out.GetTenants()))
		for _, t := range out.GetTenants() {
			lista = append(lista, tenantResp{ID: t.GetId(), Nome: t.GetNome()})
		}
		web.JSON(w, http.StatusOK, map[string]any{"tenants": lista})
	}
}

// CriarTenant serve POST /api/v1/tenants (spec 004, RF07): cria um novo tenant
// cliente sob o tenant do contexto. Restrito a dono/parceiro no IAM →
// PermissionDenied vira 403 (mesma regra de CriarSistema em 001, RN01).
func CriarTenant(iam TenantsCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body criarTenantReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if body.Nome == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "nome obrigatório")
			return
		}
		out, err := iam.CriarTenant(r.Context(), &iamv1.CriarTenantRequest{Nome: body.Nome})
		if err != nil {
			writeTenantError(w, err)
			return
		}
		web.JSON(w, http.StatusCreated, tenantResp{ID: out.GetId(), Nome: out.GetNome()})
	}
}

// writeTenantError traduz os códigos gRPC do IAM para HTTP nas rotas de
// tenants: PermissionDenied (cliente final) → 403; InvalidArgument → 400.
func writeTenantError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "nome obrigatório")
	case codes.PermissionDenied:
		web.Error(w, http.StatusForbidden, "FORBIDDEN", "sem permissão para criar tenant")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
