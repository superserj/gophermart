// Package accrual опрашивает внешний сервис начислений и обновляет статусы заказов.
package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	// requestTimeout — таймаут одного обращения к accrual (на все попытки суммарно).
	requestTimeout = 5 * time.Second
	// maxAttempts — общее число попыток запроса: одна основная плюс ретраи.
	maxAttempts = 3
	// retryBackoff — базовая пауза между попытками; растёт линейно с номером попытки.
	retryBackoff = 100 * time.Millisecond
	// defaultRetryAfter — пауза по умолчанию, если 429 пришёл без валидного Retry-After.
	defaultRetryAfter = time.Second
)

// ErrNoContent — accrual вернул 204: заказ ещё не зарегистрирован.
var ErrNoContent = errors.New("accrual: order not registered")

// TooManyRequestsError — accrual вернул 429 с Retry-After.
type TooManyRequestsError struct {
	RetryAfter time.Duration
}

// Error возвращает текст ошибки с длительностью Retry-After.
func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("accrual: too many requests, retry after %s", e.RetryAfter)
}

// OrderInfo — тело ответа accrual при 200.
type OrderInfo struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

// Client — HTTP-клиент к accrual.
type Client struct {
	baseURL string
	http    *http.Client
	log     *zap.Logger
}

// NewClient создаёт клиент с таймаутом requestTimeout и прозрачными ретраями
// транспортного уровня на сетевые ошибки и 5xx. Логгер передаётся явно.
func NewClient(baseURL string, log *zap.Logger) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout:   requestTimeout,
			Transport: &retryTransport{base: http.DefaultTransport, attempts: maxAttempts, backoff: retryBackoff},
		},
		log: log,
	}
}

// GetOrder запрашивает GET /api/orders/{number}.
func (c *Client) GetOrder(ctx context.Context, number string) (*OrderInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/orders/"+number, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var info OrderInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return nil, err
		}
		return &info, nil
	case http.StatusNoContent:
		return nil, ErrNoContent
	case http.StatusTooManyRequests:
		return nil, &TooManyRequestsError{RetryAfter: c.parseRetryAfter(resp.Header.Get("Retry-After"))}
	default:
		return nil, fmt.Errorf("accrual: unexpected status %d", resp.StatusCode)
	}
}

// parseRetryAfter переводит заголовок Retry-After в длительность. Непустое, но
// непарсимое значение логируется: это сигнал возможной смены контракта accrual.
func (c *Client) parseRetryAfter(v string) time.Duration {
	trimmed := strings.TrimSpace(v)
	secs, err := strconv.Atoi(trimmed)
	if err != nil {
		if trimmed != "" {
			c.log.Warn("accrual: unparseable Retry-After header", zap.String("value", v))
		}
		return defaultRetryAfter
	}
	if secs <= 0 {
		return defaultRetryAfter
	}
	return time.Duration(secs) * time.Second
}

// retryTransport — http.RoundTripper, повторяющий запрос на сетевых ошибках и 5xx.
// Применяется только к идемпотентным GET-запросам без тела, поэтому повтор безопасен.
type retryTransport struct {
	base     http.RoundTripper
	attempts int
	backoff  time.Duration
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 1; attempt <= t.attempts; attempt++ {
		resp, err = t.base.RoundTrip(req)
		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			return resp, nil
		}
		if attempt == t.attempts {
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(t.backoff * time.Duration(attempt)):
		}
	}
	return resp, err
}
