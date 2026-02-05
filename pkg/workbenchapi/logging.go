package workbenchapi

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(payload []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(payload)
	w.bytes += n
	return n, err
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lw, r)

		status := lw.status
		if status == 0 {
			status = http.StatusOK
		}

		event := log.With().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.RawQuery).
			Str("workspace_id", r.URL.Query().Get("workspace_id")).
			Int("status", status).
			Int("bytes", lw.bytes).
			Dur("duration", time.Since(start)).
			Logger()

		switch {
		case status >= 500:
			event.Error().Msg("API request failed")
		case status >= 400:
			event.Warn().Msg("API request client error")
		default:
			event.Info().Msg("API request")
		}
	})
}
