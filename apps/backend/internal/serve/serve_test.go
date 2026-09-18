package serve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	return addr
}

func TestEnvPrefersTheValueThenTheDefault(t *testing.T) {
	t.Setenv("MMR_TEST_SET", "from-env")
	var env Env
	if got := env.Get("MMR_TEST_SET", "fallback"); got != "from-env" {
		t.Errorf("Get = %q, want from-env", got)
	}
	if got := env.Get("MMR_TEST_UNSET", "fallback"); got != "fallback" {
		t.Errorf("Get = %q, want fallback", got)
	}
	if err := env.Err(); err != nil {
		t.Errorf("Err = %v, want nil", err)
	}
}

func TestEnvAnEmptyValueFallsBackToTheDefault(t *testing.T) {
	t.Setenv("MMR_TEST_EMPTY", "")
	var env Env
	if got := env.Get("MMR_TEST_EMPTY", "fallback"); got != "fallback" {
		t.Errorf("Get = %q, want fallback", got)
	}
	if err := env.Err(); err != nil {
		t.Errorf("Err = %v, want nil", err)
	}
}

func TestEnvReportsTheFirstMissingRequiredVariable(t *testing.T) {
	os.Unsetenv("MMR_TEST_FIRST")
	os.Unsetenv("MMR_TEST_SECOND")
	var env Env
	env.Get("MMR_TEST_FIRST", "")
	env.Get("MMR_TEST_SECOND", "")
	err := env.Err()
	if err == nil {
		t.Fatal("Err = nil, want the first missing variable")
	}
	if !strings.Contains(err.Error(), "MMR_TEST_FIRST") {
		t.Errorf("Err = %v, want it to name MMR_TEST_FIRST", err)
	}
	if strings.Contains(err.Error(), "MMR_TEST_SECOND") {
		t.Errorf("Err = %v, want only the first missing variable", err)
	}
}

func TestRunAnswersHealthzAndReturnsNilOnShutdown(t *testing.T) {
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Run(ctx, "test", addr, http.NewServeMux(), nil) }()

	body, code := getHealthz(t, addr)
	if code != http.StatusOK || strings.TrimSpace(body) != "ok" {
		t.Errorf("/healthz = %d %q, want 200 ok", code, body)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run = %v, want nil after shutdown", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return after the context was cancelled")
	}
}

func TestRunDrainsAnInFlightRequestBeforeReturning(t *testing.T) {
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started, release := make(chan struct{}), make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		fmt.Fprint(w, "done")
	})
	done := make(chan error, 1)
	go func() { done <- Run(ctx, "test", addr, mux, nil) }()
	getHealthz(t, addr)

	answered := make(chan string, 1)
	go func() {
		res, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			answered <- err.Error()
			return
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		answered <- string(body)
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		t.Fatalf("Run returned %v while a request was still in flight", err)
	case <-time.After(300 * time.Millisecond):
	}
	close(release)
	select {
	case body := <-answered:
		if body != "done" {
			t.Errorf("in-flight request got %q, want its response", body)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the in-flight request never got its response")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run = %v, want nil after draining", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return after the in-flight request finished")
	}
}

func TestRunHealthzIsUnavailableWhenReadyFails(t *testing.T) {
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notReady := func(context.Context) error { return errors.New("database is down") }
	go func() { _ = Run(ctx, "test", addr, http.NewServeMux(), notReady) }()

	body, code := getHealthz(t, addr)
	if code != http.StatusServiceUnavailable {
		t.Errorf("/healthz = %d, want 503", code)
	}
	if !strings.Contains(body, "database is down") {
		t.Errorf("/healthz body = %q, want the ready error", body)
	}
}

func TestRunReturnsTheListenError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if err := Run(context.Background(), "test", ln.Addr().String(), http.NewServeMux(), nil); err == nil {
		t.Error("Run = nil, want the address-in-use error")
	}
}

func getHealthz(t *testing.T, addr string) (body string, status int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		res, err := http.Get("http://" + addr + "/healthz")
		if err != nil {
			if time.Now().After(deadline) {
				t.Fatalf("no answer from %s: %v", addr, err)
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}
		read, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		return string(read), res.StatusCode
	}
}
