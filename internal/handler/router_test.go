package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi/v5"
)

func TestRouterRegisterThenLogin(t *testing.T) {
	f := newFakeStore()
	h, a := newTestHandler(f)
	srv := httptest.NewServer(NewRouter(h, a))
	defer srv.Close()
	body := `{"login":"alice","password":"pw"}`

	resReg, err := http.Post(srv.URL+"/api/user/register", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer resReg.Body.Close()
	assert.Equal(t, http.StatusOK, resReg.StatusCode)
	require.NotEmpty(t, resReg.Cookies())

	resLogin, err := http.Post(srv.URL+"/api/user/login", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer resLogin.Body.Close()
	assert.Equal(t, http.StatusOK, resLogin.StatusCode)
}

func TestRouterPing(t *testing.T) {
	h, a := newTestHandler(newFakeStore())
	srv := httptest.NewServer(NewRouter(h, a))
	defer srv.Close()
	res, err := http.Get(srv.URL + "/ping")
	require.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestRouterProtectedRequiresAuth(t *testing.T) {
	h, a := newTestHandler(newFakeStore())
	r := NewRouter(h, a)
	r.Group(func(pr chi.Router) {
		pr.Use(a.RequireAuth)
		pr.Get("/protected-probe", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	})
	srv := httptest.NewServer(r)
	defer srv.Close()

	res1, err := http.Get(srv.URL + "/protected-probe")
	require.NoError(t, err)
	defer res1.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, res1.StatusCode)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/protected-probe", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: a.Sign("1")})
	res2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res2.Body.Close()
	assert.Equal(t, http.StatusOK, res2.StatusCode)
}
