// Package repo defines repository interfaces and implementations.
package repo

import (
	"errors"
)

var (
	// ErrPRExists indicates that a PR with the given ID already exists.
	ErrPRExists = errors.New("pr already exists")
	// ErrPRNotFound indicates that the requested PR was not found.
	ErrPRNotFound = errors.New("pr not found")

	// ErrTeamNotFound indicates that the requested team was not found.
	ErrTeamNotFound = errors.New("team not found")
	// ErrTeamExists indicates that a team with the given name already exists.
	ErrTeamExists = errors.New("team already exists")

	// ErrUserNotFound indicates that the requested user was not found.
	ErrUserNotFound = errors.New("user not found")
)
