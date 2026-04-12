package services

import (
	"context"
	"user-service/domain/dto"
	"user-service/domain/models"
	repoUser "user-service/repositories/user"

	"github.com/stretchr/testify/mock"
)

// mockUserRepository implements IUserRepository
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Register(ctx context.Context, req *dto.RegisterRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, req *dto.UpdateRequest, uuid string) (*models.User, error) {
	args := m.Called(ctx, req, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepository) FindByUUID(ctx context.Context, uuid string) (*models.User, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// mockRepositoryRegistry implements IRepositoryRegistry
type mockRepositoryRegistry struct {
	userRepo *mockUserRepository
}

func newMockRegistry() *mockRepositoryRegistry {
	return &mockRepositoryRegistry{userRepo: &mockUserRepository{}}
}

func (m *mockRepositoryRegistry) GetUser() repoUser.IUserRepository {
	return m.userRepo
}
