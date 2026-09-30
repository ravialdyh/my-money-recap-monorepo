package currency

import (
	"context"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/redis/go-redis/v9"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

const testAppID = "test-app-id"

// rates is a fake rates site answering body with status, counting fetches and checking the App ID header.
func rates(t *testing.T, status int, body string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Header.Get("Authorization") != "Token "+testAppID || r.URL.RawQuery != "" {
			http.Error(w, "no App ID header", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(site.Close)
	return site, &hits
}

// service is a Service against site and cache, its clock on day, whose cache key it deletes before and after.
func service(t *testing.T, site string, cache *redis.Client, day string) *Service {
	t.Helper()
	at, err := time.Parse(time.DateOnly, day)
	if err != nil {
		t.Fatal(err)
	}
	drop := func() { _ = cache.Del(context.Background(), "fx:"+day+":USD").Err() }
	drop()
	t.Cleanup(drop)
	return &Service{AppID: testAppID, Rates: site, Cache: cache, Now: func() time.Time { return at }}
}

// liveRedis is the stack's Redis; the tests' clock sits in 2001, so their keys never meet the service's own.
func liveRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("BFF_TEST_REDIS")
	if addr == "" {
		t.Skip("BFF_TEST_REDIS not set")
	}
	cache := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = cache.Close() })
	return cache
}

func quote(s *Service, from, to string) (*moneyrecapv1.QuoteResponse, error) {
	res, err := s.Quote(context.Background(), connect.NewRequest(&moneyrecapv1.QuoteRequest{FromCurrencyCode: from, ToCurrencyCode: to}))
	if err != nil {
		return nil, err
	}
	return res.Msg, nil
}

const usdTable = `{"base":"USD","timestamp":1000000000,"rates":{"USD":1,"GBP":0.75,"IDR":16000,"EUR":0.9}}`

// The cross rate comes from the USD table at 8 significant digits, dated by it, and a second quote reads Redis.
func TestQuoteCrossRateAndCache(t *testing.T) {
	site, hits := rates(t, http.StatusOK, usdTable)
	s := service(t, site.URL, liveRedis(t), "2001-01-01")
	got, err := quote(s, "GBP", "IDR")
	if err != nil {
		t.Fatal(err)
	}
	if got.Rate != "21333.333" || got.Source != moneyrecapv1.RateSource_RATE_SOURCE_PROVIDER || got.AsOf != "2001-09-09" {
		t.Errorf("GBP to IDR = %v, want 21333.333 PROVIDER as of 2001-09-09", got)
	}
	if got, err := quote(s, "IDR", "EUR"); err != nil || got.Rate != "0.00005625" {
		t.Errorf("IDR to EUR = %v, %v; want 0.00005625", got, err)
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("%d fetches for two quotes, want 1: the second reads the day's table from Redis", n)
	}
	ttl, err := s.Cache.TTL(context.Background(), "fx:2001-01-01:USD").Result()
	if err != nil || ttl <= 35*time.Hour || ttl > 36*time.Hour {
		t.Errorf("the cached table lives %v, %v; want 36 hours", ttl, err)
	}
}

// A code the table lacks is failed_precondition carrying UnsupportedPair, and a code no country uses is refused.
func TestQuoteUnsupportedPair(t *testing.T) {
	site, _ := rates(t, http.StatusOK, usdTable)
	s := service(t, site.URL, liveRedis(t), "2001-01-02")
	_, err := quote(s, "GBP", "JPY")
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeFailedPrecondition || len(ce.Details()) != 1 {
		t.Fatalf("GBP to JPY: %v, want failed_precondition with one detail", err)
	}
	pair, derr := ce.Details()[0].Value()
	if p, ok := pair.(*moneyrecapv1.UnsupportedPair); derr != nil || !ok || p.FromCurrencyCode != "GBP" || p.ToCurrencyCode != "JPY" {
		t.Errorf("detail %v, %v; want UnsupportedPair GBP to JPY", pair, derr)
	}
	if _, err := quote(s, "GBP", "XXX"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no Currency: %v, want invalid_argument", err)
	}
}

// A failed fetch, a broken table and no App ID are unavailable, and no App ID fetches nothing.
func TestQuoteUnavailable(t *testing.T) {
	cache := liveRedis(t)
	down, _ := rates(t, http.StatusInternalServerError, "")
	if _, err := quote(service(t, down.URL, cache, "2001-01-03"), "GBP", "IDR"); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("a failed fetch: %v, want unavailable", err)
	}
	broken, _ := rates(t, http.StatusOK, `{"rates":{}}`)
	if _, err := quote(service(t, broken.URL, cache, "2001-01-04"), "GBP", "IDR"); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("a table with no rates: %v, want unavailable", err)
	}
	site, hits := rates(t, http.StatusOK, usdTable)
	s := service(t, site.URL, cache, "2001-01-05")
	s.AppID = ""
	if _, err := quote(s, "GBP", "IDR"); connect.CodeOf(err) != connect.CodeUnavailable || hits.Load() != 0 {
		t.Errorf("no App ID: %v after %d fetches, want unavailable and none", err, hits.Load())
	}
}

// With Redis down every quote fetches the table again and still answers.
func TestQuoteWithRedisDown(t *testing.T) {
	site, hits := rates(t, http.StatusOK, usdTable)
	down := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1})
	t.Cleanup(func() { _ = down.Close() })
	s := &Service{AppID: testAppID, Rates: site.URL, Cache: down}
	for range 2 {
		if got, err := quote(s, "USD", "GBP"); err != nil || got.Rate != "0.75" {
			t.Fatalf("USD to GBP with Redis down = %v, %v; want 0.75", got, err)
		}
	}
	if n := hits.Load(); n != 2 {
		t.Errorf("%d fetches, want 2", n)
	}
}

// significant rounds half away from zero to n digits and never prints an exponent.
func TestSignificant(t *testing.T) {
	for in, want := range map[string]string{
		"1": "1", "20150": "20150", "1/3": "0.33333333", "2/3": "0.66666667", "123456785": "123456790",
		"999999995": "1000000000", "99999999.5": "100000000", "1/80000000": "0.0000000125", "10": "10",
	} {
		r, _ := new(big.Rat).SetString(in)
		if got := significant(r, 8); got != want {
			t.Errorf("significant(%s) = %s, want %s", in, got, want)
		}
	}
}
