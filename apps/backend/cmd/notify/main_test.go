package main

import (
	"os"
	"strings"
	"testing"
)

// The migration step is what Compose runs before notify starts, and a start-up that cannot read its
// database URL must say so rather than serve without one (ARCHITECTURE §10, A13).
func TestRun(t *testing.T) {
	dsn := os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTIFY_TEST_DATABASE_URL not set")
	}
	args := os.Args
	t.Cleanup(func() { os.Args = args })

	t.Setenv("NOTIFY_DATABASE_URL", "")
	os.Args = []string{"notify", "migrate"}
	if err := run(); err == nil || !strings.Contains(err.Error(), "NOTIFY_DATABASE_URL is required") {
		t.Fatalf("no database URL: %v, want it named as required", err)
	}

	t.Setenv("NOTIFY_DATABASE_URL", dsn)
	os.Args = []string{"notify", "serve"}
	if err := run(); err == nil || !strings.Contains(err.Error(), "usage: notify [migrate]") {
		t.Fatalf("an unknown subcommand: %v, want the usage line", err)
	}

	os.Args = []string{"notify", "migrate"}
	if err := run(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Log("migrate applied the schema and returned")
}
