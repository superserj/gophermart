package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

func TestMain(m *testing.M) {
	_ = logger.Initialize("error")
	os.Exit(m.Run())
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

func TestPollerReachesProcessed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order":"79927398713","status":"PROCESSED","accrual":729.98}`))
	}))
	defer srv.Close()

	repo := newMockRepo("79927398713")
	p := NewPoller(repo, NewClient(srv.URL))
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
	p := NewPoller(repo, NewClient(srv.URL))
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

func TestPollerGracefulShutdown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer srv.Close()
	repo := newMockRepo("1", "2", "3")
	p := NewPoller(repo, NewClient(srv.URL))
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
