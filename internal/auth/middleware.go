package auth

import "net/http"

const (
	cookieName   = "user_id"
	cookieMaxAge = 60 * 60 * 24 * 30
)

// SetAuthCookie выставляет подписанную cookie с userID.
func (a *Authenticator) SetAuthCookie(w http.ResponseWriter, userID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    a.Sign(userID),
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
	})
}

// RequireAuth пропускает только запросы с валидной cookie; иначе 401.
func (a *Authenticator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID, err := a.Verify(cookie.Value)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
	})
}
