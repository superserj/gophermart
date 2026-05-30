package handler

import (
	"context"
	"sync"

	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
	"github.com/superserj/gophermart/internal/repository"
)

// fakeStore реализует handler.Store. Растёт вместе с интерфейсом по инкрементам.
type fakeStore struct {
	mu     sync.Mutex
	byID   map[string]int64
	hashes map[string]string
	nextID int64

	createErr error
	getErr    error

	// orders (Task 19)
	saveExisted bool
	saveErr     error
	orders      []model.Order
	ordersErr   error
	gotNumber   string
	gotUserID   int64
	saveCalled  bool

	// balance (Task 28-30)
	balCurrent   money.Points
	balWithdrawn money.Points
	balErr       error
	withdrawErr  error
	gotOrder     string
	gotSum       money.Points
	list         []model.Withdrawal
	listErr      error
}

func newFakeStore() *fakeStore {
	return &fakeStore{byID: map[string]int64{}, hashes: map[string]string{}}
}

func (f *fakeStore) Ping(context.Context) error { return nil }

func (f *fakeStore) CreateUser(_ context.Context, login, passwordHash string) (int64, error) {
	if f.createErr != nil {
		return 0, f.createErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[login]; ok {
		return 0, repository.ErrLoginTaken
	}
	f.nextID++
	f.byID[login] = f.nextID
	f.hashes[login] = passwordHash
	return f.nextID, nil
}

func (f *fakeStore) GetUserByLogin(_ context.Context, login string) (int64, string, error) {
	if f.getErr != nil {
		return 0, "", f.getErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.byID[login]
	if !ok {
		return 0, "", repository.ErrUserNotFound
	}
	return id, f.hashes[login], nil
}

func (f *fakeStore) SaveOrder(_ context.Context, number string, userID int64) (bool, error) {
	f.saveCalled = true
	f.gotNumber = number
	f.gotUserID = userID
	return f.saveExisted, f.saveErr
}

func (f *fakeStore) ListOrdersByUser(_ context.Context, userID int64) ([]model.Order, error) {
	f.gotUserID = userID
	return f.orders, f.ordersErr
}

func (f *fakeStore) GetBalance(_ context.Context, _ int64) (money.Points, money.Points, error) {
	return f.balCurrent, f.balWithdrawn, f.balErr
}

func (f *fakeStore) Withdraw(_ context.Context, userID int64, order string, sum money.Points) error {
	f.gotUserID, f.gotOrder, f.gotSum = userID, order, sum
	return f.withdrawErr
}

func (f *fakeStore) ListWithdrawals(_ context.Context, _ int64) ([]model.Withdrawal, error) {
	return f.list, f.listErr
}

var _ Store = (*fakeStore)(nil)
