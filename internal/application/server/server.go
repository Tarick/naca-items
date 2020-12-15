package server

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Tarick/naca-items/internal/graph/generated"
	"github.com/Tarick/naca-items/internal/graph/resolver"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

// Server defines HTTP application
type Server struct {
	httpServer *http.Server
	logger     Logger
}

// Config defines webserver configuration
type Config struct {
	Address        string `mapstructure:"address"`
	RequestTimeout int    `mapstructure:"request_timeout"`
}

// New creates new server configuration and configurates middleware
// TODO: move routes to handler file
func New(serverConfig Config, logger Logger, itemsRepository resolver.ItemsRepository) *Server {
	r := chi.NewRouter()
	s := &Server{
		httpServer: &http.Server{Addr: serverConfig.Address, Handler: r},
		logger:     logger,
	}
	r.Use(middleware.RequestID)
	r.Use(middlewareLogger(logger))
	// Basic CORS to allow API calls from browsers (Swagger-UI)
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		// Use this to allow specific origin hosts
		AllowedOrigins: []string{"http://localhost:8080"},
		// AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(serverConfig.RequestTimeout) * time.Second))
	// Healthcheck could be moved back to middleware in case of auth meddling
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("."))
	},
	)
	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.Route("/query", func(r chi.Router) {
		graphqlSchema := generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{ItemsRepository: itemsRepository}})
		graphqlSrv := handler.NewDefaultServer(graphqlSchema)
		// Set 1 second caching and requests coalescing to avoid requests stampede. Beware of any user specific responses.
		// cached := stampede.Handler(512, 1*time.Second)
		// r.With(cached).Get("/", graphsrv)
		r.Handle("/", graphqlSrv)
	})
	return s
}

// StartAndServe configures routers and starts http server
func (s *Server) StartAndServe() {
	s.logger.Info("Server is ready to serve on ", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("Server startup failed: ", err)
	}
}
