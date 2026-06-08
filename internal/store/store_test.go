package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	s, err := store.New(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		s.Close()
		os.Remove(f.Name())
	})

	return s
}

func TestCreateAndGetTeam(t *testing.T) {
	s := newTestStore(t)

	err := s.CreateTeam(&store.Team{
		ID:        "team-1",
		Name:      "Engineering",
		ChannelID: "GENERAL",
		Questions: "Feel?|Yesterday?|Today?|Blockers?",
		Timezone:  "UTC",
	})
	if err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	team, err := s.GetTeam("team-1")
	if err != nil {
		t.Fatalf("GetTeam: %v", err)
	}
	if team.Name != "Engineering" {
		t.Errorf("expected Engineering, got %s", team.Name)
	}
	if team.Questions != "Feel?|Yesterday?|Today?|Blockers?" {
		t.Errorf("unexpected questions: %s", team.Questions)
	}
}

func TestGetTeamByName(t *testing.T) {
	s := newTestStore(t)

	_ = s.CreateTeam(&store.Team{
		ID: "t1", Name: "Design", ChannelID: "DESIGN",
	})

	team, err := s.GetTeamByName("Design")
	if err != nil {
		t.Fatalf("GetTeamByName: %v", err)
	}
	if team.ID != "t1" {
		t.Errorf("expected t1, got %s", team.ID)
	}
}

func TestDuplicateTeam(t *testing.T) {
	s := newTestStore(t)

	_ = s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "CH"})
	err := s.CreateTeam(&store.Team{ID: "t2", Name: "Eng", ChannelID: "CH"})
	if err == nil {
		t.Fatal("expected error for duplicate team")
	}
}

func TestAddAndRemoveMember(t *testing.T) {
	s := newTestStore(t)
	_ = s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "CH"})

	m := &store.Member{ID: "m1", TeamID: "t1", UserID: "user1", Username: "alice", Role: "member"}
	if err := s.AddMember(m); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	isMember, _ := s.IsMember("t1", "user1")
	if !isMember {
		t.Error("expected user1 to be member")
	}

	if err := s.RemoveMember("t1", "user1"); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}

	isMember, _ = s.IsMember("t1", "user1")
	if isMember {
		t.Error("expected user1 to not be member")
	}
}

func TestTeamLead(t *testing.T) {
	s := newTestStore(t)
	_ = s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "CH"})
	_ = s.AddMember(&store.Member{ID: "m1", TeamID: "t1", UserID: "u1", Username: "alice", Role: "lead"})

	lead, err := s.GetTeamLead("t1")
	if err != nil {
		t.Fatalf("GetTeamLead: %v", err)
	}
	if lead.Username != "alice" {
		t.Errorf("expected alice, got %s", lead.Username)
	}

	isLead, _ := s.IsTeamLead("t1", "u1")
	if !isLead {
		t.Error("expected u1 to be lead")
	}
}

func TestStandupSession(t *testing.T) {
	s := newTestStore(t)
	_ = s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "CH"})

	now := time.Now().Format("2006-01-02")
	session := &store.StandupSession{ID: "s1", TeamID: "t1", Date: now, Status: "open"}
	if err := s.CreateSession(session); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	active, err := s.GetActiveSession("t1", now)
	if err != nil {
		t.Fatalf("GetActiveSession: %v", err)
	}
	if active.ID != "s1" {
		t.Errorf("expected s1, got %s", active.ID)
	}

	if err := s.CloseSession("s1"); err != nil {
		t.Fatalf("CloseSession: %v", err)
	}

	_, err = s.GetActiveSession("t1", now)
	if err == nil {
		t.Error("expected error for closed session")
	}
}

func TestSubmitResponse(t *testing.T) {
	s := newTestStore(t)
	_ = s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "CH"})
	_ = s.CreateSession(&store.StandupSession{ID: "s1", TeamID: "t1", Date: "2026-06-08", Status: "open"})

	resp := &store.StandupResponse{
		ID: "r1", SessionID: "s1", UserID: "u1",
		Username: "alice", Answers: "Great|Fixed bugs|Working on search|No blockers",
		SubmittedAt: time.Now(),
	}
	if err := s.SubmitResponse(resp); err != nil {
		t.Fatalf("SubmitResponse: %v", err)
	}

	got, err := s.GetResponse("s1", "u1")
	if err != nil {
		t.Fatalf("GetResponse: %v", err)
	}
	if got.Answers != "Great|Fixed bugs|Working on search|No blockers" {
		t.Errorf("unexpected answers: %s", got.Answers)
	}
}

func TestUpdateResponse(t *testing.T) {
	s := newTestStore(t)
	_ = s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "CH"})
	_ = s.CreateSession(&store.StandupSession{ID: "s1", TeamID: "t1", Date: "2026-06-08", Status: "open"})

	_ = s.SubmitResponse(&store.StandupResponse{
		ID: "r1", SessionID: "s1", UserID: "u1", Username: "alice",
		Answers: "Initial", SubmittedAt: time.Now(),
	})

	err := s.SubmitResponse(&store.StandupResponse{
		ID: "r2", SessionID: "s1", UserID: "u1", Username: "alice",
		Answers: "Updated|Answers", SubmittedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Update response: %v", err)
	}

	resp, _ := s.GetResponse("s1", "u1")
	if resp.Answers != "Updated|Answers" {
		t.Errorf("expected updated answers, got %s", resp.Answers)
	}
}

func TestGetTeamsForUser(t *testing.T) {
	s := newTestStore(t)
	_ = s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "CH"})
	_ = s.CreateTeam(&store.Team{ID: "t2", Name: "Design", ChannelID: "DS"})
	_ = s.AddMember(&store.Member{ID: "m1", TeamID: "t1", UserID: "u1", Username: "alice"})
	_ = s.AddMember(&store.Member{ID: "m2", TeamID: "t2", UserID: "u1", Username: "alice"})

	teams, err := s.GetTeamsForUser("u1")
	if err != nil {
		t.Fatalf("GetTeamsForUser: %v", err)
	}
	if len(teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(teams))
	}
}
