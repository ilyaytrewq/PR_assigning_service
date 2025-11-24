package service

import (
	"context"
	"errors"
	"slices"
	"time"

	"ilyaytrewq/PR_assigning_service/internal/api"
	"ilyaytrewq/PR_assigning_service/internal/repo"
)

type PRService struct {
	prs   *repo.PRRepo
	users *repo.UserRepository
	teams *repo.TeamRepo
}

func NewPRService(prs *repo.PRRepo, users *repo.UserRepository, teams *repo.TeamRepo) *PRService {
	return &PRService{
		prs:   prs,
		users: users,
		teams: teams,
	}
}

func (s *PRService) CreatePR(ctx context.Context, body *api.PostPullRequestCreateJSONBody) (*api.PullRequest, error) {
	author, err := s.users.Get(ctx, body.AuthorId)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	team, err := s.teams.GetTeam(ctx, author.TeamName)
	if err != nil {
		if errors.Is(err, repo.ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}

	var candidates []string
	for _, m := range team.Members {
		if !m.IsActive {
			continue
		}
		if m.UserId == author.UserId {
			continue
		}
		candidates = append(candidates, m.UserId)
	}

	candidates = candidates[0:min(2, len(candidates))]
	now := time.Now().UTC()

	pr := &api.PullRequest{
		AssignedReviewers: candidates,
		AuthorId:          author.UserId,
		CreatedAt:         &now,
		MergedAt:          nil,
		PullRequestId:     body.PullRequestId,
		PullRequestName:   body.PullRequestName,
		Status:            api.PullRequestStatusOPEN,
	}

	if err := s.prs.CreatePR(ctx, pr); err != nil {
		if errors.Is(err, repo.ErrPRExists) {
			return nil, ErrPRAlreadyExists
		}
		return nil, err
	}

	return pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*api.PullRequest, error) {
	now := time.Now().UTC()

	pr, err := s.prs.MergePR(ctx, prID, now)
	if err != nil {
		if errors.Is(err, repo.ErrPRNotFound) {
			return nil, ErrPRNotFound
		}
		return nil, err
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, body *api.PostPullRequestReassignJSONBody) (*api.PullRequest, string, error) {
	pr, err := s.prs.GetByID(ctx, body.PullRequestId)
	if err != nil {
		if errors.Is(err, repo.ErrPRNotFound) {
			return nil, "", ErrPRNotFound
		}
		return nil, "", err
	}

	if pr.Status == api.PullRequestStatusMERGED {
		return nil, "", ErrPRMerged
	}

	idx := slices.Index(pr.AssignedReviewers, body.OldUserId)
	if idx == -1 {
		return nil, "", ErrReviewerNotAssigned
	}

	oldReviewer, err := s.users.Get(ctx, body.OldUserId)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}

	team, err := s.teams.GetTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		if errors.Is(err, repo.ErrTeamNotFound) {
			return nil, "", ErrTeamNotFound
		}
		return nil, "", err
	}

	var candidates []string
	for _, m := range team.Members {
		if !m.IsActive {
			continue
		}
		if m.UserId == oldReviewer.UserId {
			continue
		}
		if m.UserId == pr.AuthorId {
			continue
		}
		if slices.Contains(pr.AssignedReviewers, m.UserId) {
			continue
		}
		candidates = append(candidates, m.UserId)
	}

	if len(candidates) == 0 {
		return nil, "", ErrNoCandidate
	}

	newReviewerID := candidates[0]

	updatedPR, err := s.prs.ReassignReviewer(ctx, body.PullRequestId, body.OldUserId, newReviewerID)
	if err != nil {
		if errors.Is(err, repo.ErrPRNotFound) {
			return nil, "", ErrPRNotFound
		}
		return nil, "", err
	}

	return updatedPR, newReviewerID, nil
}

func (s *PRService) GetCountPRs(ctx context.Context) (total int, open int, merged int, err error) {
	return s.prs.CountPRs(ctx)
}

func (s *PRService) GetAllUsersWithAssignmentCounts(ctx context.Context) ([]struct {
	UserID      string `json:"user_id"`
	Assignments int    `json:"assignments"`
}, error) {
	return s.prs.GetAllUsersWithAssignmentCounts(ctx)
}
