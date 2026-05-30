package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAuthRejectsMissingCookie(t *testing.T) {
	a := New("secret")
	called := false
	h := a.RequireAuth(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.False(t, called)
	assert.Empty(t, res.Cookies(), "must not auto-issue a cookie")
}

func TestRequireAuthRejectsTampered(t *testing.T) {
	a := New("secret")
	called := false
	h := a.RequireAuth(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "1:deadbeef"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.False(t, called)
}

func TestRequireAuthPassesValid(t *testing.T) {
	a := New("secret")
	var observed string
	h := a.RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		observed, _ = UserIDFromContext(r.Context())
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: a.Sign("42")})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "42", observed)
}

func TestSetAuthCookie(t *testing.T) {
	a := New("secret")
	rec := httptest.NewRecorder()
	a.SetAuthCookie(rec, "7")
	res := rec.Result()
	defer res.Body.Close()
	cookies := res.Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, cookieName, cookies[0].Name)
	got, err := a.Verify(cookies[0].Value)
	require.NoError(t, err)
	assert.Equal(t, "7", got)
}
