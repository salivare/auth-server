package storage

import (
	"context"
	"errors"
	"github.com/salivare/auth-server/pkg/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrTokenNotFound = errors.New("token not found")

type UserRepository interface {
	Save(ctx context.Context, u model.User) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	List(ctx context.Context) ([]model.User, error)
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, token model.RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	DeleteByHash(ctx context.Context, tokenHash string) error
}

type TxManager interface {
	// RunInTx performs the fn function in a new transaction
	// The context passed to fn contains an active transaction
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Storage interface {
	Users() UserRepository
	Tokens() RefreshTokenRepository
	TxManager
}
