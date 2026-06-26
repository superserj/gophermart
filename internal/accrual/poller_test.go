package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

// newTestPoller собирает поллер с no-op логгером для тестов.
func newTestPoller(repo Repository, client orderFetcher) *Poller {
	return NewPoller(repo, client, zap.NewNop())
}

// countingFetcher всегда отвечает 429 и считает число обращений.
type countingFetcher struct {
	calls      int32
	retryAfter time.Duration
}

func (c *countingFetcher) GetOrder(context.Context, string) (*OrderInfo, error) {
	atomic.AddInt32(&c.calls, 1)
	return nil, &TooManyRequestsError{RetryAfter: c.retryAfter}
}

type mockRepo struct {
	mu       sync.Mutex
	statuses map[string]string
	accruals map[string]money.Points
}

func newMockRepo(numbers ...string) *mockRepo {
	m := &mockRepo{statuses: map[string]string{}, accruals: map[string]money.Points{}}
	for _, n := range numbers {
		m.statuses[n] = model.StatusNew
	}
	return m
}

func (m *mockRepo) ListUnfinishedOrders(context.Context) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []string
	for n, s := range m.statuses {
		if s == model.StatusNew || s == model.StatusProcessing {
			out = append(out, n)
		}
	}
	return out, nil
}

func (m *mockRepo) ApplyAccrual(_ context.Context, number, status string, accrual money.Points) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cur := m.statuses[number]; cur == model.StatusInvalid || cur == model.StatusProcessed {
		return nil
	}
	m.statuses[number] = status
	m.accruals[number] = accrual
	return nil
}

func (m *mockRepo) get(number string) (string, money.Points) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statuses[number], m.accruals[number]
}

// stubFetcher всегда отдаёт фиксированный OrderInfo.
type stubFetcher struct{ info *OrderInfo }

func (s *stubFetcher) GetOrder(context.Context, string) (*OrderInfo, error) { return s.info, nil }

func TestPollerClampsNegativeAccrual(t *testing.T) {
	repo := newMockRepo("79927398713")
	p := newTestPoller(repo, &stubFetcher{info: &OrderInfo{Order: "79927398713", Status: "PROCESSED", Accrual: -50}})
	p.process(context.Background(), "79927398713")
	s, a := repo.get("79927398713")
	assert.Equal(t, model.StatusProcessed, s)
	assert.Equal(t, money.Points(0), a, "negative accrual must be clamped to 0")
}

func TestPollerPositiveAccrualUnchanged(t *testing.T) {
	repo := newMockRepo("79927398713")
	p := newTestPoller(repo, &stubFetcher{info: &OrderInfo{Order: "79927398713", Status: "PROCESSED", Accrual: 100.5}})
	p.process(context.Background(), "79927398713")
	s, a := repo.get("79927398713")
	assert.Equal(t, model.StatusProcessed, s)
	assert.Equal(t, money.FromFloat(100.5), a)
}

func TestPollerReachesProcessed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order":"79927398713","status":"PROCESSED","accrual":729.98}`))
	}))
	defer srv.Close()

	repo := newMockRepo("79927398713")
	p := newTestPoller(repo, NewClient(srv.URL, zap.NewNop()))
	p.period = 20 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { p.Run(ctx); close(done) }()

	assert.Eventually(t, func() bool {
		s, a := repo.get("79927398713")
		return s == model.StatusProcessed && a == money.FromFloat(729.98)
	}, 2*time.Second, 20*time.Millisecond)
	cancel()
	<-done
}

func TestPollerThrottlesOn429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "10")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	repo := newMockRepo("79927398713")
	p := newTestPoller(repo, NewClient(srv.URL, zap.NewNop()))
	p.period = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { p.Run(ctx); close(done) }()
	time.Sleep(250 * time.Millisecond)
	cancel()
	<-done
	assert.LessOrEqual(t, atomic.LoadInt32(&calls), int32(3))
	s, _ := repo.get("79927398713")
	assert.Equal(t, model.StatusNew, s)
}

func TestPollerProcessSkipsWhenThrottled(t *testing.T) {
	fc := &countingFetcher{retryAfter: time.Hour}
	p := newTestPoller(newMockRepo("79927398713"), fc)
	p.throttle(time.Hour)
	p.process(context.Background(), "79927398713")
	assert.Zero(t, atomic.LoadInt32(&fc.calls), "process must early-return while throttled")
}

func TestPollerThrottleBoundsBurst(t *testing.T) {
	const n = 20
	nums := make([]string, n)
	for i := range nums {
		nums[i] = strconv.Itoa(1000000 + i)
	}
	fc := &countingFetcher{retryAfter: time.Second}
	repo := newMockRepo(nums...)
	p := newTestPoller(repo, fc)
	p.period = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { p.Run(ctx); close(done) }()
	time.Sleep(200 * time.Millisecond)
	before := atomic.LoadInt32(&fc.calls)
	time.Sleep(150 * time.Millisecond)
	after := atomic.LoadInt32(&fc.calls)
	cancel()
	<-done
	assert.LessOrEqual(t, after, int32(pollWorkers), "calls bounded by worker pool while throttled")
	assert.Equal(t, before, after, "no new accrual calls after throttle set")
	for _, num := range nums {
		s, _ := repo.get(num)
		assert.Equal(t, model.StatusNew, s)
	}
}

func TestPollerGracefulShutdown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer srv.Close()
	repo := newMockRepo("1", "2", "3")
	p := newTestPoller(repo, NewClient(srv.URL, zap.NewNop()))
	p.period = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { p.Run(ctx); close(done) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poller did not shut down")
	}
}
