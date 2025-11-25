package repo

import (
	"database/sql"
)

// Repositories holds all repository instances.
type Repositories struct {
	Teams *TeamRepo
	Users *UserRepository
	PRs   *PRRepo
}

// NewRepositories creates a new Repositories instance.
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Teams: NewTeamRepo(db),
		Users: NewUserRepository(db),
		PRs:   NewPRRepo(db),
	}
}
