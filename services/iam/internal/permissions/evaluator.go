// Package permissions avalia, no servidor, o mapa de permissões por componente
// (RN03). A lógica das regras NUNCA sai do IAM Service: o cliente recebe apenas o
// resultado booleano final {view, click} por Blind Index, nunca as condições.
package permissions

// Permissao é uma linha da tabela permissoes já carregada em memória.
type Permissao struct {
	BlindIndex string
	Papel      string
	// Condicao dinâmica; nil/vazia = incondicional. Formato mínimo suportado:
	//   {"atributo": "<nome>", "igual": "<valor>"}
	Condicao map[string]any
	View     bool
	Click    bool
}

// Sujeito é a identidade sob avaliação, derivada do TenantContext/JWT.
type Sujeito struct {
	Papel string
	Tipo  string
	// Attrs alimenta condições dinâmicas (ex.: departamento, região).
	Attrs map[string]string
}

// Decisao é o veredito por componente.
type Decisao struct {
	View  bool
	Click bool
}

// Evaluator não tem estado; é reutilizável e seguro para concorrência.
type Evaluator struct{}

// Avaliar retorna o mapa blind_index → {view, click} para os componentes pedidos.
//
// Política fail-closed (RN03): todo Blind Index solicitado começa negado; uma
// permissão só concede se casar o Papel do sujeito (ou papel curinga "*") e sua
// condição dinâmica for satisfeita. Múltiplas permissões acumulam por OR.
func (Evaluator) Avaliar(suj Sujeito, perms []Permissao, blindIndexes []string) map[string]Decisao {
	out := make(map[string]Decisao, len(blindIndexes))
	pedido := make(map[string]struct{}, len(blindIndexes))
	for _, bi := range blindIndexes {
		out[bi] = Decisao{}
		pedido[bi] = struct{}{}
	}

	for _, p := range perms {
		if _, quer := pedido[p.BlindIndex]; !quer {
			continue
		}
		if p.Papel != suj.Papel && p.Papel != "*" {
			continue
		}
		if !condicaoSatisfeita(p.Condicao, suj) {
			continue
		}
		d := out[p.BlindIndex]
		d.View = d.View || p.View
		d.Click = d.Click || p.Click
		out[p.BlindIndex] = d
	}
	return out
}

// condicaoSatisfeita avalia a condição dinâmica contra o sujeito. Uma condição
// ausente é sempre verdadeira. O único operador suportado nesta fase é a igualdade
// simples; condições malformadas negam (fail-closed).
func condicaoSatisfeita(cond map[string]any, suj Sujeito) bool {
	if len(cond) == 0 {
		return true
	}
	atributo, ok := cond["atributo"].(string)
	if !ok {
		return false
	}
	esperado, ok := cond["igual"].(string)
	if !ok {
		return false
	}
	return valorAtributo(suj, atributo) == esperado
}

func valorAtributo(suj Sujeito, nome string) string {
	switch nome {
	case "tipo":
		return suj.Tipo
	case "papel":
		return suj.Papel
	default:
		return suj.Attrs[nome]
	}
}
