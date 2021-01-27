package server

import (
	"net/http"

	// "github.com/99designs/gqlgen-contrib/gqlopentracing"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/Tarick/naca-items/internal/graph/generated"
	"github.com/Tarick/naca-items/internal/graph/resolver"
	opentracing "github.com/opentracing/opentracing-go"
)

// NewHandler creates http handler
func NewHandler(logger Logger, tracer opentracing.Tracer, itemsRepository resolver.ItemsRepository) *Handler {
	graphqlSchema := generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{ItemsRepository: itemsRepository}})
	graphqlSrv := gqlHandler.NewDefaultServer(graphqlSchema)
	// gqlTracer := gqlopentracing.New()
	// graphqlSrv.Use(gqlTracer)
	return &Handler{
		logger:     logger,
		repository: itemsRepository,
		gqlHandler: graphqlSrv,
		tracer:     tracer,
	}
}

// Handler provides http handlers
type Handler struct {
	logger     Logger
	repository resolver.ItemsRepository
	gqlHandler *gqlHandler.Server
	tracer     opentracing.Tracer
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if err := h.repository.Healthcheck(r.Context()); err != nil {
		h.logger.Error("Healthcheck: repository check failed with: ", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Repository is unailable"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("."))
}

// func (h *Handler) setupTracingSpan(r *http.Request, name string) (opentracing.Span, context.Context) {
// 	// we ignore error since if there are missing headers it will start new trace
// 	spanContext, _ := h.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
// 	span := h.tracer.StartSpan(name, opentracing.ChildOf(spanContext))
// 	ctx := opentracing.ContextWithSpan(r.Context(), span)
// 	ext.Component.Set(span, "httpServer-chi")
// 	ext.HTTPMethod.Set(span, r.Method)
// 	ext.HTTPUrl.Set(span, r.URL.String())
// 	return span, ctx
// }
