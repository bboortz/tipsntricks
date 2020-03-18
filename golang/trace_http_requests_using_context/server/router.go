package server

import (
	"context"
	"fmt"
	"github.com/bboortz/goborg/pkg/appcontext"
	"github.com/gorilla/mux"
	"net/http"
)

func NewRouter(ctx context.Context) *mux.Router {
	router := mux.NewRouter()
	// http.NewServeMux()
	// .StrictSlash(true)
	logger := appcontext.Logger(ctx)

	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = LoggerMiddleware(route.Name, handler)
		handler = ContextMiddleware(ctx, handler)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

		logstr := fmt.Sprintf("Route added %s\t%s\t%s", route.Name, route.Method, route.Pattern)
		logger.Info(logstr)
	}

	return router
}
