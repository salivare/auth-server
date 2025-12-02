package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/salivare/auth-server/internal/storage"
	"github.com/salivare/auth-server/pkg/model"
	"time"
)

var _ storage.UserRepository = (*UserRepo)(nil)

var ErrNotFound = errors.New("not found")

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Save(ctx context.Context, u model.User) (model.User, error) {
	if u.ID == 0 {
		now := time.Now().UTC()
		res, err := r.db.ExecContext(
			ctx,
			"INSERT INTO users(email,password,create_at) VALUES (?,?,?)",
			u.Email,
			u.PasswordHash,
			now,
		)
		if err != nil {
			return model.User{}, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return model.User{}, err
		}
		u.ID = int(id)
		u.CreateAt = now

		return u, nil
	}

	_, err := r.db.ExecContext(
		ctx,
		"UPDATE users SET password=? WHERE id=?",
		u.PasswordHash,
		u.ID,
	)
	return u, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, email, password, create_at FROM users WHERE email = ? LIMIT 1`,
		email,
	)

	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreateAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("query user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, email, password, create_at FROM users ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreateAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}
