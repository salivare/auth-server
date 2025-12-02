package model

import "time"

type RefreshToken struct {
	TokenHash string    `json:"token_hash" db:"token_hash"`
	UserID    int       `json:"user_id" db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"create_at" db:"create_at"`
}
