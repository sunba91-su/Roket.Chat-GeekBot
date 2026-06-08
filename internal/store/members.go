package store

import (
	"fmt"
	"strings"
)

type Member struct {
	ID        string `json:"id"`
	TeamID    string `json:"team_id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
}

func (s *Store) AddMember(m *Member) error {
	_, err := s.db.Exec(
		`INSERT INTO team_members (id, team_id, user_id, username, role)
		 VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.TeamID, m.UserID, m.Username, m.Role,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("user already in team")
		}
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (s *Store) RemoveMember(teamID, userID string) error {
	res, err := s.db.Exec(
		`DELETE FROM team_members WHERE team_id = ? AND user_id = ?`, teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user not found in team")
	}
	return nil
}

func (s *Store) GetMembers(teamID string) ([]*Member, error) {
	rows, err := s.db.Query(
		`SELECT id, team_id, user_id, username, role
		 FROM team_members WHERE team_id = ? ORDER BY username`, teamID,
	)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	var members []*Member
	for rows.Next() {
		m := &Member{}
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.Username, &m.Role); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *Store) IsMember(teamID, userID string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM team_members WHERE team_id = ? AND user_id = ?`,
		teamID, userID,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) IsTeamLead(teamID, userID string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM team_members WHERE team_id = ? AND user_id = ? AND role = 'lead'`,
		teamID, userID,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) SetRole(teamID, userID, role string) error {
	_, err := s.db.Exec(
		`UPDATE team_members SET role = ? WHERE team_id = ? AND user_id = ?`,
		role, teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("set role: %w", err)
	}
	return nil
}

func (s *Store) GetTeamLead(teamID string) (*Member, error) {
	m := &Member{}
	err := s.db.QueryRow(
		`SELECT id, team_id, user_id, username, role
		 FROM team_members WHERE team_id = ? AND role = 'lead' LIMIT 1`, teamID,
	).Scan(&m.ID, &m.TeamID, &m.UserID, &m.Username, &m.Role)
	if err != nil {
		return nil, fmt.Errorf("get team lead: %w", err)
	}
	return m, nil
}
