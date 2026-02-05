package workbenchapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type Config struct {
	Addr     string
	BasePath string
}

type Server struct {
	cfg        Config
	apiMux     *http.ServeMux
	rootMux    *http.ServeMux
	httpServer *http.Server
}

func NewServer(cfg Config) *Server {
	if strings.TrimSpace(cfg.Addr) == "" {
		cfg.Addr = ":8080"
	}
	if strings.TrimSpace(cfg.BasePath) == "" {
		cfg.BasePath = "/api"
	}

	apiMux := http.NewServeMux()
	registerBaseRoutes(apiMux)

	rootMux := http.NewServeMux()
	basePath := normalizeBasePath(cfg.BasePath)
	if basePath == "/" {
		rootMux = apiMux
	} else {
		rootMux.Handle(basePath+"/", http.StripPrefix(basePath, apiMux))
	}

	return &Server{
		cfg:     cfg,
		apiMux:  apiMux,
		rootMux: rootMux,
	}
}

func (s *Server) ListenAndServe() error {
	s.httpServer = &http.Server{
		Addr:    s.cfg.Addr,
		Handler: s.rootMux,
	}

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "serve workbench api")
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return errors.Wrap(s.httpServer.Shutdown(ctx), "shutdown workbench api")
}

func normalizeBasePath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}
