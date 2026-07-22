package server

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/logic/internal/events"
	"github.com/machv4/platform/services/logic/internal/store"
	"github.com/machv4/platform/services/logic/internal/validation"
)

type fakeStore struct {
	schema     map[string]validation.CampoDef
	schemaErr  error
	inserido   map[string]string
	inserirErr error
	criarID    string
	regras     []store.Regra
	regrasErr  error
}

func (f *fakeStore) SchemaDe(context.Context, string) (map[string]validation.CampoDef, error) {
	return f.schema, f.schemaErr
}
func (f *fakeStore) InserirDados(_ context.Context, _ string, v map[string]string) (string, error) {
	f.inserido = v
	return "d-1", f.inserirErr
}
func (f *fakeStore) RegrasDoSistema(context.Context, string) ([]store.Regra, error) {
	return f.regras, f.regrasErr
}
func (f *fakeStore) CriarRegra(context.Context, store.Regra) (string, error) { return f.criarID, nil }
func (f *fakeStore) ObterRegra(context.Context, string) (store.Regra, error) {
	return store.Regra{}, store.ErrNaoEncontrado
}
func (f *fakeStore) AtualizarRegra(context.Context, store.Regra) error { return nil }
func (f *fakeStore) RemoverRegra(context.Context, string) error        { return nil }

// fakePublicador grava os eventos publicados durante o disparo de regras (RF08).
type fakePublicador struct {
	publicados []events.Evento
	err        error
}

func (f *fakePublicador) Publicar(_ context.Context, ev events.Evento) error {
	f.publicados = append(f.publicados, ev)
	return f.err
}

func comTenant() context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "t-1", Tipo: "dono"})
}

func code(err error) codes.Code { return status.Code(err) }

func schemaObrigatorio() map[string]validation.CampoDef {
	return map[string]validation.CampoDef{
		"bi-nome": {BlindIndex: "bi-nome", Tipo: validation.TipoString, Obrigatorio: true},
	}
}

func TestSalvarFormulario_SemTenant_Unauthenticated(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.SalvarFormulario(context.Background(), &logicv1.SalvarFormularioRequest{SistemaId: "s1"})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}

