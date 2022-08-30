package cryptoproxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

type Config struct {
	Port             int
	RequestTimeoutMs int
	AuthToken        string
	OriginServer     string
	Paths            []string
}

// TODO: think about request deduplication for same messages
func Start(cfg *Config) error {
	originServerURL, err := url.Parse(cfg.OriginServer)
	if err != nil {
		return fmt.Errorf("invalid origin server URL")
	}

	clients := make(map[string]*Client)
	for _, p := range cfg.Paths {
		clients[p] = NewClient(NewLimiter(1*time.Minute, 10))
		log.Debug().Str("path", p).Int("QPM", 10).Msg("path with rate limit enabled")
	}

	log.Info().Msg("server will operate in memory mode, inflight requests are not stored")
	executor := NewInMemoryExecutor(clients, cfg.RequestTimeoutMs)

	handler := BuildChain(
		executor.Handler,
		NewFilterMiddleware(cfg.Paths),
		NewRequestEnricherMiddleware(originServerURL, cfg.AuthToken),
	)

	log.Info().Int("port", cfg.Port).Msg("server up and running")
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler)
}
