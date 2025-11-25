package service

import (
	"context"
	"errors"

	"ilyaytrewq/PR_assigning_service/internal/api"
	"ilyaytrewq/PR_assigning_service/internal/repo"
)

// UserService handles business logic for users.
type UserService struct {
	users *repo.UserRepository
	prs   *repo.PRRepo
}

// NewUserService creates a new UserService instance.
func NewUserService(users *repo.UserRepository, prs *repo.PRRepo) *UserService {
	return &UserService{users: users, prs: prs}
}

// SetIsActive updates a user's active status.
func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	user, err := s.users.SetIsActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetReviewPullRequests retrieves PRs assigned to a reviewer.
func (s *UserService) GetReviewPullRequests(ctx context.Context, userID string) ([]*api.PullRequest, error) {
	prs, err := s.prs.GetByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

// GetCountUsers returns user statistics.
func (s *UserService) GetCountUsers(ctx context.Context) (total int, active int, err error) {
	return s.users.CountUsersAndActive(ctx)
}
