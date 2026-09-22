package bff

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// burstSize is one more request than the edge's stricter magic-link zone allows at once, so the last
// one is refused however long ago that zone last saw this address (ARCHITECTURE §5, infra/nginx).
const burstSize = 6

// fallbackCopy is the whole of the A3 web fallback page (SCREEN-STATES A3), which says to open the
// app on the phone and carries no control.
const fallbackCopy = "Open My Money Recap on your phone to finish signing in."

// TestEdge drives the Nginx edge in front of the same Compose stack (#72, spec #61). It skips unless
// BFF_TEST_EDGE is set beside three of TestEndToEnd's variables; infra/compose/README.md lists them.
func TestEdge(t *testing.T) {
	edge, base := os.Getenv("BFF_TEST_EDGE"), os.Getenv("BFF_TEST_URL")
	kc, pw := os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD")
	if edge == "" || base == "" || kc == "" || pw == "" {
		t.Skip("BFF_TEST_EDGE, BFF_TEST_URL, MMR_TEST_KEYCLOAK and MMR_TEST_USER_PASSWORD not all set")
	}
	ctx := t.Context()
	token := testauth.UserToken(t, kc, pw)

	t.Run("GetHome through the edge answers what the BFF answers", func(t *testing.T) {
		direct, through := home(ctx, t, base, token), home(ctx, t, edge+"/api", token)
		if !proto.Equal(direct, through) {
			t.Fatalf("through the edge: %v, straight to the BFF: %v", through, direct)
		}
		t.Logf("%s and %s/api both answer %v", base, edge, through)
	})

	t.Run("the fallback page and the association file arrive over plain HTTP with no redirect", func(t *testing.T) {
		page, status, contentType := fetch(ctx, t, edge+"/magic")
		if status != http.StatusOK || !strings.Contains(page, fallbackCopy) {
			t.Fatalf("/magic: status %d, %q", status, page)
		}
		t.Logf("/magic → %d %s, %q", status, contentType, fallbackCopy)

		tapped, _, _ := fetch(ctx, t, edge+"/magic?lt=lt:00000000-0000-4000-8000-000000000000")
		if tapped != page {
			t.Error("/magic answered a different page when the link carried its login token")
		}
		t.Logf("/magic?lt=… → the same %d bytes: the fallback never touches the token", len(tapped))

		body, status, contentType := fetch(ctx, t, edge+"/.well-known/apple-app-site-association")
		if status != http.StatusOK || contentType != "application/json" {
			t.Fatalf("the association file: status %d, Content-Type %q", status, contentType)
		}
		var association struct{ Applinks json.RawMessage }
		if err := json.Unmarshal([]byte(body), &association); err != nil || association.Applinks == nil {
			t.Fatalf("the association file is not an applinks document: %v, %q", err, body)
		}
		t.Logf("/.well-known/apple-app-site-association → %d %s, applinks %s", status, contentType, association.Applinks)
	})

	t.Run("the stricter zone is the RequestMagicLink path alone", func(t *testing.T) {
		magic := burst(ctx, t, edge+"/api/"+moneyrecapv1connect.BffServiceRequestMagicLinkProcedure[1:], "", `{"email":""}`)
		if !slices.Contains(magic, http.StatusServiceUnavailable) {
			t.Fatalf("%d requests to the magic-link path: %v, want one refused by the edge", burstSize, magic)
		}
		for _, status := range magic {
			if status != http.StatusServiceUnavailable && status != http.StatusBadRequest {
				t.Fatalf("a request the edge let through answered %d, want the BFF's 400 for an empty address", status)
			}
		}
		t.Logf("%d rapid RequestMagicLink → %v: the BFF's 400, then the edge's 503", burstSize, magic)

		gets := burst(ctx, t, edge+"/api/"+moneyrecapv1connect.BffServiceGetHomeProcedure[1:], token, `{}`)
		for _, status := range gets {
			if status != http.StatusOK {
				t.Fatalf("%d rapid GetHome through the edge: %v, want every one answered by the BFF", burstSize, gets)
			}
		}
		t.Logf("%d rapid GetHome → %v: the stricter zone matched none of them", burstSize, gets)
	})
}

// home is GetHome for token against one base address, which is the BFF itself or the edge's /api.
func home(ctx context.Context, t *testing.T, base, token string) *moneyrecapv1.HomeView {
	t.Helper()
	res, err := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(token)).
		GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	return res.Msg.Home
}

// fetch is the plain HTTP get a laptop makes, on a client that reports a redirect rather than
// following one, because the association file may not be behind one (ARCHITECTURE §5).
func fetch(ctx context.Context, t *testing.T, address string) (body string, status int, contentType string) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	read, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(read), res.StatusCode, res.Header.Get("Content-Type")
}

// burst posts the same body burstSize times as fast as one connection allows and answers the status
// codes, so an edge zone that refuses one of them is visible as a status rather than as an RPC error.
func burst(ctx context.Context, t *testing.T, address, token, body string) []int {
	t.Helper()
	statuses := make([]int, 0, burstSize)
	for range burstSize {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewReader([]byte(body)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
		statuses = append(statuses, res.StatusCode)
	}
	return statuses
}
