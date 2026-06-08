package store

import (
	"fmt"
	"time"
)

type StandupSession struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	Date      string    `json:"date"`
	Status    string    `json:"status"`
}

type StandupResponse struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Answers     string    `json:"answers"`
	SubmittedAt time.Time `json:"submitted_at"`
}

func (s *Store) CreateSession(session *StandupSession) error {
	_, err := s.db.Exec(
		`INSERT INTO standup_sessions (id, team_id, date, status) VALUES (?, ?, ?, ?)`,
		session.ID, session.TeamID, session.Date, session.Status,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (s *Store) GetSession(id string) (*StandupSession, error) {
	session := &StandupSession{}
	err := s.db.QueryRow(
		`SELECT id, team_id, date, status FROM standup_sessions WHERE id = ?`, id,
	).Scan(&session.ID, &session.TeamID, &session.Date, &session.Status)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return session, nil
}

func (s *Store) GetActiveSession(teamID string, date string) (*StandupSession, error) {
	session := &StandupSession{}
	err := s.db.QueryRow(
		`SELECT id, team_id, date, status FROM standup_sessions
		 WHERE team_id = ? AND date = ? AND status = 'open'`, teamID, date,
	).Scan(&session.ID, &session.TeamID, &session.Date, &session.Status)
	if err != nil {
		return nil, fmt.Errorf("get active session: %w", err)
	}
	return session, nil
}

func (s *Store) CloseSession(id string) error {
	_, err := s.db.Exec(
		`UPDATE standup_sessions SET status = 'closed' WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("close session: %w", err)
	}
	return nil
}

func (s *Store) SubmitResponse(resp *StandupResponse) error {
	_, err := s.db.Exec(
		`INSERT INTO standup_responses (id, session_id, user_id, username, answers, submitted_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(session_id, user_id) DO UPDATE SET answers=excluded.answers, submitted_at=excluded.submitted_at`,
		resp.ID, resp.SessionID, resp.UserID, resp.Username, resp.Answers, resp.SubmittedAt,
	)
	if err != nil {
		return fmt.Errorf("submit response: %w", err)
	}
	return nil
}

func (s *Store) GetResponse(sessionID, userID string) (*StandupResponse, error) {
	r := &StandupResponse{}
	err := s.db.QueryRow(
		`SELECT id, session_id, user_id, username, answers, submitted_at
		 FROM standup_responses WHERE session_id = ? AND user_id = ?`, sessionID, userID,
	).Scan(&r.ID, &r.SessionID, &r.UserID, &r.Username, &r.Answers, &r.SubmittedAt)
	if err != nil {
		return nil, fmt.Errorf("get response: %w", err)
	}
	return r, nil
}

func (s *Store) GetResponses(sessionID string) ([]*StandupResponse, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, user_id, username, answers, submitted_at
		 FROM standup_responses WHERE session_id = ? ORDER BY username`, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get responses: %w", err)
	}
	defer rows.Close()

	var responses []*StandupResponse
	for rows.Next() {
		r := &StandupResponse{}
		if err := rows.Scan(&r.ID, &r.SessionID, &r.UserID, &r.Username, &r.Answers, &r.SubmittedAt); err != nil {
			return nil, fmt.Errorf("scan response: %w", err)
		}
		responses = append(responses, r)
	}
	return responses, rows.Err()
}

func (s *Store) HasSubmitted(sessionID, userID string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM standup_responses WHERE session_id = ? AND user_id = ?`,
		sessionID, userID,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) GetSessionStatus(teamID, date string) (submitted int, total int, err error) {
	row := s.db.QueryRow(
		`SELECT COUNT(*) FROM standup_responses r
		 JOIN standup_sessions sess ON sess.id = r.session_id
		 WHERE sess.team_id = ? AND sess.date = ?`, teamID, date,
	)
	if err := row.Scan(&submitted); err != nil {
		return 0, 0, fmt.Errorf("count submitted: %w", err)
	}

	row = s.db.QueryRow(
		`SELECT COUNT(*) FROM team_members WHERE team_id = ?`, teamID,
	)
	if err := row.Scan(&total); err != nil {
		return 0, 0, fmt.Errorf("count members: %w", err)
	}

	return submitted, total, nil
}

func (s *Store) ListSessions(teamID string, limit int) ([]*StandupSession, error) {
	rows, err := s.db.Query(
		`SELECT id, team_id, date, status FROM standup_sessions
		 WHERE team_id = ? ORDER BY created_at DESC LIMIT ?`, teamID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*StandupSession
	for rows.Next() {
		session := &StandupSession{}
		if err := rows.Scan(&session.ID, &session.TeamID, &session.Date, &session.Status); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}
