package commands_test

import (
	"testing"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/commands"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/config"
)

type mockMessenger struct {
	sent []string
}

func (m *mockMessenger) SendMessage(roomID, text string) error {
	m.sent = append(m.sent, text)
	return nil
}

func newTestContext(text string) (*commands.Context, *mockMessenger) {
	m := &mockMessenger{}
	return &commands.Context{
		UserID:    "user1",
		Username:  "alice",
		RoomID:    "GENERAL",
		RawText:   text,
		Messenger: m,
		Config:    &config.Config{MainAdmin: "admin"},
	}, m
}

func TestDispatchIgnoresNonCommand(t *testing.T) {
	ctx, _ := newTestContext("hello world")
	reg := commands.New()

	handled, err := reg.Dispatch(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handled {
		t.Error("expected non-command to not be handled")
	}
}

func TestDispatchHelp(t *testing.T) {
	ctx, msgr := newTestContext("/standup help")
	reg := commands.New()

	handled, err := reg.Dispatch(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Error("expected /standup help to be handled")
	}
	if len(msgr.sent) == 0 {
		t.Error("expected help text to be sent")
	}
}

func TestDispatchBareStandupShowsHelp(t *testing.T) {
	ctx, msgr := newTestContext("/standup")
	reg := commands.New()

	handled, err := reg.Dispatch(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Error("expected bare /standup to be handled (help)")
	}
	if len(msgr.sent) == 0 {
		t.Error("expected help text to be sent")
	}
}

func TestDispatchUnknownCommand(t *testing.T) {
	ctx, msgr := newTestContext("/standup nonexistent")
	reg := commands.New()

	handled, err := reg.Dispatch(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Error("expected unknown command to still be handled")
	}
	if len(msgr.sent) == 0 {
		t.Error("expected error message to be sent")
	}
}

func TestDispatchAdminCommandDenied(t *testing.T) {
	ctx, msgr := newTestContext("/standup nonexistent-admin-cmd")

	var called bool
	reg := commands.New()
	reg.Register("nonexistent-admin-cmd", func(ctx *commands.Context) error {
		called = true
		return nil
	}, commands.PermissionAdmin)

	_, _ = reg.Dispatch(ctx)
	if called {
		t.Error("admin command should not have been called for non-admin")
	}
	if len(msgr.sent) == 0 {
		t.Error("expected permission denied message")
	}
}

func TestRegisterAndDispatch(t *testing.T) {
	ctx, _ := newTestContext("/standup ping")

	var called bool
	reg := commands.New()
	reg.Register("ping", func(ctx *commands.Context) error {
		called = true
		return nil
	}, commands.PermissionAny)

	handled, err := reg.Dispatch(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Error("expected ping to be handled")
	}
	if !called {
		t.Error("expected handler to be called")
	}
}
