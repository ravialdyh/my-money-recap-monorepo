package testauth

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// Inbox must wait for the count its caller asks for: it returned on the first message it saw, so a
// two-message assertion did no waiting at all.
func TestInboxWaitsForTheCountItIsGiven(t *testing.T) {
	var polls atomic.Int32
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		found := `{"messages":[{"ID":"one"}]}`
		if polls.Add(1) > 2 {
			found = `{"messages":[{"ID":"one"},{"ID":"two"}]}`
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(found))
	}))
	t.Cleanup(stub.Close)
	if inbox := Inbox(t, stub.URL, "someone@example.test", 2); len(inbox) != 2 {
		t.Fatalf("%d messages, want 2: the helper answered before the stub held them both", len(inbox))
	}
	t.Logf("Inbox polled %d times for 2 messages", polls.Load())
}
