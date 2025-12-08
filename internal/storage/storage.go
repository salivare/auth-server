package storage

import (
	"errors"
)

var ErrUserNotFound = errors.New("user not found")
var ErrTokenNotFound = errors.New("token not found")