func TestSalvarFormulario_Invalido_NaoPersiste(t *testing.T) {
	fs := &fakeStore{schema: schemaObrigatorio()}
	s := New(fs)
	// Falta o campo obrigatório bi-nome e injeta chave fora do schema.
	resp, err := s.SalvarFormulario(comTenant(), &logicv1.SalvarFormularioRequest{
		SistemaId:       "s1",
		DadosFormulario: map[string]string{"bi-x": "malicioso"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resp.GetSucesso() {
		t.Fatal("payload inválido não deveria ter sucesso")
	}
	if _, ok := resp.GetErrosValidacao()["bi-nome"]; !ok {
		t.Fatalf("esperava erro em bi-nome; got %v", resp.GetErrosValidacao())
	}
	if _, ok := resp.GetErrosValidacao()["bi-x"]; !ok {
		t.Fatalf("esperava rejeição da chave desconhecida bi-x; got %v", resp.GetErrosValidacao())
	}
	if fs.inserido != nil {
		t.Fatal("não deveria persistir com erros de validação")
	}
}

func TestSalvarFormulario_Valido_Persiste(t *testing.T) {
	fs := &fakeStore{schema: schemaObrigatorio()}
	s := New(fs)
	resp, err := s.SalvarFormulario(comTenant(), &logicv1.SalvarFormularioRequest{
		SistemaId:       "s1",
		DadosFormulario: map[string]string{"bi-nome": "Ana"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !resp.GetSucesso() || fs.inserido["bi-nome"] != "Ana" {
		t.Fatalf("deveria persistir; resp=%v inserido=%v", resp, fs.inserido)
	}
}

func TestSalvarFormulario_RegraDispara_PublicaEvento(t *testing.T) {
	arvore := []byte(`{"no":"condicao","operador":"igual","blind_index":"bi-nome","valor":"Ana",
		"entao":{"no":"acao","tipo":"webhook.disparo","config":{"url_destino":"https://x"}}}`)
	fs := &fakeStore{
		schema: schemaObrigatorio(),
		regras: []store.Regra{{ID: "r-1", SistemaID: "s1", ArvoreDecisao: arvore}},
	}
	pub := &fakePublicador{}
	s := New(fs).ComPublicador(pub)

	resp, err := s.SalvarFormulario(comTenant(), &logicv1.SalvarFormularioRequest{
		SistemaId:       "s1",
		DadosFormulario: map[string]string{"bi-nome": "Ana"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !resp.GetSucesso() {
		t.Fatalf("deveria persistir; resp=%v", resp)
	}
	if len(pub.publicados) != 1 || pub.publicados[0].Tipo != "webhook.disparo" {
		t.Fatalf("esperava 1 evento webhook.disparo publicado; got %+v", pub.publicados)
	}
}

func TestSalvarFormulario_RegraNaoAtingeAcao_NaoPublica(t *testing.T) {
	arvore := []byte(`{"no":"condicao","operador":"igual","blind_index":"bi-nome","valor":"outro-nome",
		"entao":{"no":"acao","tipo":"webhook.disparo"}}`)
	fs := &fakeStore{
		schema: schemaObrigatorio(),
		regras: []store.Regra{{ID: "r-1", SistemaID: "s1", ArvoreDecisao: arvore}},
	}
	pub := &fakePublicador{}
	s := New(fs).ComPublicador(pub)

	if _, err := s.SalvarFormulario(comTenant(), &logicv1.SalvarFormularioRequest{
		SistemaId:       "s1",
		DadosFormulario: map[string]string{"bi-nome": "Ana"},
	}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(pub.publicados) != 0 {
		t.Fatalf("condição não deveria disparar ação; got %+v", pub.publicados)
	}
}

func TestSalvarFormulario_SemPublicador_ContinuaFuncionando(t *testing.T) {
	fs := &fakeStore{schema: schemaObrigatorio()}
	s := New(fs) // sem ComPublicador
	resp, err := s.SalvarFormulario(comTenant(), &logicv1.SalvarFormularioRequest{
		SistemaId:       "s1",
		DadosFormulario: map[string]string{"bi-nome": "Ana"},
	})
	if err != nil || !resp.GetSucesso() {
		t.Fatalf("deveria persistir mesmo sem publicador; resp=%v err=%v", resp, err)
	}
}

func TestSalvarFormulario_ErroAoPublicar_NaoFalhaResposta(t *testing.T) {
	arvore := []byte(`{"no":"acao","tipo":"webhook.disparo"}`)
	fs := &fakeStore{
		schema: schemaObrigatorio(),
		regras: []store.Regra{{ID: "r-1", SistemaID: "s1", ArvoreDecisao: arvore}},
	}
	pub := &fakePublicador{err: errors.New("rabbitmq indisponível")}
	s := New(fs).ComPublicador(pub)

	resp, err := s.SalvarFormulario(comTenant(), &logicv1.SalvarFormularioRequest{
		SistemaId:       "s1",
		DadosFormulario: map[string]string{"bi-nome": "Ana"},
	})
	if err != nil || !resp.GetSucesso() {
		t.Fatalf("falha ao publicar não deveria afetar a resposta; resp=%v err=%v", resp, err)
	}
}

func TestCriarRegra_ArvoreInvalida_InvalidArgument(t *testing.T) {
	s := New(&fakeStore{criarID: "r-1"})
	_, err := s.CriarRegra(comTenant(), &logicv1.Regra{SistemaId: "s1", ArvoreDecisao: []byte(`{nao-json`)})
	if code(err) != codes.InvalidArgument {
		t.Fatalf("esperava InvalidArgument; got %v", err)
	}
}

func TestCriarRegra_OK(t *testing.T) {
	s := New(&fakeStore{criarID: "r-1"})
	arvore := []byte(`{"no":"acao","tipo":"webhook.disparo"}`)
	out, err := s.CriarRegra(comTenant(), &logicv1.Regra{SistemaId: "s1", ArvoreDecisao: arvore})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.GetId() != "r-1" {
		t.Fatalf("id inesperado: %+v", out)
	}
}

func TestObterRegra_NaoEncontrada_NotFound(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.ObterRegra(comTenant(), &logicv1.ObterRegraRequest{Id: "x"})
	if code(err) != codes.NotFound {
		t.Fatalf("esperava NotFound; got %v", err)
	}
}
