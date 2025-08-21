package permissionservice

import (
	"context"
	"testing"

	"github.com/eragon-mdi/sso/internal/domain"
	mocks_userrepo "github.com/eragon-mdi/sso/internal/service/sso/permission/mocks/userrepo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPermission_IsAdmin(t *testing.T) {
	ctx := context.Background()

	user := domain.User{ID: "user123"}

	t.Run("user is admin", func(t *testing.T) {
		repo := &mocks_userrepo.UserRepository{}
		repo.On("CheckUserIsAdminByID", mock.Anything, user.ID).Return(true, nil)

		s := New(repo)

		ok, err := s.IsAdmin(ctx, user)
		require.NoError(t, err)
		require.True(t, ok)

		repo.AssertExpectations(t)
	})

	t.Run("user is not admin", func(t *testing.T) {
		repo := &mocks_userrepo.UserRepository{}
		repo.On("CheckUserIsAdminByID", mock.Anything, user.ID).Return(false, nil)

		s := New(repo)

		ok, err := s.IsAdmin(ctx, user)
		require.NoError(t, err)
		require.False(t, ok)

		repo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mocks_userrepo.UserRepository{}
		repo.On("CheckUserIsAdminByID", mock.Anything, user.ID).Return(false, assert.AnError)

		s := New(repo)

		ok, err := s.IsAdmin(ctx, user)
		require.Error(t, err)
		require.False(t, ok)

		repo.AssertExpectations(t)
	})
}
