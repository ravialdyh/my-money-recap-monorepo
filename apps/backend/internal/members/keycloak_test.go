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
	admin := func(method, url string, body string) *http.Response {
		t.Helper()
		tok, err := k.Account.Token(ctx)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(body))
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
	name := "throwaway-" + strings.ToLower(rand.Text())
	created := admin(http.MethodPost, k.Admin+"/users", `{"username":"`+name+`","email":"`+name+`@example.test","enabled":true}`)
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("creating the throwaway account: %s", created.Status)
	}
	subject := path.Base(created.Header.Get("Location"))
	t.Cleanup(func() {
		if err := k.DeleteAccount(context.Background(), subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})

	if err := k.DeleteAccount(ctx, subject); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if res := admin(http.MethodGet, k.Admin+"/users/"+subject, ""); res.StatusCode != http.StatusNotFound {
		t.Errorf("the account after its delete: %s, want 404", res.Status)
	}
	if err := k.DeleteAccount(ctx, subject); err != nil {
		t.Errorf("a second delete: %v, want done", err)
	}
	wrong := Keycloak{Admin: k.Admin, Account: &auth.ServiceAccount{Realm: k.Account.Realm, Client: "members", Secret: "not-the-secret"}}
	if err := wrong.DeleteAccount(ctx, subject); err == nil || strings.Contains(err.Error(), secret) {
		t.Errorf("a wrong secret: %v, want refused", err)
	}
}
