// Package auth подписывает идентификатор пользователя HMAC-SHA256 и проверяет cookie.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type ctxKey int

const ctxUserID ctxKey = iota

// Authenticator подписывает/проверяет userID по HMAC.
type Authenticator struct {
	secret []byte
}

// New создаёт Authenticator с заданным секретом.
func New(secret string) *Authenticator { return &Authenticator{secret: []byte(secret)} }

// Sign возвращает "<userID>:<hex(hmac)>".
func (a *Authenticator) Sign(userID string) string {
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(userID))
	return userID + ":" + hex.EncodeToString(mac.Sum(nil))
}

// Verify проверяет подпись и возвращает userID.
func (a *Authenticator) Verify(value string) (string, error) {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", errors.New("invalid cookie format")
	}
	got, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("invalid cookie signature encoding")
	}
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(parts[0]))
	if !hmac.Equal(got, mac.Sum(nil)) {
		return "", errors.New("signature mismatch")
	}
	return parts[0], nil
}

// WithUserID кладёт userID в контекст.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserID, userID)
}

// UserIDFromContext достаёт userID из контекста.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxUserID).(string)
	return v, ok && v != ""
}
