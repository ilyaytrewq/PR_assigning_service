package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	api "ilyaytrewq/PR_assigning_service/internal/api"

	"github.com/lib/pq"
)

// PRRepo manages pull request data.
type PRRepo struct {
	db *sql.DB
}

// NewPRRepo creates a new PRRepo.
func NewPRRepo(db *sql.DB) *PRRepo {
	return &PRRepo{db: db}
}

// CreatePR creates a new pull request.
func (r *PRRepo) CreatePR(ctx context.Context, pr *api.PullRequest) error {
	const query = `
        INSERT INTO pull_requests (
            pull_request_id,
            pull_request_name,
            author_id,
            status,
            assigned_reviewers,
            created_at,
            merged_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (pull_request_id) DO NOTHING;
    `

	res, err := r.db.ExecContext(ctx, query,
		pr.PullRequestId,
		pr.PullRequestName,
		pr.AuthorId,
		pr.Status,
		pq.Array(pr.AssignedReviewers),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return fmt.Errorf("insert pr id=%s name=%s failed: %w",
			pr.PullRequestId, pr.PullRequestName, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("insert pr id=%s name=%s: rows affected failed: %w",
			pr.PullRequestId, pr.PullRequestName, err)
	}

	if rows == 0 {
		return ErrPRExists
	}

	return nil
}

// MergePR marks a PR as merged.
func (r *PRRepo) MergePR(ctx context.Context, prID string, mergedAt time.Time) (*api.PullRequest, error) {
	const query = `
        UPDATE pull_requests
        SET
            status    = 'MERGED',
            merged_at = COALESCE(merged_at, $2)
        WHERE pull_request_id = $1
        RETURNING
            pull_request_id,
            pull_request_name,
            author_id,
            status,
            assigned_reviewers,
            created_at,
            merged_at;
    `

	var pr api.PullRequest

	row := r.db.QueryRowContext(ctx, query, prID, mergedAt)

	err := row.Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.Status,
		pq.Array(&pr.AssignedReviewers),
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPRNotFound
		}
		return nil, fmt.Errorf("merge pr id=%s failed: %w", prID, err)
	}

	return &pr, nil
}

// ReassignReviewer replaces a reviewer for a PR.
func (r *PRRepo) ReassignReviewer(
	ctx context.Context,
	prID string,
	oldReviewer string,
	newReviewer string,
) (*api.PullRequest, error) {

	const reviewersQuery = `
        SELECT assigned_reviewers
        FROM pull_requests
        WHERE pull_request_id = $1
    `

	var reviewers []string

	err := r.db.QueryRowContext(ctx, reviewersQuery, prID).Scan(
		pq.Array(&reviewers),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPRNotFound
		}
		return nil, fmt.Errorf("reassign reviewer: get reviewers pr=%s: %w", prID, err)
	}

	idx := slices.Index(reviewers, oldReviewer)
	if idx == -1 {
		return nil, ErrUserNotFound
	}

	reviewers[idx] = newReviewer

	const updateQuery = `
        UPDATE pull_requests
        SET assigned_reviewers = $2
        WHERE pull_request_id = $1
        RETURNING
            pull_request_id,
            pull_request_name,
            author_id,
            status,
            assigned_reviewers,
            created_at,
            merged_at;
    `

	var pr api.PullRequest

	err = r.db.QueryRowContext(
		ctx,
		updateQuery,
		prID,
		pq.Array(reviewers),
	).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.Status,
		pq.Array(&pr.AssignedReviewers),
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("reassign reviewer: update pr=%s failed: %w", prID, err)
	}

	return &pr, nil
}

// GetByID retrieves a PR by its ID.
func (r *PRRepo) GetByID(ctx context.Context, prID string) (*api.PullRequest, error) {
	const query = `
        SELECT
            pull_request_id,
            pull_request_name,
            author_id,
            status,
            assigned_reviewers,
            created_at,
            merged_at
        FROM pull_requests
        WHERE pull_request_id = $1;
    `

	var pr api.PullRequest

	err := r.db.QueryRowContext(ctx, query, prID).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.Status,
		pq.Array(&pr.AssignedReviewers),
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPRNotFound
		}
		return nil, fmt.Errorf("get pr id=%s failed: %w", prID, err)
	}

	return &pr, nil
}

// GetByReviewer retrieves PRs assigned to a reviewer.
func (r *PRRepo) GetByReviewer(ctx context.Context, userID string) ([]*api.PullRequest, error) {
	const query = `
        SELECT
            pull_request_id,
            pull_request_name,
            author_id,
            status,
            assigned_reviewers,
            created_at,
            merged_at
        FROM pull_requests
        WHERE $1 = ANY (assigned_reviewers)
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get PRs by reviewer %s failed: %w", userID, err)
	}
	defer func() { _ = rows.Close() }()

	var result []*api.PullRequest

	for rows.Next() {
		var pr api.PullRequest

		err := rows.Scan(
			&pr.PullRequestId,
			&pr.PullRequestName,
			&pr.AuthorId,
			&pr.Status,
			pq.Array(&pr.AssignedReviewers),
			&pr.CreatedAt,
			&pr.MergedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan PR reviewer=%s failed: %w", userID, err)
		}

		result = append(result, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return result, nil
}

// CountPRs returns PR statistics.
func (r *PRRepo) CountPRs(ctx context.Context) (total int, open int, merged int, err error) {
	const query = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'OPEN') AS open,
			COUNT(*) FILTER (WHERE status = 'MERGED') AS merged
		FROM pull_requests;
	`

	err = r.db.QueryRowContext(ctx, query).Scan(&total, &open, &merged)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("count PRs failed: %w", err)
	}
	return total, open, merged, nil
}

// GetAllUsersWithAssignmentCounts returns assignment counts for all users.
func (r *PRRepo) GetAllUsersWithAssignmentCounts(ctx context.Context) ([]struct {
	UserID      string `json:"user_id"`
	Assignments int    `json:"assignments"`
}, error) {
	const query = `
		SELECT user_id, COUNT(*) as assignments
		FROM pull_requests, unnest(assigned_reviewers) AS user_id
		GROUP BY user_id;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all users with assignment counts failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []struct {
		UserID      string `json:"user_id"`
		Assignments int    `json:"assignments"`
	}

	for rows.Next() {
		var user struct {
			UserID      string `json:"user_id"`
			Assignments int    `json:"assignments"`
		}
		if err := rows.Scan(&user.UserID, &user.Assignments); err != nil {
			return nil, fmt.Errorf("scan user assignment count failed: %w", err)
		}
		result = append(result, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}
	return result, nil
}
