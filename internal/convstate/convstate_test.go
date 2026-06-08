package convstate_test

import (
	"testing"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/convstate"
)

func TestStartAndGetConversation(t *testing.T) {
	m := convstate.NewManager()
	m.StartConversation("u1", "alice", "team1", "room1",
		[]string{"Q1?", "Q2?", "Q3?"}, "sess1")

	conv, ok := m.GetConversation("u1")
	if !ok {
		t.Fatal("expected conversation to exist")
	}
	if conv.Username != "alice" {
		t.Errorf("expected alice, got %s", conv.Username)
	}
	if conv.CurrentQ != 0 {
		t.Errorf("expected currentQ 0, got %d", conv.CurrentQ)
	}
}

func TestGetConversationIdle(t *testing.T) {
	m := convstate.NewManager()

	_, ok := m.GetConversation("nonexistent")
	if ok {
		t.Error("expected no conversation for unknown user")
	}
}

func TestRecordAnswer(t *testing.T) {
	m := convstate.NewManager()
	m.StartConversation("u1", "alice", "team1", "room1",
		[]string{"Q1?", "Q2?", "Q3?"}, "sess1")

	finished, nextQ, err := m.RecordAnswer("u1", "Answer 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if finished {
		t.Error("expected not finished after first answer")
	}
	if nextQ != "Q2?" {
		t.Errorf("expected Q2?, got %s", nextQ)
	}

	finished, nextQ, _ = m.RecordAnswer("u1", "Answer 2")
	if finished {
		t.Error("expected not finished after second answer")
	}
	if nextQ != "Q3?" {
		t.Errorf("expected Q3?, got %s", nextQ)
	}

	finished, _, _ = m.RecordAnswer("u1", "Answer 3")
	if !finished {
		t.Error("expected finished after third answer")
	}

	_, _, err = m.RecordAnswer("u1", "extra")
	if err == nil {
		t.Error("expected error after conversation ended")
	}
}

func TestRecordAnswerNoConversation(t *testing.T) {
	m := convstate.NewManager()

	_, _, err := m.RecordAnswer("u1", "answer")
	if err == nil {
		t.Error("expected error for no conversation")
	}
}

func TestGetProgress(t *testing.T) {
	m := convstate.NewManager()
	m.StartConversation("u1", "alice", "team1", "room1",
		[]string{"Q1?", "Q2?"}, "sess1")

	current, total, answers, ok := m.GetProgress("u1")
	if !ok {
		t.Fatal("expected progress")
	}
	if current != 0 {
		t.Errorf("expected 0, got %d", current)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(answers) != 2 {
		t.Errorf("expected 2 answers, got %d", len(answers))
	}

	m.RecordAnswer("u1", "A1")

	current, _, _, _ = m.GetProgress("u1")
	if current != 1 {
		t.Errorf("expected current 1, got %d", current)
	}
}

func TestCancel(t *testing.T) {
	m := convstate.NewManager()
	m.StartConversation("u1", "alice", "team1", "room1",
		[]string{"Q1?", "Q2?"}, "sess1")

	m.RecordAnswer("u1", "Partial answer")

	answers, err := m.Cancel("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if answers == "" {
		t.Error("expected non-empty answers")
	}

	_, ok := m.GetConversation("u1")
	if ok {
		t.Error("expected conversation to be ended")
	}
}

func TestCancelNoConversation(t *testing.T) {
	m := convstate.NewManager()

	_, err := m.Cancel("u1")
	if err == nil {
		t.Error("expected error for no conversation")
	}
}

func TestEndConversation(t *testing.T) {
	m := convstate.NewManager()
	m.StartConversation("u1", "alice", "team1", "room1",
		[]string{"Q1?"}, "sess1")

	m.EndConversation("u1")

	_, ok := m.GetConversation("u1")
	if ok {
		t.Error("expected conversation to be ended")
	}
}

func TestMultipleUsers(t *testing.T) {
	m := convstate.NewManager()
	m.StartConversation("u1", "alice", "team1", "r1",
		[]string{"Q1?", "Q2?"}, "s1")
	m.StartConversation("u2", "bob", "team1", "r2",
		[]string{"Q1?"}, "s1")

	finished, _, _ := m.RecordAnswer("u1", "A1")
	if finished {
		t.Error("u1 should not be finished")
	}

	finished, _, _ = m.RecordAnswer("u2", "A1")
	if !finished {
		t.Error("u2 should be finished after 1 question")
	}

	conv, _ := m.GetConversation("u1")
	if conv.CurrentQ != 1 {
		t.Errorf("u1 should be on Q2, got Q%d", conv.CurrentQ)
	}
}
