package accrual

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/money"
)

const (
	pollWorkers = 5
	pollPeriod  = time.Second
)

// Repository — узкий набор методов хранилища, нужный поллеру.
type Repository interface {
	ListUnfinishedOrders(ctx context.Context) ([]string, error)
	ApplyAccrual(ctx context.Context, number, status string, accrual money.Points) error
}

type orderFetcher interface {
	GetOrder(ctx context.Context, number string) (*OrderInfo, error)
}

// Poller сканирует незавершённые заказы и опрашивает accrual.
type Poller struct {
	repo   Repository
	client orderFetcher
	log    *zap.Logger
	sem    chan struct{}
	period time.Duration

	mu             sync.Mutex
	throttledUntil time.Time
	wg             sync.WaitGroup
}

// NewPoller создаёт поллер с пулом pollWorkers и периодом pollPeriod.
// Логгер передаётся явно, без обращения к глобальному состоянию.
func NewPoller(repo Repository, client orderFetcher, log *zap.Logger) *Poller {
	return &Poller{repo: repo, client: client, log: log, sem: make(chan struct{}, pollWorkers), period: pollPeriod}
}

// Run сканирует БД по тикеру до отмены контекста, затем дожидается воркеров.
func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			p.wg.Wait()
			return
		case <-ticker.C:
			p.scan(ctx)
		}
	}
}

func (p *Poller) scan(ctx context.Context) {
	if p.isThrottled() {
		return
	}
	numbers, err := p.repo.ListUnfinishedOrders(ctx)
	if err != nil {
		p.log.Warn("list unfinished orders", zap.Error(err))
		return
	}
	for _, number := range numbers {
		if p.isThrottled() || ctx.Err() != nil {
			return
		}
		select {
		case p.sem <- struct{}{}:
		case <-ctx.Done():
			return
		}
		p.wg.Add(1)
		go func(num string) {
			defer p.wg.Done()
			defer func() { <-p.sem }()
			p.process(ctx, num)
		}(number)
	}
}

func (p *Poller) process(ctx context.Context, number string) {
	if p.isThrottled() { // дешёвая защита: троттл мог быть выставлен соседним воркером из того же burst
		return
	}
	info, err := p.client.GetOrder(ctx, number)
	if err != nil {
		var tooMany *TooManyRequestsError
		switch {
		case errors.Is(err, ErrNoContent):
			return
		case errors.As(err, &tooMany):
			p.throttle(tooMany.RetryAfter)
		default:
			p.log.Warn("accrual get order", zap.String("number", number), zap.Error(err))
		}
		return
	}
	status := mapStatus(info.Status)
	if info.Accrual < 0 { // начисление не может быть отрицательным — защищаемся от порчи баланса
		p.log.Warn("negative accrual clamped to zero", zap.String("number", number), zap.Float64("accrual", info.Accrual))
		info.Accrual = 0
	}
	if err := p.repo.ApplyAccrual(ctx, number, status, money.FromFloat(info.Accrual)); err != nil {
		p.log.Warn("apply accrual", zap.String("number", number), zap.Error(err))
	}
}

func (p *Poller) throttle(d time.Duration) {
	until := time.Now().Add(d)
	p.mu.Lock()
	if until.After(p.throttledUntil) { // только ПРОДЛЕВАЕМ троттл, не укорачиваем более длинный Retry-After
		p.throttledUntil = until
	}
	p.mu.Unlock()
}

func (p *Poller) isThrottled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return time.Now().Before(p.throttledUntil)
}
