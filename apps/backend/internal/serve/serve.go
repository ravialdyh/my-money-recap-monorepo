// Package serve is what every binary shares: required-or-default environment, and one HTTP server
// with /healthz and graceful shutdown (ARCHITECTURE §10: each Go container listens on 8080).
package serve

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Env is the variable's value, or def; an empty def means the variable is required.
func Env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if def == "" {
		log.Fatalf("%s is required", key)
	}
	return def
}

// Run adds /healthz to mux (ready, when set, must succeed for "ok"; members pings its database) and
// serves until ctx is done, then drains for five seconds. It returns only on a listen error.
func Run(ctx context.Context, name, addr string, mux *http.ServeMux, ready func(context.Context) error) {
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
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("%s listening on %s", name, addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
