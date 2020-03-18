package server

import (
	"context"
	"github.com/bboortz/goborg/pkg/appcontext"
	"net/http"
)

var serverCtx context.Context

type Server struct {
	X, Y float64
}

func (v *Server) ListenAndServe(ctx context.Context, addr string) {
	serverCtx = appcontext.WithPkgName(ctx, "server")
	logger := appcontext.Logger(serverCtx)

	logger.Info("Starting Server and listen on port: " + addr)
	mux := NewRouter(serverCtx)

	server := &http.Server{Addr: addr, Handler: mux}
	logger.Fatal(server.ListenAndServe().Error())
}
