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
}

// NewClient создаёт клиент с таймаутом 5 с.
func NewClient(baseURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: 5 * time.Second}}
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
		return nil, &TooManyRequestsError{RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	default:
		return nil, fmt.Errorf("accrual: unexpected status %d", resp.StatusCode)
	}
}

func parseRetryAfter(v string) time.Duration {
	if secs, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return time.Second
}
