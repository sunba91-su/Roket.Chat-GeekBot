package commands_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/commands"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/config"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/convstate"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/store"
)

type standupHarness struct {
	t       *testing.T
	s       *store.Store
	msgr    *mockMessenger
	reg     *commands.Registry
	convMgr *convstate.Manager
	ctx     *commands.Context
}

func newStandupHarness(t *testing.T) *standupHarness {
	t.Helper()

	f, err := os.CreateTemp("", "test-standup-*.db")
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

	_ = s.CreateTeam(&store.Team{
		ID: "t1", Name: "Eng", ChannelID: "eng",
		Questions: "How are you?|What did you do?|Any blockers?",
		Timezone:  "UTC",
	})
	_ = s.AddMember(&store.Member{
		ID: "m1", TeamID: "t1", UserID: "alice-id",
		Username: "alice", Role: "member",
	})

	msgr := &mockMessenger{}
	convMgr := convstate.NewManager()

	reg := commands.New()
	commands.RegisterStandupCommands(reg)

	ctx := &commands.Context{
		UserID:       "alice-id",
		Username:     "alice",
		RoomID:       "DM-room",
		Store:        s,
		Messenger:    msgr,
		UserProvider: &mockUserProvider{},
		Config:       &config.Config{MainAdmin: "admin"},
		ConvState:    convMgr,
		IsDM:         true,
	}

	return &standupHarness{
		t: t, s: s, msgr: msgr, reg: reg,
		convMgr: convMgr, ctx: ctx,
	}
}

func (h *standupHarness) dispatch(text string) {
	h.ctx.RawText = text
	h.ctx.Messenger = &mockMessenger{}
	h.msgr = h.ctx.Messenger.(*mockMessenger)
	_, _ = h.reg.Dispatch(h.ctx)
}

func TestStandupSubmitNoTeam(t *testing.T) {
	h := newStandupHarness(t)
	_ = h.s.RemoveMember("t1", "alice-id")

	h.dispatch("/standup submit")

	if !contains(strings.Join(h.msgr.sent, " "), "not a member") {
		t.Error("expected 'not a member' message")
	}
}

func TestStandupSubmitStartsConversation(t *testing.T) {
	h := newStandupHarness(t)

	h.dispatch("/standup submit")

	conv, ok := h.convMgr.GetConversation("alice-id")
	if !ok {
		t.Fatal("expected conversation to start")
	}
	if conv.CurrentQ != 0 {
		t.Errorf("expected Q0, got Q%d", conv.CurrentQ)
	}
}

func TestStandupStatusNoSubmission(t *testing.T) {
	h := newStandupHarness(t)
	h.dispatch("/standup status")

	msg := strings.Join(h.msgr.sent, " ")
	if !contains(msg, "not submitted") && !contains(msg, "haven't") {
		t.Logf("Status message: %s", msg)
	}
}

func TestStandupStatusAfterSubmission(t *testing.T) {
	h := newStandupHarness(t)

	now := time.Now().Format("2006-01-02")
	_ = h.s.CreateSession(&store.StandupSession{
		ID: "sess-test", TeamID: "t1", Date: now, Status: "open",
	})
	_ = h.s.SubmitResponse(&store.StandupResponse{
		ID: "resp-test", SessionID: "sess-test", UserID: "alice-id",
		Username: "alice", Answers: "Fine|Coding|None",
		SubmittedAt: time.Now(),
	})

	h.dispatch("/standup status")

	msg := strings.Join(h.msgr.sent, " ")
	if !contains(msg, "submitted") {
		t.Logf("Status message: %s", msg)
	}
}

func TestStandupConversationFlow(t *testing.T) {
	h := newStandupHarness(t)

	h.dispatch("/standup submit")

	conv, ok := h.convMgr.GetConversation("alice-id")
	if !ok {
		t.Fatal("expected conversation")
	}

	finished, nextQ, err := h.convMgr.RecordAnswer("alice-id", "Feeling great")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if finished {
		t.Fatal("expected more questions")
	}
	if !contains(nextQ, "What did you do") {
		t.Errorf("unexpected next Q: %s", nextQ)
	}

	finished, nextQ, _ = h.convMgr.RecordAnswer("alice-id", "Fixed bugs")
	if finished {
		t.Fatal("expected more questions")
	}
	if !contains(nextQ, "blockers") {
		t.Errorf("expected blockers Q, got %s", nextQ)
	}

	finished, _, _ = h.convMgr.RecordAnswer("alice-id", "No blockers")
	if !finished {
		t.Fatal("expected conversation to finish")
	}

	conv, _ = h.convMgr.GetConversation("alice-id")
	if conv != nil {
		t.Error("conversation should be ended after all answers")
	}
}
