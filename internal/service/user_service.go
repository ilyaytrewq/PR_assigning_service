package service

import (
	"context"
	"errors"

	"ilyaytrewq/PR_assigning_service/internal/api"
	"ilyaytrewq/PR_assigning_service/internal/repo"
)

type UserService struct {
	users *repo.UserRepository
	prs   *repo.PRRepo
}

func NewUserService(users *repo.UserRepository, prs *repo.PRRepo) *UserService {
	return &UserService{users: users, prs: prs}
}

func (s *UserService) SetIsActive(ctx context.Context, userId string, isActive bool) (*api.User, error) {
	if user, err := s.users.SetIsActive(ctx, userId, isActive); err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	} else {
		return user, nil
	}

}

func (s *UserService) GetReviewPullRequests(ctx context.Context, userId string) ([]*api.PullRequest, error) {
	prs, err := s.prs.GetByReviewer(ctx, userId)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (s *UserService) GetCountUsers(ctx context.Context) (total int, active int, err error) {
	return s.users.CountUsersAndActive(ctx)
}