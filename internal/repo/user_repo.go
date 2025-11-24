package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ilyaytrewq/PR_assigning_service/internal/api"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) InsertOrUpdate(ctx context.Context, user *api.User) error {
	const query = `
        INSERT INTO users (user_id, username, team_name, is_active)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id)
        DO UPDATE SET
            username  = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active;
    `

	_, err := ur.db.ExecContext(ctx, query,
		user.UserId,
		user.Username,
		user.TeamName,
		user.IsActive,
	)

	if err != nil {
		return fmt.Errorf("user insert/update (%s) failed: %w", user.UserId, err)
	}

	return nil
}

func (ur *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	const query = `
		UPDATE users
		SET is_active = $2
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active
	`

	var u api.User

	err := ur.db.
		QueryRowContext(ctx, query, userID, isActive).
		Scan(&u.UserId, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("set is_active for user=%s failed: %w", userID, err)
	}

	return &u, nil
}

func (ur *UserRepository) Get(ctx context.Context, userId string) (*api.User, error) {
	const query = `
        SELECT user_id, username, team_name, is_active
        FROM users
        WHERE user_id = $1;
    `

	var user api.User

	err := ur.db.QueryRowContext(ctx, query, userId).Scan(
		&user.UserId,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user(%s) failed: %w", userId, err)
	}

	return &user, nil
}

func (ur *UserRepository) CountUsersAndActive(ctx context.Context) (total int, active int, err error) {
	const query = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE is_active) AS active
		FROM users;
	`

	err = ur.db.QueryRowContext(ctx, query).Scan(&total, &active)
	if err != nil {
		return 0, 0, fmt.Errorf("count users failed: %w", err)
	}

	return total, active, nil
}