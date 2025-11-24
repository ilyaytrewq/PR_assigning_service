package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ilyaytrewq/PR_assigning_service/internal/api"
)

type TeamRepo struct {
	db *sql.DB
}

func NewTeamRepo(db *sql.DB) *TeamRepo {
	return &TeamRepo{
		db: db,
	}
}

func (tr *TeamRepo) InsertTeam(ctx context.Context, team *api.Team) error {
	const query = `
        INSERT INTO teams (team_name)
        VALUES ($1)
        ON CONFLICT (team_name) DO NOTHING;
    `

	res, err := tr.db.ExecContext(ctx, query, team.TeamName)
	if err != nil {
		return fmt.Errorf("insert team %s failed: %w", team.TeamName, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("insert team %s: rows affected: %w", team.TeamName, err)
	}

	if rows == 0 {
		return ErrTeamExists
	}

	return nil
}

func (tr *TeamRepo) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	const teamQuery = `
        SELECT team_name FROM teams WHERE team_name = $1
    `
	var team api.Team

	err := tr.db.QueryRowContext(ctx, teamQuery, teamName).Scan(&team.TeamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, fmt.Errorf("get team %s: %w", teamName, err)
	}

	const membersQuery = `
        SELECT user_id, username, is_active 
        FROM users 
        WHERE team_name = $1
    `
	rows, err := tr.db.QueryContext(ctx, membersQuery, team.TeamName)
	if err != nil {
		return nil, fmt.Errorf("get team %s members query failed: %w", teamName, err)
	}
	defer rows.Close()

	team.Members = []api.TeamMember{}
	for rows.Next() {
		var member api.TeamMember
		if err := rows.Scan(&member.UserId, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("scan team %s member: %w", teamName, err)
		}
		team.Members = append(team.Members, member)
	}

	return &team, nil
}
