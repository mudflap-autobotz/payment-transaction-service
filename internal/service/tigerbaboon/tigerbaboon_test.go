package tigerbaboon_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain/mocks"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/service/tigerbaboon"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errDatabase = errors.New("database error")

func setup(t *testing.T) (*mocks.TigerbaboonRepository, *tigerbaboon.TigerbaboonService) {
	t.Helper()
	repo := mocks.NewTigerbaboonRepository(t)
	return repo, tigerbaboon.NewTigerbaboonService(repo)
}

func TestTigerbaboonServiceGetByID(t *testing.T) {
	t.Run("returns tigerbaboon when found", func(t *testing.T) {
		repo, svc := setup(t)
		expected := &domain.Tigerbaboon{ID: 1, Username: "alice"}
		repo.EXPECT().GetByID(mock.Anything, 1).Return(expected, nil)

		got, err := svc.GetByID(context.Background(), 1)

		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("returns not found error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().GetByID(mock.Anything, 99).Return(nil, domain.ErrTigerbaboonNotFound)

		got, err := svc.GetByID(context.Background(), 99)

		require.ErrorIs(t, err, domain.ErrTigerbaboonNotFound)
		assert.Nil(t, got)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().GetByID(mock.Anything, 1).Return(nil, errDatabase)

		got, err := svc.GetByID(context.Background(), 1)

		require.ErrorIs(t, err, errDatabase)
		assert.Nil(t, got)
	})
}

func TestTigerbaboonServiceGetList(t *testing.T) {
	query := domain.ListQuery{Page: 1, Size: 10, OrderBy: "desc", SortBy: "id"}

	t.Run("returns list", func(t *testing.T) {
		repo, svc := setup(t)
		expected := &domain.TigerbaboonList{
			Tigerbaboons: []domain.Tigerbaboon{{ID: 1, Username: "alice"}},
			Total:        1,
		}
		repo.EXPECT().GetList(mock.Anything, query).Return(expected, nil)

		got, err := svc.GetList(context.Background(), query)

		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().GetList(mock.Anything, query).Return(nil, errDatabase)

		got, err := svc.GetList(context.Background(), query)

		require.ErrorIs(t, err, errDatabase)
		assert.Nil(t, got)
	})
}

func TestTigerbaboonServiceCreate(t *testing.T) {
	t.Run("creates tigerbaboon", func(t *testing.T) {
		repo, svc := setup(t)
		expected := &domain.Tigerbaboon{ID: 1, Username: "alice"}
		repo.EXPECT().Create(mock.Anything, &domain.Tigerbaboon{Username: "alice"}).Return(expected, nil)

		got, err := svc.Create(context.Background(), "alice")

		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("returns conflict error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().Create(mock.Anything, &domain.Tigerbaboon{Username: "alice"}).Return(nil, domain.ErrTigerbaboonConflict)

		got, err := svc.Create(context.Background(), "alice")

		require.ErrorIs(t, err, domain.ErrTigerbaboonConflict)
		assert.Nil(t, got)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().Create(mock.Anything, &domain.Tigerbaboon{Username: "alice"}).Return(nil, errDatabase)

		got, err := svc.Create(context.Background(), "alice")

		require.ErrorIs(t, err, errDatabase)
		assert.Nil(t, got)
	})
}

func TestTigerbaboonServiceUpdate(t *testing.T) {
	t.Run("updates tigerbaboon", func(t *testing.T) {
		repo, svc := setup(t)
		expected := &domain.Tigerbaboon{ID: 1, Username: "bob"}
		repo.EXPECT().Update(mock.Anything, &domain.Tigerbaboon{ID: 1, Username: "bob"}).Return(expected, nil)

		got, err := svc.Update(context.Background(), 1, "bob")

		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("returns not found error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().Update(mock.Anything, &domain.Tigerbaboon{ID: 99, Username: "bob"}).Return(nil, domain.ErrTigerbaboonNotFound)

		got, err := svc.Update(context.Background(), 99, "bob")

		require.ErrorIs(t, err, domain.ErrTigerbaboonNotFound)
		assert.Nil(t, got)
	})

	t.Run("returns conflict error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().Update(mock.Anything, &domain.Tigerbaboon{ID: 1, Username: "bob"}).Return(nil, domain.ErrTigerbaboonConflict)

		got, err := svc.Update(context.Background(), 1, "bob")

		require.ErrorIs(t, err, domain.ErrTigerbaboonConflict)
		assert.Nil(t, got)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().Update(mock.Anything, &domain.Tigerbaboon{ID: 1, Username: "bob"}).Return(nil, errDatabase)

		got, err := svc.Update(context.Background(), 1, "bob")

		require.ErrorIs(t, err, errDatabase)
		assert.Nil(t, got)
	})
}

func TestTigerbaboonServiceDelete(t *testing.T) {
	t.Run("deletes tigerbaboon", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().Delete(mock.Anything, 1).Return(nil)

		err := svc.Delete(context.Background(), 1)

		require.NoError(t, err)
	})

	t.Run("returns not found error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().Delete(mock.Anything, 99).Return(domain.ErrTigerbaboonNotFound)

		err := svc.Delete(context.Background(), 99)

		require.ErrorIs(t, err, domain.ErrTigerbaboonNotFound)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repo, svc := setup(t)
		repo.EXPECT().Delete(mock.Anything, 1).Return(errDatabase)

		err := svc.Delete(context.Background(), 1)

		require.ErrorIs(t, err, errDatabase)
	})
}
