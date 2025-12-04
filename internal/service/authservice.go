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

type AuthService struct {
	storage         storage.Storage
	tokenManager    token.Manager
	refreshTokenTTL time.Duration
}

func NewAuthService(
	storage storage.Storage,
	tokenManager token.Manager,
	refreshTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		storage:         storage,
		tokenManager:    tokenManager,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, string, error) {
	var accessToken, refreshToken string

	err := s.storage.RunInTx(
		ctx, func(ctxTx context.Context) error {
			_, err := s.storage.Users().GetByEmail(ctxTx, email)
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
			created, err := s.storage.Users().Save(ctxTx, u)
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
	u, err := s.storage.Users().GetByEmail(ctx, email)
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

	if err := s.storage.Tokens().Save(ctx, rtModel); err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenPlain, nil
}

func (s *AuthService) RefreshUserToken(ctx context.Context, oldPlainToken string) (string, string, error) {
	oldHash, err := s.tokenManager.HashToken(oldPlainToken)

	if err != nil {
		return "", "", err
	}

	rtModel, err := s.storage.Tokens().FindByHash(ctx, oldHash)

	if err != nil {
		return "", "", errors.New("refresh token not found")
	}

	if time.Now().After(rtModel.ExpiresAt) {
		s.storage.Tokens().DeleteByHash(ctx, oldHash)
		return "", "", token.ErrInvalidRefreshToken
	}

	DelTErr := s.storage.Tokens().DeleteByHash(ctx, oldHash)
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
