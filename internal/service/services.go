package service

import (
	"context"

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

type StatsResult struct {
	TotalTeams        int `json:"total_teams"`

	TotalPullRequests int `json:"total_pull_requests"`
	OpenPullRequests  int `json:"open_pull_requests"`
	MergedPullRequests int `json:"merged_pull_requests"`

	TotalUsers        int `json:"total_users"`
	ActiveUsers       int `json:"active_users"`
	users []struct {
		UserID      string `json:"user_id"`
		Assignments int    `json:"assignments"`
	}
}

func (s *Services) GetStats(ctx context.Context) (*StatsResult, error) {
	var result StatsResult
	
	totalUsers, activeUsers, err := s.Users.GetCountUsers(ctx)
	if err != nil {
		return nil, err
	}
	result.TotalUsers = totalUsers
	result.ActiveUsers = activeUsers

	totalTeams, err := s.Teams.CountTeams(ctx)
	if err != nil {
		return nil, err
	}
	result.TotalTeams = totalTeams

	totalPRs, openPRs, mergedPRs, err := s.PRs.GetCountPRs(ctx)
	if err != nil {
		return nil, err
	}
	result.TotalPullRequests = totalPRs
	result.OpenPullRequests = openPRs
	result.MergedPullRequests = mergedPRs

	users, err := s.PRs.GetAllUsersWithAssignmentCounts(ctx)
	if err != nil {
		return nil, err
	}
	result.users = users

	return &result, nil
}
