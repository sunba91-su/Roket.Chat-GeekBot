package store

import (
	"fmt"
	"strings"
)

type Team struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ChannelID    string `json:"channel_id"`
	LeadID       string `json:"lead_id,omitempty"`
	ScheduleCron string `json:"schedule_cron,omitempty"`
	Questions    string `json:"questions"`
	Timezone     string `json:"timezone"`
}

func (s *Store) CreateTeam(t *Team) error {
	_, err := s.db.Exec(
		`INSERT INTO teams (id, name, channel_id, lead_id, schedule_cron, questions, timezone)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, t.ChannelID, t.LeadID, t.ScheduleCron, t.Questions, t.Timezone,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("team %q already exists", t.Name)
		}
		return fmt.Errorf("create team: %w", err)
	}
	return nil
}

func (s *Store) GetTeam(id string) (*Team, error) {
	t := &Team{}
	err := s.db.QueryRow(
		`SELECT id, name, channel_id, COALESCE(lead_id,''), COALESCE(schedule_cron,''), questions, timezone
		 FROM teams WHERE id = ?`, id,
	).Scan(&t.ID, &t.Name, &t.ChannelID, &t.LeadID, &t.ScheduleCron, &t.Questions, &t.Timezone)
	if err != nil {
		return nil, fmt.Errorf("get team: %w", err)
	}
	return t, nil
}

func (s *Store) GetTeamByName(name string) (*Team, error) {
	t := &Team{}
	err := s.db.QueryRow(
		`SELECT id, name, channel_id, COALESCE(lead_id,''), COALESCE(schedule_cron,''), questions, timezone
		 FROM teams WHERE name = ?`, name,
	).Scan(&t.ID, &t.Name, &t.ChannelID, &t.LeadID, &t.ScheduleCron, &t.Questions, &t.Timezone)
	if err != nil {
		return nil, fmt.Errorf("get team by name: %w", err)
	}
	return t, nil
}

func (s *Store) ListTeams() ([]*Team, error) {
	rows, err := s.db.Query(
		`SELECT id, name, channel_id, COALESCE(lead_id,''), COALESCE(schedule_cron,''), questions, timezone
		 FROM teams ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		t := &Team{}
		if err := rows.Scan(&t.ID, &t.Name, &t.ChannelID, &t.LeadID, &t.ScheduleCron, &t.Questions, &t.Timezone); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (s *Store) UpdateTeam(t *Team) error {
	_, err := s.db.Exec(
		`UPDATE teams SET name=?, channel_id=?, lead_id=?, schedule_cron=?, questions=?, timezone=?
		 WHERE id=?`,
		t.Name, t.ChannelID, t.LeadID, t.ScheduleCron, t.Questions, t.Timezone, t.ID,
	)
	if err != nil {
		return fmt.Errorf("update team: %w", err)
	}
	return nil
}

func (s *Store) DeleteTeam(id string) error {
	_, err := s.db.Exec(`DELETE FROM teams WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete team: %w", err)
	}
	return nil
}

func (s *Store) GetTeamsForUser(userID string) ([]*Team, error) {
	rows, err := s.db.Query(
		`SELECT t.id, t.name, t.channel_id, COALESCE(t.lead_id,''), COALESCE(t.schedule_cron,''), t.questions, t.timezone
		 FROM teams t
		 JOIN team_members m ON m.team_id = t.id
		 WHERE m.user_id = ?
		 ORDER BY t.name`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get teams for user: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		t := &Team{}
		if err := rows.Scan(&t.ID, &t.Name, &t.ChannelID, &t.LeadID, &t.ScheduleCron, &t.Questions, &t.Timezone); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}
