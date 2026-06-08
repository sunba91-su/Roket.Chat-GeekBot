package store

import (
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS teams (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		channel_id TEXT NOT NULL,
		lead_id TEXT,
		schedule_cron TEXT,
		questions TEXT NOT NULL DEFAULT 'How are you feeling?|What did you do yesterday?|What are you doing today?|Any blockers?',
		timezone TEXT NOT NULL DEFAULT 'UTC',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS team_members (
		id TEXT PRIMARY KEY,
		team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL,
		username TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'member',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(team_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS standup_sessions (
		id TEXT PRIMARY KEY,
		team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
		date TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'open',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS standup_responses (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES standup_sessions(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL,
		username TEXT NOT NULL,
		answers TEXT NOT NULL,
		submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(session_id, user_id)
	);
	`

	_, err := s.db.Exec(schema)
	return err
}
