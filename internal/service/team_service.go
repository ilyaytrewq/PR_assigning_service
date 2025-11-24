package service

import (
	"context"
	"errors"
	"fmt"

	"ilyaytrewq/PR_assigning_service/internal/api"
	"ilyaytrewq/PR_assigning_service/internal/repo"
)

type TeamService struct {
	teams *repo.TeamRepo
	users *repo.UserRepository
}

func NewTeamService(teams *repo.TeamRepo, users *repo.UserRepository) *TeamService {
	return &TeamService{teams: teams, users: users}
}

func (s *TeamService) AddTeam(ctx context.Context, team *api.Team) error {
	if err := s.teams.InsertTeam(ctx, team); err != nil {
		if errors.Is(err, repo.ErrTeamExists) {
			return ErrTeamAlreadyExists
		}
		return err
	}

	for _, m := range team.Members {
		u := &api.User{
			UserId:   m.UserId,
			Username: m.Username,
			TeamName: team.TeamName,
			IsActive: m.IsActive,
		}

		if err := s.users.InsertOrUpdate(ctx, u); err != nil {
			return fmt.Errorf("add team: upsert user %s: %w", m.UserId, err)
		}
	}

	return nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	team, err := s.teams.GetTeam(ctx, teamName)
	if err != nil {
		if errors.Is(err, repo.ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	return team, nil
}

func (s *TeamService) CountTeams(ctx context.Context) (int, error) {
	count, err := s.teams.CountTeams(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}