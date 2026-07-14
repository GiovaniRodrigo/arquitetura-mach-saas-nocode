// Package middleware reúne os middlewares HTTP do Gateway.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/gateway/internal/web"
	"github.com/machv4/platform/pkg/tenantctx"
)

// IAMValidator é a dependência de autenticação — apenas ValidarToken. O
// *IAMServiceClient gerado a satisfaz.
type IAMValidator interface {
	ValidarToken(ctx context.Context, in *iamv1.ValidarTokenRequest, opts ...grpc.CallOption) (*iamv1.ValidarTokenResponse, error)
}

// Auth valida o header Authorization: Bearer <JWT> chamando o IAM Service e, em
// caso de sucesso, anexa o TenantContext ao contexto da requisição (RF03, RNF02).
// O tenantctx client interceptor depois o propaga aos serviços gRPC como Metadata.
func Auth(iam IAMValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearer(r)
			if !ok {
				web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "token ausente")
				return
			}
			resp, err := iam.ValidarToken(r.Context(), &iamv1.ValidarTokenRequest{Jwt: token})
			if err != nil || !resp.GetValido() {
				web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "token inválido ou expirado")
				return
			}
			tc := &commonv1.TenantContext{
				TenantId: resp.GetTenantId(),
				UserId:   resp.GetUserId(),
				Tipo:     resp.GetTipo(),
			}
			next.ServeHTTP(w, r.WithContext(tenantctx.NewContext(r.Context(), tc)))
		})
	}
}

func bearer(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	const p = "Bearer "
	if len(h) <= len(p) || !strings.EqualFold(h[:len(p)], p) {
		return "", false
	}
	return strings.TrimSpace(h[len(p):]), true
}
