package service

import (
	"context"
	"github.com/salivare/auth-server/pkg/model"
)

type Users interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
}

type UsersConfig struct {
	Users Users
}

type UsersService struct {
	users Users
}

func NewUsersService(cfg UsersConfig) *UsersService {
	return &UsersService{users: cfg.Users}
}
