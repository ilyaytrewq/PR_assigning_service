CREATE TABLE IF NOT EXISTS teams (
    team_name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
    user_id   TEXT PRIMARY KEY,
    username  TEXT NOT NULL,
    team_name TEXT NOT NULL REFERENCES teams(team_name)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id         TEXT NOT NULL REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    status            TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    assigned_reviewers TEXT[] NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ,
    merged_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pull_requests_assigned_reviewers
    ON pull_requests
    USING GIN (assigned_reviewers);

CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id
    ON pull_requests (author_id);
