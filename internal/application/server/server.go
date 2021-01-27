package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	// "github.com/99designs/gqlgen/graphql/playground"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Server defines HTTP application
type Server struct {
	httpServer *http.Server
	handler    *Handler
	logger     Logger
}

// Config defines webserver configuration
type Config struct {
	Address        string `mapstructure:"address"`
	RequestTimeout int    `mapstructure:"request_timeout"`
}

// New creates new server configuration and configurates middleware
// TODO: move routes to handler file
func New(serverConfig Config, logger Logger, handler *Handler) *Server {
	r := chi.NewRouter()
	s := &Server{
		httpServer: &http.Server{Addr: serverConfig.Address, Handler: r},
		logger:     logger,
		handler:    handler,
	}
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(time.Duration(serverConfig.RequestTimeout) * time.Second))
		// Prometheus metrics
		r.Handle("/metrics", promhttp.Handler())
		r.Get("/healthz", http.HandlerFunc(handler.healthCheck))
	})
	r.Group(func(r chi.Router) {
		// r.Handle("/", playground.Handler("GraphQL playground", "/query"))
		r.Use(middleware.RequestID)
		r.Use(middlewareLogger(logger))
		r.Route("/query", func(r chi.Router) {
			// Set 1 second caching and requests coalescing to avoid requests stampede. Beware of any user specific responses.
			// cached := stampede.Handler(512, 1*time.Second)
			// r.With(cached).Get("/", gqlHandler)
			r.Handle("/", handler.gqlHandler)

		})

	})
	return s
}

// StartAndServe configures routers and starts http server
func (s *Server) StartAndServe() error {
	s.logger.Info("Server is ready to serve on ", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
