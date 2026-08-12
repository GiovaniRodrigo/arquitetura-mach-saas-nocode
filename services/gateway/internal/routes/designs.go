package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// DesignCliente é o subconjunto do DesignEngineServiceClient usado pelas rotas.
type DesignCliente interface {
	CriarDesign(ctx context.Context, in *designv1.Design, opts ...grpc.CallOption) (*designv1.Design, error)
	ObterDesign(ctx context.Context, in *designv1.ObterDesignRequest, opts ...grpc.CallOption) (*designv1.Design, error)
	AtualizarDesign(ctx context.Context, in *designv1.Design, opts ...grpc.CallOption) (*designv1.Design, error)
	RemoverDesign(ctx context.Context, in *designv1.ObterDesignRequest, opts ...grpc.CallOption) (*designv1.RemoverResponse, error)
	ListarDesigns(ctx context.Context, in *designv1.ListarDesignsRequest, opts ...grpc.CallOption) (*designv1.ListarDesignsResponse, error)
	CriarSistema(ctx context.Context, in *designv1.CriarSistemaRequest, opts ...grpc.CallOption) (*designv1.Sistema, error)
	ListarSistemas(ctx context.Context, in *designv1.ListarSistemasRequest, opts ...grpc.CallOption) (*designv1.ListarSistemasResponse, error)
	AtualizarWhiteLabel(ctx context.Context, in *designv1.AtualizarWhiteLabelRequest, opts ...grpc.CallOption) (*designv1.WhiteLabel, error)
}

// componenteDTO espelha o nó Composite no corpo JSON. `propriedades` trafega como
// objeto JSON aninhado (RawMessage), convertido para os bytes do proto.
type componenteDTO struct {
	BlindIndex   string          `json:"blind_index"`
	Tipo         string          `json:"tipo"`
	Propriedades json.RawMessage `json:"propriedades,omitempty"`
	Filhos       []componenteDTO `json:"componente_filhos,omitempty"`
}

func (c *componenteDTO) toProto() *designv1.Componente {
	if c == nil {
		return nil
	}
	p := &designv1.Componente{BlindIndex: c.BlindIndex, Tipo: c.Tipo}
	if len(c.Propriedades) > 0 {
		p.Propriedades = []byte(c.Propriedades)
	}
	for i := range c.Filhos {
		p.ComponenteFilhos = append(p.ComponenteFilhos, c.Filhos[i].toProto())
	}
	return p
}

func dtoFromProto(c *designv1.Componente) componenteDTO {
	d := componenteDTO{BlindIndex: c.GetBlindIndex(), Tipo: c.GetTipo()}
	if p := c.GetPropriedades(); len(p) > 0 {
		d.Propriedades = json.RawMessage(p)
	}
	for _, f := range c.GetComponenteFilhos() {
		d.Filhos = append(d.Filhos, dtoFromProto(f))
	}
	return d
}

type designReq struct {
	SistemaID string        `json:"sistema_id"`
	Nome      string        `json:"nome"`
	Arvore    componenteDTO `json:"arvore"`
}

type designResp struct {
	ID        string        `json:"id"`
	SistemaID string        `json:"sistema_id"`
	Nome      string        `json:"nome"`
	Arvore    componenteDTO `json:"arvore"`
}

func respFromProto(d *designv1.Design) designResp {
	return designResp{
		ID:        d.GetId(),
		SistemaID: d.GetSistemaId(),
		Nome:      d.GetNome(),
		Arvore:    dtoFromProto(d.GetArvore()),
	}
}

// CriarDesign serve POST /api/v1/designs (RF01). O tenant vem do TenantContext no
// contexto (posto pelo Auth) e é propagado ao Design Engine via Metadata gRPC.
func CriarDesign(design DesignCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body designReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if body.SistemaID == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "sistema_id obrigatório")
			return
		}
		out, err := design.CriarDesign(r.Context(), &designv1.Design{
			SistemaId: body.SistemaID,
			Nome:      body.Nome,
			Arvore:    body.Arvore.toProto(),
		})
		if err != nil {
			writeDesignError(w, err)
			return
		}
		web.JSON(w, http.StatusCreated, map[string]string{
			"design_id":  out.GetId(),
			"sistema_id": out.GetSistemaId(),
		})
	}
}

// telaResumoResp espelha um DesignResumo no corpo JSON da listagem.
type telaResumoResp struct {
	ID   string `json:"id"`
	Nome string `json:"nome"`
}

// ListarDesigns serve GET /api/v1/designs?sistema_id={id} (RF09): lista as
// telas de um sistema, forma leve sem a árvore.
func ListarDesigns(design DesignCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sistemaID := r.URL.Query().Get("sistema_id")
		if sistemaID == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "sistema_id obrigatório")
			return
		}
		out, err := design.ListarDesigns(r.Context(), &designv1.ListarDesignsRequest{SistemaId: sistemaID})
		if err != nil {
			writeDesignError(w, err)
			return
		}
		lista := make([]telaResumoResp, 0, len(out.GetDesigns()))
		for _, d := range out.GetDesigns() {
			lista = append(lista, telaResumoResp{ID: d.GetId(), Nome: d.GetNome()})
		}
		web.JSON(w, http.StatusOK, map[string]any{"telas": lista})
	}
}

// ObterDesign serve GET /api/v1/designs/{id} (RF01).
func ObterDesign(design DesignCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := design.ObterDesign(r.Context(), &designv1.ObterDesignRequest{Id: chi.URLParam(r, "id")})
		if err != nil {
			writeDesignError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, respFromProto(out))
	}
}

// AtualizarDesign serve PUT /api/v1/designs/{id} (RF01).
func AtualizarDesign(design DesignCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body designReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		out, err := design.AtualizarDesign(r.Context(), &designv1.Design{
			Id:        chi.URLParam(r, "id"),
			SistemaId: body.SistemaID,
			Nome:      body.Nome,
			Arvore:    body.Arvore.toProto(),
		})
		if err != nil {
			writeDesignError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, respFromProto(out))
	}
}

// RemoverDesign serve DELETE /api/v1/designs/{id} (RF01).
func RemoverDesign(design DesignCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := design.RemoverDesign(r.Context(), &designv1.ObterDesignRequest{Id: chi.URLParam(r, "id")}); err != nil {
			writeDesignError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// writeDesignError traduz os códigos gRPC do Design Engine para HTTP conforme
// api.md — em particular InvalidArgument → 422 INVALID_TREE (árvore inválida).
func writeDesignError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusUnprocessableEntity, "INVALID_TREE", "árvore recursiva estruturalmente inválida")
	case codes.NotFound:
		web.Error(w, http.StatusNotFound, "NOT_FOUND", "design não encontrado")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
