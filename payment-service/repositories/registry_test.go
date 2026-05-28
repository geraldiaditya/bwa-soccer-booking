package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRegistryWithTransactionUsesCallbackRegistry(t *testing.T) {
	db, mock, cleanup := newMockRegistryDB(t)
	defer cleanup()

	registry := NewRepositoryRegistry(db)
	mock.ExpectBegin()
	mock.ExpectCommit()

	err := registry.WithTransaction(context.Background(), func(txRegistry IRepositoryRegistry) error {
		if txRegistry == registry {
			t.Fatal("expected transaction callback to receive a scoped registry")
		}
		if txRegistry.GetTx() == registry.GetTx() {
			t.Fatal("expected transaction callback registry to use transactional database handle")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected transaction to commit, got error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRegistryWithTransactionRollsBackCallbackError(t *testing.T) {
	db, mock, cleanup := newMockRegistryDB(t)
	defer cleanup()

	expectedErr := errors.New("callback failed")
	registry := NewRepositoryRegistry(db)
	mock.ExpectBegin()
	mock.ExpectRollback()

	err := registry.WithTransaction(context.Background(), func(IRepositoryRegistry) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected callback error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func newMockRegistryDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("open sql mock: %v", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		_ = sqlDB.Close()
		t.Fatalf("open gorm db: %v", err)
	}

	return db, mock, func() {
		_ = sqlDB.Close()
	}
}
