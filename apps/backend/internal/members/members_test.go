package members

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// testDB opens the Compose stack's members database and applies the schema, skipping unless
// MMR_TEST_MEMBERS_DATABASE_URL is set (infra/compose/README.md).
func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	if dsn == "" {
		t.Skip("MMR_TEST_MEMBERS_DATABASE_URL not set")
	}
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestConcurrentFirstCallsCreateOneRow(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	subject := "concurrent-" + t.Name() + "-" + strings.ReplaceAll(t.TempDir(), "/", "")
	t.Cleanup(func() {
		if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})

	const n = 16
	ids := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var m Row
			m, errs[i] = Upsert(ctx, db, subject, "race@example.com")
			ids[i] = m.ID
		}()
	}
	wg.Wait()
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("call %d: %v", i, errs[i])
		}
		if ids[i] != ids[0] {
			t.Fatalf("call %d returned id %s, call 0 returned %s", i, ids[i], ids[0])
		}
	}
	var rows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, subject).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("%d rows for one subject, want 1", rows)
	}
	if staged := outbox(t, db, ids[0]); len(staged) != 1 || staged[0].key != "members.created" {
		t.Fatalf("%d concurrent first calls staged %v, want one members.created", n, staged)
	}
	t.Logf("%d concurrent first calls → 1 row, every call returned member id %s", n, ids[0])
}

