// Package currency quotes the day's exchange rate from Open Exchange Rates' USD table, cached in Redis (ARCHITECTURE
// §7.4, PRD D35). It never touches an amount (A2) and holds nothing about any Member.
package currency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/redis/go-redis/v9"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/answers"
)

// LatestURL is the provider's table of the latest rates against the US dollar, the only base its free plan serves.
const LatestURL = "https://openexchangerates.org/api/latest.json"

// cacheFor is how long a day's table stays in Redis under fx:{date}:USD (spec #169).
const cacheFor = 36 * time.Hour

// digits is how many significant digits a quoted rate keeps (spec #169).
const digits = 8

// errUnavailable is every failed fetch, and no App ID: the phone shows D6's failed state (PRD §8.9 step 1).
var errUnavailable = errors.New("the exchange rate could not be fetched")

// errNoAppID is a stack with no Open Exchange Rates App ID, as every laptop stack is.
var errNoAppID = errors.New("no Open Exchange Rates App ID is configured")

// fetchClient bounds one fetch from the provider.
var fetchClient = &http.Client{Timeout: 10 * time.Second}

// Service implements CurrencyService. An empty AppID answers unavailable without a fetch; Rates is the table's URL.
type Service struct {
	moneyrecapv1connect.UnimplementedCurrencyServiceHandler
	AppID string
	Rates string
	Cache *redis.Client
	// Now names the day's cache key; nil means time.Now.
	Now func() time.Time
}

// table is the provider's answer: units of each code per US dollar, and the Unix time it was taken.
type table struct {
	Timestamp int64                  `json:"timestamp"`
	Rates     map[string]json.Number `json:"rates"`
}

// Quote is the cross rate from the day's table, dated by it: failed_precondition with UnsupportedPair for a code the
// table lacks, unavailable when the table cannot be had.
func (s *Service) Quote(ctx context.Context, req *connect.Request[moneyrecapv1.QuoteRequest]) (*connect.Response[moneyrecapv1.QuoteResponse], error) {
	from, to := req.Msg.FromCurrencyCode, req.Msg.ToCurrencyCode
	if err := errors.Join(answers.CurrencyCode(from), answers.CurrencyCode(to)); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	t, err := s.today(ctx)
	if err != nil {
		log.Printf("currency: no table: %v", err)
		return nil, connect.NewError(connect.CodeUnavailable, errUnavailable)
	}
	rate, ok := t.cross(from, to)
	if !ok {
		return nil, unsupported(from, to)
	}
	return connect.NewResponse(&moneyrecapv1.QuoteResponse{
		Rate: rate, Source: moneyrecapv1.RateSource_RATE_SOURCE_PROVIDER, AsOf: time.Unix(t.Timestamp, 0).UTC().Format(time.DateOnly),
	}), nil
}

// today is the day's table from Redis, or fetched and cached when Redis has none, holds a broken one or is down.
func (s *Service) today(ctx context.Context) (*table, error) {
	if s.AppID == "" {
		return nil, errNoAppID
	}
	key := "fx:" + s.now().Format(time.DateOnly) + ":USD"
	if body, err := s.Cache.Get(ctx, key).Bytes(); err == nil {
		if t, err := parse(body); err == nil {
			return t, nil
		}
	}
	body, err := s.fetch(ctx)
	if err != nil {
		return nil, err
	}
	t, err := parse(body)
	if err != nil {
		return nil, err
	}
	if err := s.Cache.Set(ctx, key, body, cacheFor).Err(); err != nil {
		log.Printf("currency: the table was not cached: %v", err)
	}
	return t, nil
}

// now is the clock in UTC: Now, or time.Now when it is nil.
func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now().UTC()
}

// fetch is the provider's table, the App ID sent as a header so no address carries it; any answer but 200 fails.
func (s *Service) fetch(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Rates, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+s.AppID)
	res, err := fetchClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the rates site answered %s", res.Status)
	}
	return io.ReadAll(io.LimitReader(res.Body, 1<<20))
}

// parse reads a table, refusing one with no timestamp or no rates.
func parse(body []byte) (*table, error) {
	var t table
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}
	if t.Timestamp <= 0 || len(t.Rates) == 0 {
		return nil, errors.New("the rates table has no timestamp or no rates")
	}
	return &t, nil
}

// cross is units of to per one from, rounded to 8 significant digits; false when either code has no positive rate.
func (t *table) cross(from, to string) (string, bool) {
	f, okFrom := t.rate(from)
	x, okTo := t.rate(to)
	if !okFrom || !okTo {
		return "", false
	}
	return significant(new(big.Rat).Quo(x, f), digits), true
}

// rate is the table's exact decimal for code, and whether it is a positive number.
func (t *table) rate(code string) (*big.Rat, bool) {
	r, ok := new(big.Rat).SetString(string(t.Rates[code]))
	return r, ok && r.Sign() > 0
}

// significant is r, positive, as a plain decimal string rounded half away from zero to n significant digits, with no
// trailing zeros after the point.
func significant(r *big.Rat, n int) string {
	e := len(r.Num().String()) - len(r.Denom().String())
	if r.Cmp(pow10(e)) < 0 {
		e--
	}
	places := n - 1 - e
	if places < 0 {
		return new(big.Rat).Quo(r, pow10(-places)).FloatString(0) + strings.Repeat("0", -places)
	}
	s := r.FloatString(places)
	if places == 0 {
		return s
	}
	return strings.TrimSuffix(strings.TrimRight(s, "0"), ".")
}

// pow10 is ten to the power e, e negative included.
func pow10(e int) *big.Rat {
	abs := e
	if e < 0 {
		abs = -e
	}
	p := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs)), nil))
	if e < 0 {
		return p.Inv(p)
	}
	return p
}

// unsupported is failed_precondition carrying UnsupportedPair (D35): the phone asks the Member for a rate.
func unsupported(from, to string) error {
	err := connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("the rates site has no rate from %s to %s", from, to))
	if d, derr := connect.NewErrorDetail(&moneyrecapv1.UnsupportedPair{FromCurrencyCode: from, ToCurrencyCode: to}); derr == nil {
		err.AddDetail(d)
	}
	return err
}
