package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUserAndGet(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()
	id, err := st.CreateUser(ctx, "alice", "hash-1")
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	gotID, gotHash, err := st.GetUserByLogin(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, id, gotID)
	assert.Equal(t, "hash-1", gotHash)
}

func TestCreateUserDuplicate(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()
	_, err := st.CreateUser(ctx, "bob", "h1")
	require.NoError(t, err)
	_, err = st.CreateUser(ctx, "bob", "h2")
	assert.True(t, errors.Is(err, ErrLoginTaken), "want ErrLoginTaken, got %v", err)
}

func TestGetUserNotFound(t *testing.T) {
	st := newTestStore(t)
	_, _, err := st.GetUserByLogin(t.Context(), "ghost")
	assert.True(t, errors.Is(err, ErrUserNotFound), "want ErrUserNotFound, got %v", err)
}