// A Member whose email an admin cleared still refreshes a token, one without the claim; the stored
// address must survive it (PRD §7: it addresses the magic link and the CSV export).
func TestAnEmptyEmailClaimKeepsTheStoredAddress(t *testing.T) {
	db := testDB(t)
	subject := "empty-claim-" + t.Name()
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: subject})
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM member WHERE keycloak_subject = $1`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	if _, err := Upsert(ctx, db, subject, "kept@example.com"); err != nil {
		t.Fatal(err)
	}
	res, err := (&Service{DB: db}).GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: subject}))
	if err != nil || res.Msg.Profile.Email != "kept@example.com" {
		t.Fatalf("empty email claim: %v, %v, want the stored kept@example.com", res, err)
	}
}

func TestStoreFailureIsUnavailableWithoutTheDriverText(t *testing.T) {
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: "sub-1", Email: "test@example.com"})
	db, err := pgxpool.New(ctx, "postgres://x:y@127.0.0.1:1/x?connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	_, raw := Upsert(ctx, db, "sub-1", "test@example.com")
	if raw == nil {
		t.Fatal("Upsert reached a database on port 1")
	}
	s := &Service{DB: db}
	_, createErr := s.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: "sub-1", Email: "test@example.com"}))
	_, profileErr := s.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: "00000000-0000-0000-0000-000000000001"}))
	for name, err := range map[string]error{"GetOrCreateBySubject": createErr, "GetProfile": profileErr} {
		if connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "127.0.0.1") || strings.Contains(err.Error(), raw.Error()) {
			t.Errorf("%s: %v, want unavailable without the driver's %q", name, err, raw)
		}
	}
}

func TestTheCallersDeadlineOrCancelIsNotAStoreFailure(t *testing.T) {
	claims := auth.NewContext(context.Background(), auth.Claims{Subject: "sub-1", Email: "test@example.com"})
	db, err := pgxpool.New(claims, "postgres://x:y@127.0.0.1:1/x?connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	s := &Service{DB: db}
	expired, stop := context.WithDeadline(claims, time.Now().Add(-time.Second))
	defer stop()
	cancelled, cancel := context.WithCancel(claims)
	cancel()
	var logged bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(prev) })
	for want, ctx := range map[connect.Code]context.Context{connect.CodeDeadlineExceeded: expired, connect.CodeCanceled: cancelled} {
		_, createErr := s.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: "sub-1", Email: "test@example.com"}))
		_, profileErr := s.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: "00000000-0000-0000-0000-000000000001"}))
		for name, err := range map[string]error{"GetOrCreateBySubject": createErr, "GetProfile": profileErr} {
			if connect.CodeOf(err) != want || strings.Contains(err.Error(), "127.0.0.1") {
				t.Errorf("%s: %v, want %v", name, err, want)
			}
		}
	}
	if strings.Contains(logged.String(), "members store:") {
		t.Errorf("logged the caller's own deadline or cancel as a store failure: %q", logged.String())
	}
}

func TestConnectBoundary(t *testing.T) {
	base := os.Getenv("MEMBERS_TEST_URL")
	if base == "" || os.Getenv("MMR_TEST_KEYCLOAK") == "" || os.Getenv("MMR_TEST_USER_PASSWORD") == "" {
		t.Skip("MEMBERS_TEST_URL, MMR_TEST_KEYCLOAK and MMR_TEST_USER_PASSWORD not all set")
	}
	db := testDB(t)
	ctx := context.Background()
	kc := os.Getenv("MMR_TEST_KEYCLOAK")
	tok := testauth.UserToken(t, kc, os.Getenv("MMR_TEST_USER_PASSWORD"))
	claims := testauth.Decode(t, tok)
	sub, email := claims.Sub, claims.Email
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, sub); err != nil {
		t.Fatal(err)
	}

	client := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, base)
	authed := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	getOrCreate := func(c moneyrecapv1connect.MembersServiceClient, subject, email string) (*moneyrecapv1.Profile, error) {
		res, err := c.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: subject, Email: email}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Profile, nil
	}

	first, err := getOrCreate(authed, sub, email)
	if err != nil {
		t.Fatal(err)
	}
	if first.MemberId == "" || first.Email != email {
		t.Fatalf("first call: %v", first)
	}
	var rows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, sub).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("%d rows after first call, want 1", rows)
	}
	second, err := getOrCreate(authed, sub, email)
	if err != nil {
		t.Fatal(err)
	}
	if second.MemberId != first.MemberId {
		t.Fatalf("second call returned %s, first %s", second.MemberId, first.MemberId)
	}
	t.Logf("first and second call: member id %s", first.MemberId)

	storedEmail := func() string {
		var stored string
		if err := db.QueryRow(ctx, `SELECT email FROM member WHERE id = $1`, first.MemberId).Scan(&stored); err != nil {
			t.Fatal(err)
		}
		return stored
	}
	if _, err := db.Exec(ctx, `UPDATE member SET email = 'stale@example.com' WHERE id = $1`, first.MemberId); err != nil {
		t.Fatal(err)
	}
	third, err := getOrCreate(authed, sub, email)
	if err != nil {
		t.Fatal(err)
	}
	if stored := storedEmail(); third.MemberId != first.MemberId || third.Email != email || stored != email {
		t.Fatalf("email refresh: response %v, stored %q, want the token's %q", third, stored, email)
	}
	t.Logf("stale email refreshed to the token's claim %q", email)

	_, err = getOrCreate(authed, sub, "someone-else@example.com")
	if connect.CodeOf(err) != connect.CodePermissionDenied || !strings.Contains(err.Error(), "email does not match the token") {
		t.Errorf("email mismatch: %v, want permission_denied", err)
	}
	if stored := storedEmail(); stored != email {
		t.Errorf("email mismatch stored %q, want %q unchanged", stored, email)
	}

	prof, err := authed.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: first.MemberId}))
	if err != nil {
		t.Fatal(err)
	}
	if p := prof.Msg.Profile; p.MemberId != first.MemberId || p.Email != email || p.DisplayName != "" || p.Timezone != "" {
		t.Fatalf("GetProfile: %v", p)
	}
	t.Logf("GetProfile: %v", prof.Msg.Profile)

	other := "other-" + t.Name()
	otherRow, err := Upsert(ctx, db, other, "other@example.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, other); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	for name, id := range map[string]string{"another Member's id": otherRow.ID, "unknown id": "00000000-0000-0000-0000-000000000001", "malformed id": "not-a-uuid"} {
		if _, err := authed.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: id})); connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("GetProfile of %s: %v, want not_found", name, err)
		}
	}

	if _, err := getOrCreate(authed, "someone-else", email); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("subject mismatch: %v, want permission_denied", err)
	}
	unauth := map[string]moneyrecapv1connect.MembersServiceClient{
		"no token":      client,
		"garbage token": moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, base, testauth.Bearer("not.a.jwt")),
	}
	if secret := os.Getenv("MEMBERS_TEST_BFF_CLIENT_SECRET"); secret != "" {
		bffTok := testauth.Token(t, kc, url.Values{"grant_type": {"client_credentials"}, "client_id": {"bff"}, "client_secret": {secret}})
		unauth["wrong audience"] = moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, base, testauth.Bearer(bffTok))
	}
	for name, c := range unauth {
		if _, err := getOrCreate(c, sub, email); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("GetOrCreateBySubject %s: %v, want unauthenticated", name, err)
		}
		if _, err := c.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: first.MemberId})); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("GetProfile %s: %v, want unauthenticated", name, err)
		}
		t.Logf("%s: both RPCs unauthenticated", name)
	}
}
