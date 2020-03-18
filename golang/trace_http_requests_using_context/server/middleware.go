package server

import (
	"context"
	"fmt"
	"github.com/bboortz/goborg/pkg/appcontext"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}


// setting the context
func ContextMiddleware(ctx context.Context, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rqId, _ := uuid.NewRandom()
		rqCtx := appcontext.WithRqId(r.Context(), rqId.String())
		r = r.WithContext(rqCtx)

		inner.ServeHTTP(w, r)
	})
}

func LoggerMiddleware(name string, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sw := statusWriter{ResponseWriter: w}
		inner.ServeHTTP(&sw, r)
		logger := appcontext.Logger(r.Context())

		logstr := fmt.Sprintf("HTTP Request %s\t%s\t%s\t%d\t%d\t%s", name, r.Method, r.RequestURI, sw.status, sw.length, time.Since(start))
		logger.Info(logstr)
	})
}
