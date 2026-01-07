package service

import (
	"context"
	"errors"
	"github.com/salivare/auth-server/internal/storage"
	"github.com/salivare/auth-server/internal/token"
	"github.com/salivare/auth-server/pkg/model"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrUserExists = errors.New("user exists")
	ErrBadCreds   = errors.New("bad creds")
	ErrNotFound   = errors.New("not found")
)

type User = model.User

//go:generate sh -c "mkdir -p ./mocks && go run github.com/vektra/mockery/v3@v3.6.1 --name=Authentication --dir=. --output=./mocks --outpkg=mocks"
type Authentication interface {
	Save(ctx context.Context, u model.User) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
}

type RefreshTokenRepo interface {
	Save(ctx context.Context, token model.RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	DeleteByHash(ctx context.Context, tokenHash string) error
}

type TxManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type AuthConfig struct {
	Authentication  Authentication
	Tokens          RefreshTokenRepo
	Tx              TxManager
	TokenManager    token.Manager
	RefreshTokenTTL time.Duration
}

type AuthService struct {
	authentication  Authentication
	tokens          RefreshTokenRepo
	tx              TxManager
	tokenManager    token.Manager
	refreshTokenTTL time.Duration
}

func NewAuthService(cfg AuthConfig) *AuthService {
	return &AuthService{
		authentication:  cfg.Authentication,
		tokens:          cfg.Tokens,
		tx:              cfg.Tx,
		tokenManager:    cfg.TokenManager,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, string, error) {
	var accessToken, refreshToken string

	err := s.tx.RunInTx(
		ctx, func(ctxTx context.Context) error {
			_, err := s.authentication.GetByEmail(ctxTx, email)
			if err == nil {
				return ErrUserExists
			}
			if !errors.Is(err, storage.ErrUserNotFound) {
				return err
			}

			if len(password) < 6 {
				return errors.New("password too short")
			}

			hashed, err := hashPassword(password)
			if err != nil {
				return err
			}

			u := User{
				Email:        email,
				PasswordHash: hashed,
			}
			created, err := s.authentication.Save(ctxTx, u)
			if err != nil {
				return err
			}

			accessToken, refreshToken, err = s.GenerateUserSession(ctxTx, created.ID)
			if err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil

}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	u, err := s.authentication.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", "", ErrNotFound
		}
	}
	if err := comparePassword(u.PasswordHash, password); err != nil {
		return "", "", ErrBadCreds
	}
	return s.GenerateUserSession(ctx, u.ID)
}

func (s *AuthService) GenerateUserSession(ctx context.Context, userId int) (string, string, error) {
	accessToken, err := s.tokenManager.GenerateAccessToken(userId)

	if err != nil {
		return "", "", err
	}

	refreshTokenPlain, err := s.tokenManager.GenerateRefreshTokenPlain()

	if err != nil {
		return "", "", err
	}

	refreshTokenHash, err := s.tokenManager.HashToken(refreshTokenPlain)

	if err != nil {
		return "", "", err
	}

	rtExpirestAt := time.Now().Add(s.refreshTokenTTL)

	rtModel := model.RefreshToken{
		TokenHash: refreshTokenHash,
		UserID:    userId,
		ExpiresAt: rtExpirestAt,
		CreatedAt: time.Now(),
	}

	if err := s.tokens.Save(ctx, rtModel); err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenPlain, nil
}

func (s *AuthService) RefreshUserToken(ctx context.Context, oldPlainToken string) (string, string, error) {
	oldHash, err := s.tokenManager.HashToken(oldPlainToken)

	if err != nil {
		return "", "", err
	}

	rtModel, err := s.tokens.FindByHash(ctx, oldHash)

	if err != nil {
		return "", "", errors.New("refresh token not found")
	}

	if time.Now().After(rtModel.ExpiresAt) {
		s.tokens.DeleteByHash(ctx, oldHash)
		return "", "", token.ErrInvalidRefreshToken
	}

	DelTErr := s.tokens.DeleteByHash(ctx, oldHash)
	if DelTErr != nil {
		// log
	}

	return s.GenerateUserSession(ctx, rtModel.UserID)
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func comparePassword(hashedPassword, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plain))
}
