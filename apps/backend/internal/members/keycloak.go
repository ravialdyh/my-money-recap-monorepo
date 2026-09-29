package members

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// Keycloak is the realm's admin API as members' own service account sees it, holding manage-users for the
// deletion job alone (ARCHITECTURE §7.1 step 4, §4.2). Admin is the realm's admin address.
type Keycloak struct {
	Admin   string
	Account *auth.ServiceAccount
}

// adminClient bounds one admin call.
var adminClient = &http.Client{Timeout: 5 * time.Second}

// DeleteAccount deletes the Keycloak account whose id is subject. A 404 counts as done, so a retry after a
// half-finished run is harmless.
func (k Keycloak) DeleteAccount(ctx context.Context, subject string) error {
	tok, err := k.Account.Token(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, k.Admin+"/users/"+url.PathEscape(subject), http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	res, err := adminClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusNotFound {
		return nil
	}
	return fmt.Errorf("deleting a Keycloak account answered %s", res.Status)
}
