package repo

import (
	"database/sql"
)

type Repositories struct {
	Teams *TeamRepo
	Users *UserRepository
	PRs   *PRRepo
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Teams: NewTeamRepo(db),
		Users: NewUserRepository(db),
		PRs:   NewPRRepo(db),
	}
}
