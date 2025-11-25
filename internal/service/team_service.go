package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"ilyaytrewq/PR_assigning_service/internal/api"
	"ilyaytrewq/PR_assigning_service/internal/repo"
)

// TeamService handles business logic for teams.
type TeamService struct {
	db    *sql.DB
	teams *repo.TeamRepo
	users *repo.UserRepository
}

// NewTeamService creates a new TeamService instance.
func NewTeamService(db *sql.DB, teams *repo.TeamRepo, users *repo.UserRepository) *TeamService {
	return &TeamService{db: db, teams: teams, users: users}
}

// AddTeam creates a new team and its members.
func (s *TeamService) AddTeam(ctx context.Context, team *api.Team) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx AddTeam: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("AddTeam rollback error: %v", rbErr)
			}
		}
	}()

	if err = s.teams.InsertTeamTx(ctx, tx, team); err != nil {
		if errors.Is(err, repo.ErrTeamExists) {
			return ErrTeamAlreadyExists
		}
		return err
	}

	for _, member := range team.Members {
		u := &api.User{
			UserId:   member.UserId,
			Username: member.Username,
			TeamName: team.TeamName,
			IsActive: member.IsActive,
		}

		if err = s.users.InsertOrUpdateTx(ctx, tx, u); err != nil {
			return fmt.Errorf("add team %s: user %s: %w", team.TeamName, member.UserId, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx AddTeam: %w", err)
	}

	return nil
}

// GetTeam retrieves a team by name.
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

// CountTeams returns the total number of teams.
func (s *TeamService) CountTeams(ctx context.Context) (int, error) {
	count, err := s.teams.CountTeams(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}
