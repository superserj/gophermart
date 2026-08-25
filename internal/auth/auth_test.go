package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	a := New("secret")
	got, err := a.Verify(a.Sign("1"))
	require.NoError(t, err)
	assert.Equal(t, "1", got)
}

func TestVerifyRejectsTampered(t *testing.T) {
	a := New("secret")
	_, err := a.Verify(a.Sign("1") + "ff")
	assert.Error(t, err)
}

func TestVerifyRejectsForeignSecret(t *testing.T) {
	_, err := New("b").Verify(New("a").Sign("1"))
	assert.Error(t, err)
}

func TestVerifyRejectsMalformed(t *testing.T) {
	a := New("secret")
	for _, v := range []string{"", "nocolon", ":sig", "id:zz"} {
		_, err := a.Verify(v)
		assert.Error(t, err, "value %q must be rejected", v)
	}
}

func TestContextRoundTrip(t *testing.T) {
	ctx := WithUserID(context.Background(), "42")
	got, ok := UserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "42", got)
	_, ok = UserIDFromContext(context.Background())
	assert.False(t, ok)
}
