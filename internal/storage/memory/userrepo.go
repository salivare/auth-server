package memory

import (
	"context"
	"github.com/salivare/auth-server/internal/storage"
	"github.com/salivare/auth-server/pkg/model"
	"sync"
	"time"
)

var _ storage.UserRepository = (*UserRepo)(nil)

type UserRepo struct {
	mu      sync.RWMutex
	byId    map[int]model.User
	byEmail map[string]int
	next    int
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		byId:    make(map[int]model.User),
		byEmail: make(map[string]int),
		next:    1,
	}
}

func (r *UserRepo) Save(ctx context.Context, u model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if u.ID == 0 {
		u.ID = r.next
		r.next++
	}

	if u.CreateAt.IsZero() {
		u.CreateAt = time.Now().UTC()
	}

	r.byId[u.ID] = u
	r.byEmail[u.Email] = u.ID
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, ok := r.byEmail[email]
	if !ok {
		return model.User{}, storage.ErrUserNotFound
	}

	return r.byId[id], nil
}

func (r *UserRepo) List(ctx context.Context) ([]model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]model.User, 0, len(r.byId))
	for _, u := range r.byId {
		out = append(out, u)
	}
	return out, nil
}
