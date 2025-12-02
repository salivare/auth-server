package model

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"password" db:"password"`
	CreateAt     time.Time `json:"create_at" db:"create_at"`
}
