// Package server implementa o ExportEngineService gRPC (RF05): cria o Job e
// responde de imediato, executando a coleta+upload em segundo plano; expõe o
// estado e o Presigned URL; e serve a coleta por Server Streaming (RNF01).
package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	exportv1 "github.com/machv4/platform/gen/go/construtor/export/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/export/internal/collector"
	"github.com/machv4/platform/services/export/internal/jobs"
	"github.com/machv4/platform/services/export/internal/storage"
)

// Jobs, Coletor e Armazenamento são os subconjuntos usados pelo servidor
// (facilitam testes com fakes).
type Jobs interface {
	Criar(ctx context.Context, sistemaID string) (string, error)
	Obter(ctx context.Context, id string) (jobs.Job, error)
	Transicionar(ctx context.Context, id, novo string) error
	MarcarPronto(ctx context.Context, id, url string, expiraEm time.Time) error
}

type Coletor interface {
	Coletar(ctx context.Context, sistemaID string, emit collector.Emit) error
}

// ExportServer implementa exportv1.ExportEngineServiceServer.
type ExportServer struct {
	exportv1.UnimplementedExportEngineServiceServer
	jobs      Jobs
	coletor   Coletor
	storage   storage.Armazenamento
}

// New cria o servidor com as dependências.
func New(j Jobs, c Coletor, s storage.Armazenamento) *ExportServer {
	return &ExportServer{jobs: j, coletor: c, storage: s}
}

// registro é uma linha do NDJSON exportado.
type registro struct {
	Origem   string          `json:"origem"`
	Conteudo json.RawMessage `json:"conteudo"`
}

// CriarJob cria o Job e devolve imediatamente (HTTP 202 no Gateway); a coleta e o
// upload correm em segundo plano (RF05).
func (s *ExportServer) CriarJob(ctx context.Context, req *exportv1.CriarJobRequest) (*exportv1.Job, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	if req.GetSistemaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "sistema_id obrigatório")
	}

	id, err := s.jobs.Criar(ctx, req.GetSistemaId())
	if err != nil {
		return nil, mapErr(err)
	}

	// Contexto próprio para o trabalho de fundo (o da requisição termina já).
	bg := tenantctx.NewContext(context.Background(), tc)
	go s.executar(bg, id, req.GetSistemaId())

	return &exportv1.Job{Id: id, Status: jobs.StatusCriado}, nil
}

// executar corre a coleta em streaming até o upload e conclui o Job com o
// Presigned URL. Coleta e upload são ligados por um io.Pipe: o produtor (coleta)
// aplica backpressure ao consumidor (upload), mantendo a RAM constante (RNF01).
func (s *ExportServer) executar(ctx context.Context, jobID, sistemaID string) {
	if err := s.jobs.Transicionar(ctx, jobID, jobs.StatusColetando); err != nil {
		log.Printf("export %s: transição coletando falhou: %v", jobID, err)
		return
	}

	pr, pw := io.Pipe()
	go func() {
		enc := json.NewEncoder(pw)
		err := s.coletor.Coletar(ctx, sistemaID, func(origem string, conteudo []byte) error {
			return enc.Encode(registro{Origem: origem, Conteudo: conteudo})
		})
		_ = pw.CloseWithError(err) // err nil → EOF limpo para o leitor
	}()

	objeto := "exports/" + sistemaID + "/" + jobID + ".ndjson"
	if err := s.storage.Upload(ctx, objeto, pr); err != nil {
		log.Printf("export %s: upload falhou: %v", jobID, err)
		s.marcarErro(ctx, jobID)
		return
	}

	url, expira, err := s.storage.PresignedGet(ctx, objeto)
	if err != nil {
		log.Printf("export %s: presign falhou: %v", jobID, err)
		s.marcarErro(ctx, jobID)
		return
	}

	if err := s.jobs.MarcarPronto(ctx, jobID, url, expira); err != nil {
		log.Printf("export %s: marcar pronto falhou: %v", jobID, err)
	}
}

func (s *ExportServer) marcarErro(ctx context.Context, jobID string) {
	if err := s.jobs.Transicionar(ctx, jobID, jobs.StatusErro); err != nil {
		log.Printf("export %s: marcar erro falhou: %v", jobID, err)
	}
}

// ObterJob devolve o estado do Job; quando pronto, inclui o Presigned URL (RF05).
func (s *ExportServer) ObterJob(ctx context.Context, req *exportv1.ObterJobRequest) (*exportv1.Job, error) {
	if _, err := tenantctx.Require(ctx); err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	j, err := s.jobs.Obter(ctx, req.GetId())
	if err != nil {
		return nil, mapErr(err)
	}
	out := &exportv1.Job{Id: j.ID, Status: j.Status, ArquivoUrl: j.ArquivoURL}
	if j.ExpiraEm != nil {
		out.ExpiraEm = j.ExpiraEm.UTC().Format(time.RFC3339)
	}
	return out, nil
}

// ColetarDados serve a coleta por Server Streaming em chunks (RNF01). O tenant vem
// do contexto (Metadata), nunca do request.
func (s *ExportServer) ColetarDados(req *exportv1.ColetarDadosRequest, stream exportv1.ExportEngineService_ColetarDadosServer) error {
	if _, err := tenantctx.Require(stream.Context()); err != nil {
		return status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	return s.coletor.Coletar(stream.Context(), req.GetSistemaId(), func(origem string, conteudo []byte) error {
		return stream.Send(&exportv1.ChunkDados{Origem: origem, Conteudo: conteudo})
	})
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, jobs.ErrNaoEncontrado):
		return status.Error(codes.NotFound, "job não encontrado")
	case errors.Is(err, database.ErrNoTenant):
		return status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	default:
		return status.Error(codes.Internal, "erro interno")
	}
}
