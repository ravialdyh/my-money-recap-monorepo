package members

import (
	"bytes"
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// members' own key deletes a throwaway Keycloak account the test creates with it, and a second delete of the same
// account counts as done (ARCHITECTURE §7.1 step 4). The realm's test Member is never touched.
func TestMembersKeyDeletesAKeycloakAccount(t *testing.T) {
	kc, secret := os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MEMBERS_TEST_CLIENT_SECRET")
	if kc == "" || secret == "" {
		t.Skip("MMR_TEST_KEYCLOAK and MEMBERS_TEST_CLIENT_SECRET not both set")
	}
	k := Keycloak{Admin: kc + "/admin/realms/mmr", Account: &auth.ServiceAccount{Realm: kc + "/realms/mmr", Client: "members", Secret: secret}}
	ctx := context.Background()
	subject := throwaway(t, k)

	if err := k.DeleteAccount(ctx, subject); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if status := admin(t, k, http.MethodGet, "/users/"+subject, ""); status != http.StatusNotFound {
		t.Errorf("the account after its delete answered %d, want 404", status)
	}
	if err := k.DeleteAccount(ctx, subject); err != nil {
		t.Errorf("a second delete: %v, want done", err)
	}
	wrong := Keycloak{Admin: k.Admin, Account: &auth.ServiceAccount{Realm: k.Account.Realm, Client: "members", Secret: "not-the-secret"}}
	if err := wrong.DeleteAccount(ctx, subject); err == nil || strings.Contains(err.Error(), secret) {
		t.Errorf("a wrong secret: %v, want refused", err)
	}
}

// throwaway creates a Keycloak account with members' own key and answers its subject; the account is deleted when the
// test ends. The realm's test Member is never touched.
func throwaway(t *testing.T, k Keycloak) string {
	t.Helper()
	name := "throwaway-" + strings.ToLower(rand.Text())
	res := request(t, k, http.MethodPost, "/users", `{"username":"`+name+`","email":"`+name+`@example.test","enabled":true}`)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("creating the throwaway account: %s", res.Status)
	}
	subject := path.Base(res.Header.Get("Location"))
	t.Cleanup(func() {
		if err := k.DeleteAccount(context.Background(), subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	return subject
}

// admin answers the status of one admin call made with members' own key.
func admin(t *testing.T, k Keycloak, method, at, body string) int {
	t.Helper()
	return request(t, k, method, at, body).StatusCode
}

func request(t *testing.T, k Keycloak, method, at, body string) *http.Response {
	t.Helper()
	tok, err := k.Account.Token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, k.Admin+at, bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}
