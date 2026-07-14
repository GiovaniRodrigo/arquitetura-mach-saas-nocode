package server

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	exportv1 "github.com/machv4/platform/gen/go/construtor/export/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/export/internal/collector"
	"github.com/machv4/platform/services/export/internal/jobs"
)

// fakeJobs registra transições para verificar o pipeline.
type fakeJobs struct {
	mu       sync.Mutex
	criarID  string
	criarErr error
	obterJob jobs.Job
	obterErr error
	estados  []string
	pronto   bool
}

func (f *fakeJobs) Criar(context.Context, string) (string, error) { return f.criarID, f.criarErr }
func (f *fakeJobs) Obter(context.Context, string) (jobs.Job, error) {
	return f.obterJob, f.obterErr
}
func (f *fakeJobs) Transicionar(_ context.Context, _, novo string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.estados = append(f.estados, novo)
	return nil
}
func (f *fakeJobs) MarcarPronto(_ context.Context, _, _ string, _ time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pronto = true
	return nil
}
func (f *fakeJobs) foiPronto() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.pronto
}

type fakeColetor struct{}

func (fakeColetor) Coletar(_ context.Context, _ string, emit collector.Emit) error {
	return emit(collector.OrigemDesign, []byte(`{"id":"d1"}`))
}

type fakeStorage struct {
	mu       sync.Mutex
	uploaded []byte
}

func (s *fakeStorage) Upload(_ context.Context, _ string, r io.Reader) error {
	b, err := io.ReadAll(r)
	s.mu.Lock()
	s.uploaded = b
	s.mu.Unlock()
	return err
}
func (s *fakeStorage) PresignedGet(context.Context, string) (string, time.Time, error) {
	return "https://signed/exports/x?X-Amz-Expires=600", time.Now().Add(10 * time.Minute), nil
}

func comTenant() context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "t-1"})
}

func code(err error) codes.Code { return status.Code(err) }

func TestCriarJob_SemTenant(t *testing.T) {
	s := New(&fakeJobs{}, fakeColetor{}, &fakeStorage{})
	_, err := s.CriarJob(context.Background(), &exportv1.CriarJobRequest{SistemaId: "s1"})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}

func TestCriarJob_RespondeImediatoEExecutaPipeline(t *testing.T) {
	fj := &fakeJobs{criarID: "j-1"}
	fs := &fakeStorage{}
	s := New(fj, fakeColetor{}, fs)

	out, err := s.CriarJob(comTenant(), &exportv1.CriarJobRequest{SistemaId: "s1"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if out.GetId() != "j-1" || out.GetStatus() != jobs.StatusCriado {
		t.Fatalf("resposta imediata inesperada: %+v", out)
	}

	// O pipeline corre em segundo plano: aguardamos concluir (pronto).
	esperarAte(t, fj.foiPronto, time.Second)

	fs.mu.Lock()
	uploaded := string(fs.uploaded)
	fs.mu.Unlock()
	if uploaded == "" || uploaded[0] != '{' {
		t.Fatalf("upload NDJSON inesperado: %q", uploaded)
	}
}

func TestObterJob_Pronto_IncluiURL(t *testing.T) {
	expira := time.Now().Add(10 * time.Minute)
	fj := &fakeJobs{obterJob: jobs.Job{ID: "j-1", Status: jobs.StatusPronto, ArquivoURL: "https://x", ExpiraEm: &expira}}
	s := New(fj, fakeColetor{}, &fakeStorage{})

	out, err := s.ObterJob(comTenant(), &exportv1.ObterJobRequest{Id: "j-1"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if out.GetArquivoUrl() != "https://x" || out.GetExpiraEm() == "" {
		t.Fatalf("resposta inesperada: %+v", out)
	}
}

func TestObterJob_NaoEncontrado(t *testing.T) {
	s := New(&fakeJobs{obterErr: jobs.ErrNaoEncontrado}, fakeColetor{}, &fakeStorage{})
	_, err := s.ObterJob(comTenant(), &exportv1.ObterJobRequest{Id: "x"})
	if code(err) != codes.NotFound {
		t.Fatalf("esperava NotFound; got %v", err)
	}
}

func esperarAte(t *testing.T, cond func() bool, prazo time.Duration) {
	t.Helper()
	fim := time.Now().Add(prazo)
	for time.Now().Before(fim) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condição não satisfeita no prazo")
}
