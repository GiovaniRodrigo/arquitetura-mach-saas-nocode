package telemetry

import (
	"regexp"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

// DefaultMask substitui termos sensíveis redigidos.
const DefaultMask = "[REDACTED]"

// Redactor mascara termos sensíveis (nomes reais de tabelas/campos) que não devem
// aparecer em logs, mensagens de erro ou atributos de span (RNF08). A comparação é
// case-insensitive. Um Redactor é imutável e seguro para uso concorrente.
type Redactor struct {
	re   *regexp.Regexp
	mask string
}

// NewRedactor compila um Redactor para os termos informados. Termos vazios são
// ignorados; sem termos, Redact é uma no-op.
func NewRedactor(terms ...string) *Redactor {
	quoted := make([]string, 0, len(terms))
	for _, t := range terms {
		if t == "" {
			continue
		}
		quoted = append(quoted, regexp.QuoteMeta(t))
	}
	r := &Redactor{mask: DefaultMask}
	if len(quoted) > 0 {
		r.re = regexp.MustCompile("(?i)" + strings.Join(quoted, "|"))
	}
	return r
}

// WithMask devolve uma cópia do Redactor usando outra máscara.
func (r *Redactor) WithMask(mask string) *Redactor {
	return &Redactor{re: r.re, mask: mask}
}

// Redact substitui todas as ocorrências dos termos sensíveis pela máscara.
func (r *Redactor) Redact(s string) string {
	if r.re == nil {
		return s
	}
	return r.re.ReplaceAllString(s, r.mask)
}

// RedactAttrs devolve os atributos com valores string redigidos. Chaves e valores
// não-string permanecem intactos (um Blind Index, por não conter nome real, passa
// incólume).
func (r *Redactor) RedactAttrs(attrs []attribute.KeyValue) []attribute.KeyValue {
	if r.re == nil {
		return attrs
	}
	out := make([]attribute.KeyValue, len(attrs))
	for i, a := range attrs {
		if a.Value.Type() == attribute.STRING {
			out[i] = a.Key.String(r.Redact(a.Value.AsString()))
			continue
		}
		out[i] = a
	}
	return out
}
