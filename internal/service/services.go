package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sync"

	"ilyaytrewq/PR_assigning_service/internal/repo"
)

// Services holds all service instances.
type Services struct {
	db    *sql.DB
	Teams *TeamService
	Users *UserService
	PRs   *PRService
}

// NewServices creates a new Services instance.
func NewServices(
	db *sql.DB,
	teamRepo *repo.TeamRepo,
	userRepo *repo.UserRepository,
	prRepo *repo.PRRepo,
) *Services {
	return &Services{
		db:    db,
		Teams: NewTeamService(db, teamRepo, userRepo),
		Users: NewUserService(userRepo, prRepo),
		PRs:   NewPRService(prRepo, userRepo, teamRepo),
	}
}

// StatsResult holds system statistics.
type StatsResult struct {
	TotalTeams int `json:"total_teams"`

	TotalPullRequests  int `json:"total_pull_requests"`
	OpenPullRequests   int `json:"open_pull_requests"`
	MergedPullRequests int `json:"merged_pull_requests"`

	TotalUsers  int `json:"total_users"`
	ActiveUsers int `json:"active_users"`
	users       []struct {
		UserID      string `json:"user_id"`
		Assignments int    `json:"assignments"`
	}
}

// GetStats retrieves system statistics.
func (s *Services) GetStats(ctx context.Context) (*StatsResult, error) {
	var (
		result StatsResult
		wg     sync.WaitGroup
		mu     sync.Mutex
		errs   = make(chan error, 4)
	)

	wg.Go(func() {
		totalUsers, activeUsers, err := s.Users.GetCountUsers(ctx)
		if err != nil {
			errs <- err
			return
		}
		mu.Lock()
		result.TotalUsers = totalUsers
		result.ActiveUsers = activeUsers
		mu.Unlock()
	})

	wg.Go(func() {
		totalTeams, err := s.Teams.CountTeams(ctx)
		if err != nil {
			errs <- err
			return
		}
		mu.Lock()
		result.TotalTeams = totalTeams
		mu.Unlock()
	})

	wg.Go(func() {
		totalPRs, openPRs, mergedPRs, err := s.PRs.GetCountPRs(ctx)
		if err != nil {
			errs <- err
			return
		}
		mu.Lock()
		result.TotalPullRequests = totalPRs
		result.OpenPullRequests = openPRs
		result.MergedPullRequests = mergedPRs
		mu.Unlock()
	})

	wg.Go(func() {
		users, err := s.PRs.GetAllUsersWithAssignmentCounts(ctx)
		if err != nil {
			errs <- err

			return
		}
		mu.Lock()
		result.users = users
		mu.Unlock()
	})

	wg.Wait()
	close(errs)

	var resErr error
	for err := range errs {
		log.Printf("GetStats error: %v", err)
		resErr = errors.Join(resErr, err)
	}

	if resErr != nil {
		return nil, resErr
	}

	return &result, nil
}
