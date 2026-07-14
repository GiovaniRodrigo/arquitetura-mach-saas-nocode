// Package tenantctx transporta o contexto de tenant/identidade entre serviços.
//
// O contrato (RNF02) exige que a identidade trafegue exclusivamente como
// Metadata binário gRPC sob a chave "x-tenant-context-bin" — nunca no corpo das
// mensagens de negócio. Este pacote centraliza a (de)serialização e oferece
// interceptores cliente/servidor que propagam o TenantContext automaticamente,
// para que nenhum serviço reimplemente esse transporte.
package tenantctx

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
)

// MetadataKey é a chave do Metadata binário gRPC. O sufixo "-bin" faz o grpc-go
// tratar o valor como binário (base64 no transporte, bytes crus na aplicação).
const MetadataKey = "x-tenant-context-bin"

type ctxKey struct{}

// ErrNoTenantContext indica ausência de contexto de tenant onde ele era exigido.
var ErrNoTenantContext = errors.New("tenantctx: contexto de tenant ausente")

// NewContext anexa o TenantContext ao context.Context local.
func NewContext(ctx context.Context, tc *commonv1.TenantContext) context.Context {
	return context.WithValue(ctx, ctxKey{}, tc)
}

// FromContext recupera o TenantContext anexado, se houver.
func FromContext(ctx context.Context) (*commonv1.TenantContext, bool) {
	tc, ok := ctx.Value(ctxKey{}).(*commonv1.TenantContext)
	return tc, ok && tc != nil
}

// TenantID retorna o tenant_id corrente, ou "" se ausente.
func TenantID(ctx context.Context) string {
	if tc, ok := FromContext(ctx); ok {
		return tc.GetTenantId()
	}
	return ""
}

// UserID retorna o user_id corrente, ou "" se ausente.
func UserID(ctx context.Context) string {
	if tc, ok := FromContext(ctx); ok {
		return tc.GetUserId()
	}
	return ""
}

// Require retorna o TenantContext ou ErrNoTenantContext se ausente/sem tenant_id.
func Require(ctx context.Context) (*commonv1.TenantContext, error) {
	tc, ok := FromContext(ctx)
	if !ok || tc.GetTenantId() == "" {
		return nil, ErrNoTenantContext
	}
	return tc, nil
}

// inject serializa o TenantContext do ctx no Metadata de saída.
func inject(ctx context.Context) context.Context {
	tc, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	raw, err := proto.Marshal(tc)
	if err != nil {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, MetadataKey, string(raw))
}

// extract lê o TenantContext do Metadata de entrada e o anexa ao ctx.
func extract(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	vals := md.Get(MetadataKey)
	if len(vals) == 0 {
		return ctx
	}
	var tc commonv1.TenantContext
	if err := proto.Unmarshal([]byte(vals[0]), &tc); err != nil {
		return ctx
	}
	return NewContext(ctx, &tc)
}

// UnaryClientInterceptor propaga o TenantContext do ctx para o Metadata de saída.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(inject(ctx), method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor propaga o TenantContext em chamadas streaming.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(inject(ctx), desc, cc, method, opts...)
	}
}

// UnaryServerInterceptor extrai o TenantContext do Metadata de entrada para o ctx.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(extract(ctx), req)
	}
}

// StreamServerInterceptor extrai o TenantContext em handlers streaming.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &wrappedStream{ServerStream: ss, ctx: extract(ss.Context())})
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }
