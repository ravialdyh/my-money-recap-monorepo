package serve

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Env struct{ err error }

func (e *Env) Get(key, def string) string {
	if v := e.fromFile(key); v != "" {
		return v
	}
	if v := os.Getenv(key); v != "" {
		return v
	}
	if def == "" && e.err == nil {
		e.err = fmt.Errorf("%s is required", key)
	}
	return def
}

func (e *Env) Optional(key string) string {
	if v := e.fromFile(key); v != "" {
		return v
	}
	return os.Getenv(key)
}

func (e *Env) fromFile(key string) string {
	path := os.Getenv(key + "_FILE")
	if path == "" {
		return ""
	}
	body, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if e.err == nil {
			e.err = err
		}
		return ""
	}
	return strings.TrimSpace(string(body))
}

func (e *Env) Err() error { return e.err }

func Run(ctx context.Context, name, addr string, mux *http.ServeMux, ready func(context.Context) error) error {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if ready != nil {
			if err := ready(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
		fmt.Fprintln(w, "ok")
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("%s listening on %s", name, addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-drained
	return nil
}
