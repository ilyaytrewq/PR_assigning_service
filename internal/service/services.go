package service

import (
	"ilyaytrewq/PR_assigning_service/internal/repo"
)

type Services struct {
	Teams *TeamService
	Users *UserService
	PRs   *PRService
}

func NewServices(
	teamRepo *repo.TeamRepo,
	userRepo *repo.UserRepository,
	prRepo *repo.PRRepo,
) *Services {
	return &Services{
		Teams: NewTeamService(teamRepo, userRepo),
		Users: NewUserService(userRepo, prRepo),
		PRs:   NewPRService(prRepo, userRepo, teamRepo),
	}
}

type PRService struct {
	prs   *repo.PRRepo
	users *repo.UserRepository
	teams *repo.TeamRepo
}
