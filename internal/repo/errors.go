package repo

import (
	"errors"
)

var (
	ErrPRExists   = errors.New("pr already exists")
	ErrPRNotFound = errors.New("pr not found")

	ErrTeamNotFound = errors.New("team not found")
	ErrTeamExists   = errors.New("team already exists")

	ErrUserNotFound = errors.New("user not found")
)
