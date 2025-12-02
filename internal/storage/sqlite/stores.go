package sqlite

import (
	"database/sql"
	"github.com/salivare/auth-server/internal/storage"
)

type Repos struct {
	Users storage.UserRepository
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Users: NewUserRepo(db),
	}
}
