package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// TenantsCliente é o subconjunto do IAMServiceClient usado por estas rotas.
type TenantsCliente interface {
	ListarTenants(ctx context.Context, in *iamv1.ListarTenantsRequest, opts ...grpc.CallOption) (*iamv1.ListarTenantsResponse, error)
	CriarTenant(ctx context.Context, in *iamv1.CriarTenantRequest, opts ...grpc.CallOption) (*iamv1.Tenant, error)
	ObterTenant(ctx context.Context, in *iamv1.ObterTenantRequest, opts ...grpc.CallOption) (*iamv1.Tenant, error)
	AtualizarTenant(ctx context.Context, in *iamv1.AtualizarTenantRequest, opts ...grpc.CallOption) (*iamv1.Tenant, error)
	ExcluirTenant(ctx context.Context, in *iamv1.ExcluirTenantRequest, opts ...grpc.CallOption) (*iamv1.ExcluirTenantResponse, error)
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

// atualizarTenantReq é o corpo de PATCH /api/v1/tenants/{id}.
type atualizarTenantReq struct {
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

// ObterTenant serve GET /api/v1/tenants/{id} (spec 004, RF07): devolve um
// cliente específico sob o tenant do contexto. Um id inexistente ou fora da
// hierarquia do contexto vira 404 (o IAM não distingue os dois casos, RN05).
func ObterTenant(iam TenantsCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := iam.ObterTenant(r.Context(), &iamv1.ObterTenantRequest{Id: chi.URLParam(r, "id")})
		if err != nil {
			writeTenantError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, tenantResp{ID: out.GetId(), Nome: out.GetNome()})
	}
}

// AtualizarTenant serve PATCH /api/v1/tenants/{id} (spec 004, RF07): renomeia
// um cliente sob o tenant do contexto. Mesma regra de permissão/hierarquia de
// ObterTenant.
func AtualizarTenant(iam TenantsCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body atualizarTenantReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if body.Nome == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "nome obrigatório")
			return
		}
		out, err := iam.AtualizarTenant(r.Context(), &iamv1.AtualizarTenantRequest{
			Id:   chi.URLParam(r, "id"),
			Nome: body.Nome,
		})
		if err != nil {
			writeTenantError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, tenantResp{ID: out.GetId(), Nome: out.GetNome()})
	}
}

// ExcluirTenant serve DELETE /api/v1/tenants/{id} (spec 004, RF07): exclui um
// cliente sob o tenant do contexto — em cascata, todos os sistemas/dados dele
// (ver services/iam/internal/store.ExcluirTenant).
func ExcluirTenant(iam TenantsCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := iam.ExcluirTenant(r.Context(), &iamv1.ExcluirTenantRequest{Id: chi.URLParam(r, "id")}); err != nil {
			writeTenantError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// writeTenantError traduz os códigos gRPC do IAM para HTTP nas rotas de
// tenants: PermissionDenied (cliente final) → 403; InvalidArgument → 400;
// NotFound (inexistente ou fora da hierarquia) → 404.
func writeTenantError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "nome obrigatório")
	case codes.PermissionDenied:
		web.Error(w, http.StatusForbidden, "FORBIDDEN", "sem permissão para gerenciar clientes")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	case codes.NotFound:
		web.Error(w, http.StatusNotFound, "NOT_FOUND", "cliente não encontrado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
