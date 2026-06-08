package commands_test

import (
	"os"
	"testing"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/commands"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/config"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/store"
)

type mockUserProvider struct{}

func (m *mockUserProvider) UserInfo(username string) (*commands.UserInfo, error) {
	return &commands.UserInfo{
		ID:       username + "-id",
		Username: username,
		Name:     username,
	}, nil
}

type teamTestHarness struct {
	t      *testing.T
	s      *store.Store
	msgr   *mockMessenger
	reg    *commands.Registry
	admin  *commands.Context
	lead   *commands.Context
	member *commands.Context
}

func newTeamTestHarness(t *testing.T) *teamTestHarness {
	t.Helper()

	f, err := os.CreateTemp("", "test-team-*.db")
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

	reg := commands.New()
	commands.RegisterTeamCommands(reg)

	cfg := &config.Config{MainAdmin: "admin"}

	makeCtx := func(username, userID string) *commands.Context {
		return &commands.Context{
			Username:     username,
			UserID:       userID,
			RoomID:       "GENERAL",
			Store:        s,
			Messenger:    &mockMessenger{},
			UserProvider: &mockUserProvider{},
			Config:       cfg,
		}
	}

	return &teamTestHarness{
		t:      t,
		s:      s,
		msgr:   &mockMessenger{},
		reg:    reg,
		admin:  makeCtx("admin", "admin-id"),
		lead:   makeCtx("alice", "alice-id"),
		member: makeCtx("bob", "bob-id"),
	}
}

func (h *teamTestHarness) dispatch(ctx *commands.Context, text string) {
	ctx.RawText = text
	ctx.Messenger = &mockMessenger{}
	_, _ = h.reg.Dispatch(ctx)
}

func TestTeamCreateAdminOnly(t *testing.T) {
	h := newTeamTestHarness(t)

	h.dispatch(h.member, "/standup team create Eng --channel #eng")

	if _, err := h.s.GetTeamByName("Eng"); err == nil {
		t.Error("non-admin should not be able to create team")
	}
}

func TestTeamCreateSuccess(t *testing.T) {
	h := newTeamTestHarness(t)

	h.dispatch(h.admin, "/standup team create Engineering --channel #engineering")

	team, err := h.s.GetTeamByName("Engineering")
	if err != nil {
		t.Fatalf("team should exist: %v", err)
	}
	if team.ChannelID != "engineering" {
		t.Errorf("expected channel engineering, got %s", team.ChannelID)
	}
	if team.Questions == "" {
		t.Error("expected default questions")
	}
}

func TestTeamCreateWithoutChannel(t *testing.T) {
	h := newTeamTestHarness(t)

	h.dispatch(h.admin, "/standup team create My-Team")

	team, err := h.s.GetTeamByName("My-Team")
	if err != nil {
		t.Fatalf("team should exist: %v", err)
	}
	if team.ChannelID != "my-team" {
		t.Errorf("expected channel my-team, got %s", team.ChannelID)
	}
}

func TestTeamSetLeadCreatesMember(t *testing.T) {
	h := newTeamTestHarness(t)

	_ = h.s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "eng"})

	h.dispatch(h.admin, "/standup team set-lead Eng alice")

	isLead, _ := h.s.IsTeamLead("t1", "alice-id")
	if !isLead {
		t.Error("alice should be lead of Eng")
	}
}

func TestTeamSetLeadUpdatesRole(t *testing.T) {
	h := newTeamTestHarness(t)

	_ = h.s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "eng"})
	_ = h.s.AddMember(&store.Member{ID: "m1", TeamID: "t1", UserID: "alice-id", Username: "alice", Role: "member"})

	h.dispatch(h.admin, "/standup team set-lead Eng alice")

	isLead, _ := h.s.IsTeamLead("t1", "alice-id")
	if !isLead {
		t.Error("alice should now be lead")
	}
}

func TestTeamAddAndList(t *testing.T) {
	h := newTeamTestHarness(t)

	_ = h.s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "eng"})
	_ = h.s.AddMember(&store.Member{ID: "m1", TeamID: "t1", UserID: "alice-id", Username: "alice", Role: "lead"})

	h.dispatch(h.lead, "/standup team add bob")

	members, _ := h.s.GetMembers("t1")
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}

	h.lead.Messenger = &mockMessenger{}
	h.dispatch(h.lead, "/standup team members")

	found := false
	for _, msg := range h.lead.Messenger.(*mockMessenger).sent {
		if contains(msg, "bob") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected bob in members output")
	}
}

func TestTeamRemove(t *testing.T) {
	h := newTeamTestHarness(t)

	_ = h.s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "eng"})
	_ = h.s.AddMember(&store.Member{ID: "m1", TeamID: "t1", UserID: "alice-id", Username: "alice", Role: "lead"})
	_ = h.s.AddMember(&store.Member{ID: "m2", TeamID: "t1", UserID: "bob-id", Username: "bob", Role: "member"})

	h.dispatch(h.lead, "/standup team remove bob")

	members, _ := h.s.GetMembers("t1")
	if len(members) != 1 {
		t.Errorf("expected 1 member after removal, got %d", len(members))
	}
}

func TestTeamSetSchedule(t *testing.T) {
	h := newTeamTestHarness(t)

	_ = h.s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "eng", Questions: "Q1|Q2", Timezone: "UTC"})
	_ = h.s.AddMember(&store.Member{ID: "m1", TeamID: "t1", UserID: "alice-id", Username: "alice", Role: "lead"})

	h.dispatch(h.lead, "/standup team set schedule 0 9 * * 1-5")

	team, _ := h.s.GetTeam("t1")
	if team.ScheduleCron != "0 9 * * 1-5" {
		t.Errorf("expected cron '0 9 * * 1-5', got %s", team.ScheduleCron)
	}
}

func TestTeamSetChannel(t *testing.T) {
	h := newTeamTestHarness(t)

	_ = h.s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "eng", Questions: "Q1|Q2", Timezone: "UTC"})
	_ = h.s.AddMember(&store.Member{ID: "m1", TeamID: "t1", UserID: "alice-id", Username: "alice", Role: "lead"})

	h.dispatch(h.lead, "/standup team set channel #rocket")

	team, _ := h.s.GetTeam("t1")
	if team.ChannelID != "rocket" {
		t.Errorf("expected channel 'rocket', got %s", team.ChannelID)
	}
}

func TestTeamSetTimezone(t *testing.T) {
	h := newTeamTestHarness(t)

	_ = h.s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "eng", Questions: "Q1|Q2", Timezone: "UTC"})
	_ = h.s.AddMember(&store.Member{ID: "m1", TeamID: "t1", UserID: "alice-id", Username: "alice", Role: "lead"})

	h.dispatch(h.lead, "/standup team set timezone America/New_York")

	team, _ := h.s.GetTeam("t1")
	if team.Timezone != "America/New_York" {
		t.Errorf("expected 'America/New_York', got %s", team.Timezone)
	}
}

func TestTeamDelete(t *testing.T) {
	h := newTeamTestHarness(t)

	_ = h.s.CreateTeam(&store.Team{ID: "t1", Name: "Eng", ChannelID: "eng"})

	h.dispatch(h.admin, "/standup team delete Eng")

	if _, err := h.s.GetTeam("t1"); err == nil {
		t.Error("team should be deleted")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
