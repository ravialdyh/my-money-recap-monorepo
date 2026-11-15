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

const LatestURL = "https://openexchangerates.org/api/latest.json"

const cacheFor = 36 * time.Hour

const digits = 8

var errUnavailable = errors.New("the exchange rate could not be fetched")

var errNoAppID = errors.New("no Open Exchange Rates App ID is configured")

var fetchClient = &http.Client{Timeout: 10 * time.Second}

type Service struct {
	moneyrecapv1connect.UnimplementedCurrencyServiceHandler
	AppID string
	Rates string
	Cache *redis.Client
	Now   func() time.Time
}

type table struct {
	Timestamp int64                  `json:"timestamp"`
	Rates     map[string]json.Number `json:"rates"`
}

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

func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now().UTC()
}

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

func (t *table) cross(from, to string) (string, bool) {
	f, okFrom := t.rate(from)
	x, okTo := t.rate(to)
	if !okFrom || !okTo {
		return "", false
	}
	return significant(new(big.Rat).Quo(x, f), digits), true
}

func (t *table) rate(code string) (*big.Rat, bool) {
	r, ok := new(big.Rat).SetString(string(t.Rates[code]))
	return r, ok && r.Sign() > 0
}

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

func unsupported(from, to string) error {
	err := connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("the rates site has no rate from %s to %s", from, to))
	if d, derr := connect.NewErrorDetail(&moneyrecapv1.UnsupportedPair{FromCurrencyCode: from, ToCurrencyCode: to}); derr == nil {
		err.AddDetail(d)
	}
	return err
}
