// Package routes — login social (third-party) do Gateway.
//
// O Gateway conduz o protocolo OAuth2 com o provedor (Google/GitHub), obtém o
// perfil e delega ao IAM (AutenticarThirdParty) a materialização da identidade e
// a emissão do JWT MACH. O token volta ao SPA por fragmento de URL num destino
// da allowlist (defesa contra open redirect / vazamento de token). Padrão portado
// do serviço /var/www/autenticacao-app (oauth.go, oauth_redirect.go).
package routes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
)

// OAuthAutenticador é o subconjunto do IAM usado no login social.
type OAuthAutenticador interface {
	AutenticarThirdParty(ctx context.Context, in *iamv1.AutenticarThirdPartyRequest, opts ...grpc.CallOption) (*iamv1.AutenticarThirdPartyResponse, error)
}

// OAuthHandler serve os endpoints /auth/{provedor} e /auth/{provedor}/callback.
type OAuthHandler struct {
	iam      OAuthAutenticador
	configs  map[string]*oauth2.Config
	allowed  []string // origens (scheme://host[:port]) permitidas no redirect final
	fallback string   // destino padrão quando não há redirect_uri
}

// NewOAuthHandler monta o handler a partir da configuração OAuth. Devolve nil se
// nenhum provedor estiver configurado (feature desligada) — o router então não
// registra as rotas de login.
func NewOAuthHandler(iam OAuthAutenticador, appURL, googleID, googleSecret, githubID, githubSecret, allowedCSV string) *OAuthHandler {
	appURL = strings.TrimRight(appURL, "/")
	configs := map[string]*oauth2.Config{}
	if googleID != "" && googleSecret != "" {
		configs["google"] = &oauth2.Config{
			ClientID:     googleID,
			ClientSecret: googleSecret,
			RedirectURL:  appURL + "/auth/google/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		}
	}
	if githubID != "" && githubSecret != "" {
		configs["github"] = &oauth2.Config{
			ClientID:     githubID,
			ClientSecret: githubSecret,
			RedirectURL:  appURL + "/auth/github/callback",
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		}
	}
	if len(configs) == 0 {
		return nil
	}
	var allowed []string
	for _, o := range strings.Split(allowedCSV, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowed = append(allowed, strings.TrimRight(o, "/"))
		}
	}
	return &OAuthHandler{iam: iam, configs: configs, allowed: allowed, fallback: appURL + "/app"}
}

// Registra as rotas de login (fora do grupo autenticado).
func (h *OAuthHandler) Registrar(r chi.Router) {
	r.Get("/auth/{provedor}", h.Login)
	r.Get("/auth/{provedor}/callback", h.Callback)
}

// Login inicia o fluxo: redireciona ao provedor com o state carregando o
// redirect_uri (já validado contra a allowlist).
func (h *OAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	cfg, ok := h.configs[chi.URLParam(r, "provedor")]
	if !ok {
		http.Error(w, "provedor não suportado", http.StatusNotFound)
		return
	}
	redirectURI := h.validarRedirect(r.URL.Query().Get("redirect_uri"))
	http.Redirect(w, r, cfg.AuthCodeURL(codificarState(redirectURI)), http.StatusTemporaryRedirect)
}

// Callback troca o code por perfil, materializa a identidade no IAM e devolve o
// JWT ao destino (fragmento #token=), sempre um alvo da allowlist.
func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	prov := chi.URLParam(r, "provedor")
	cfg, ok := h.configs[prov]
	if !ok {
		http.Error(w, "provedor não suportado", http.StatusNotFound)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code OAuth ausente", http.StatusBadRequest)
		return
	}
	user, err := h.perfil(r.Context(), prov, cfg, code)
	if err != nil {
		log.Printf("oauth %s: %v", prov, err)
		http.Error(w, "falha na autenticação", http.StatusBadGateway)
		return
	}
	resp, err := h.iam.AutenticarThirdParty(r.Context(), &iamv1.AutenticarThirdPartyRequest{
		Provedor:   prov,
		ExternalId: user.ID,
		Email:      user.Email,
		Nome:       user.Name,
	})
	if err != nil {
		log.Printf("iam autenticar third-party: %v", err)
		http.Error(w, "falha ao emitir sessão", http.StatusBadGateway)
		return
	}
	dest := h.destino(r.URL.Query().Get("state"))
	http.Redirect(w, r, dest+"#token="+url.QueryEscape(resp.GetJwt()), http.StatusTemporaryRedirect)
}

type socialUser struct {
	ID    string
	Email string
	Name  string
}

// perfil troca o code e busca o perfil no provedor.
func (h *OAuthHandler) perfil(ctx context.Context, prov string, cfg *oauth2.Config, code string) (*socialUser, error) {
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("troca de code: %w", err)
	}
	client := cfg.Client(ctx, tok)
	switch prov {
	case "google":
		var g struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := getJSON(client, "https://www.googleapis.com/oauth2/v2/userinfo", &g); err != nil {
			return nil, err
		}
		return &socialUser{ID: g.ID, Email: g.Email, Name: g.Name}, nil
	case "github":
		var p struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := getJSON(client, "https://api.github.com/user", &p); err != nil {
			return nil, err
		}
		email := p.Email
		if email == "" {
			email = emailPrimarioGitHub(client)
		}
		nome := p.Name
		if nome == "" {
			nome = p.Login
		}
		return &socialUser{ID: fmt.Sprintf("%d", p.ID), Email: email, Name: nome}, nil
	default:
		return nil, fmt.Errorf("provedor não suportado: %s", prov)
	}
}

// emailPrimarioGitHub busca o e-mail primário verificado quando o perfil não o expõe.
func emailPrimarioGitHub(client *http.Client) string {
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := getJSON(client, "https://api.github.com/user/emails", &emails); err != nil {
		return ""
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	return ""
}

func getJSON(client *http.Client, u string, dst any) error {
	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: status %d", u, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}

// validarRedirect devolve o redirect_uri só se a origem estiver na allowlist.
func (h *OAuthHandler) validarRedirect(redirectURI string) string {
	if redirectURI == "" || len(h.allowed) == 0 {
		return ""
	}
	parsed, err := url.Parse(redirectURI)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	origem := parsed.Scheme + "://" + parsed.Host
	for _, permitida := range h.allowed {
		if origem == permitida {
			return redirectURI
		}
	}
	return ""
}

// destino resolve o alvo final a partir do state (revalidando a allowlist, pois
// o valor volta do provedor). Sem redirect_uri válido, cai no fallback (APP_URL/app).
func (h *OAuthHandler) destino(state string) string {
	if redirectURI := decodificarState(state); redirectURI != "" {
		if v := h.validarRedirect(redirectURI); v != "" {
			return v
		}
	}
	return h.fallback
}

// codificarState empacota o redirect_uri (já validado) no parâmetro state, que os
// provedores devolvem inalterado no callback.
func codificarState(redirectURI string) string {
	if redirectURI == "" {
		return "mach-state"
	}
	return "r_" + base64.URLEncoding.EncodeToString([]byte(redirectURI))
}

func decodificarState(state string) string {
	if !strings.HasPrefix(state, "r_") {
		return ""
	}
	decoded, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(state, "r_"))
	if err != nil {
		return ""
	}
	return string(decoded)
}
