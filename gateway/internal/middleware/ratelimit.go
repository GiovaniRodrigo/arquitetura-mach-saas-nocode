package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"github.com/machv4/platform/gateway/internal/web"
	"github.com/machv4/platform/pkg/tenantctx"
)

// RateLimiter aplica um token bucket por tenant (RNF04/isolamento de vizinho
// barulhento). Deve ser encadeado DEPOIS do Auth, para haver tenant no contexto.
type RateLimiter struct {
	rps      rate.Limit
	burst    int
	mu       sync.Mutex
	buckets  map[string]*rate.Limiter
}

// NewRateLimiter cria um limitador com taxa (req/s) e burst por tenant.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		rps:     rate.Limit(rps),
		burst:   burst,
		buckets: make(map[string]*rate.Limiter),
	}
}

func (rl *RateLimiter) limiter(tenantID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	l, ok := rl.buckets[tenantID]
	if !ok {
		l = rate.NewLimiter(rl.rps, rl.burst)
		rl.buckets[tenantID] = l
	}
	return l
}

// Handler é o middleware.
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := tenantctx.TenantID(r.Context())
		if tenantID == "" {
			// Sem tenant não há cota por tenant — deixa o Auth já ter barrado.
			next.ServeHTTP(w, r)
			return
		}
		if !rl.limiter(tenantID).Allow() {
			web.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "limite de requisições excedido")
			return
		}
		next.ServeHTTP(w, r)
	})
}
