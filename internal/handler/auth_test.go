package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/auth"
	"github.com/superserj/gophermart/internal/repository"
)

func TestRegisterBodyTooLarge(t *testing.T) {
	f := newFakeStore()
	h, _ := newTestHandler(f)
	body := `{"login":"` + strings.Repeat("a", maxJSONBodyBytes+1) + `","password":"pw"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Empty(t, f.byID, "store must not be called when body exceeds limit")
}

func TestLoginBodyTooLarge(t *testing.T) {
	f := newFakeStore()
	h, _ := newTestHandler(f)
	body := `{"login":"` + strings.Repeat("a", maxJSONBodyBytes+1) + `","password":"pw"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func newTestHandler(store Store) (*Handler, *auth.Authenticator) {
	a := auth.New("test-secret")
	return New(store, a), a
}

func TestRegisterSuccess(t *testing.T) {
	h, _ := newTestHandler(newFakeStore())
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"bob","password":"pw"}`))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	require.NotEmpty(t, res.Cookies(), "register must set auth cookie")
}

func TestRegisterBadJSON(t *testing.T) {
	h, _ := newTestHandler(newFakeStore())
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{not-json`))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRegisterEmptyFields(t *testing.T) {
	h, _ := newTestHandler(newFakeStore())
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"","password":""}`))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRegisterLoginTaken(t *testing.T) {
	f := newFakeStore()
	f.createErr = repository.ErrLoginTaken
	h, _ := newTestHandler(f)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"bob","password":"pw"}`))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusConflict, res.StatusCode)
}

func TestRegisterStorageError(t *testing.T) {
	f := newFakeStore()
	f.createErr = errors.New("db down")
	h, _ := newTestHandler(f)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"bob","password":"pw"}`))
	rec := httptest.NewRecorder()
	h.Register(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestLoginSuccess(t *testing.T) {
	f := newFakeStore()
	hash, err := auth.HashPassword("pw")
	require.NoError(t, err)
	f.byID["bob"] = 5
	f.hashes["bob"] = hash
	h, _ := newTestHandler(f)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"bob","password":"pw"}`))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	require.NotEmpty(t, res.Cookies())
}

func TestLoginBadJSON(t *testing.T) {
	h, _ := newTestHandler(newFakeStore())
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{`))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestLoginUnknownUser(t *testing.T) {
	h, _ := newTestHandler(newFakeStore())
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"ghost","password":"pw"}`))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestLoginWrongPassword(t *testing.T) {
	f := newFakeStore()
	hash, err := auth.HashPassword("right")
	require.NoError(t, err)
	f.byID["bob"] = 5
	f.hashes["bob"] = hash
	h, _ := newTestHandler(f)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"bob","password":"wrong"}`))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestLoginStorageError(t *testing.T) {
	f := newFakeStore()
	f.getErr = errors.New("db down")
	h, _ := newTestHandler(f)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"bob","password":"pw"}`))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
