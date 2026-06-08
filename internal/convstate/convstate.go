package convstate

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type State int

const (
	Idle State = iota
	Answering
)

type Conversation struct {
	UserID        string
	Username      string
	TeamID        string
	RoomID        string
	State         State
	Questions     []string
	CurrentQ      int
	Answers       []string
	SessionID     string
	StartedAt     time.Time
}

type Manager struct {
	mu   sync.RWMutex
	conv map[string]*Conversation
}

func NewManager() *Manager {
	return &Manager{conv: make(map[string]*Conversation)}
}

func (m *Manager) StartConversation(userID, username, teamID, roomID string, questions []string, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.conv[userID] = &Conversation{
		UserID:    userID,
		Username:  username,
		TeamID:    teamID,
		RoomID:    roomID,
		State:     Answering,
		Questions: questions,
		CurrentQ:  0,
		Answers:   make([]string, len(questions)),
		SessionID: sessionID,
		StartedAt: time.Now(),
	}
}

func (m *Manager) GetConversation(userID string) (*Conversation, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.conv[userID]
	if !ok || c.State == Idle {
		return nil, false
	}
	return c, true
}

func (m *Manager) RecordAnswer(userID, answer string) (finished bool, question string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.conv[userID]
	if !ok || c.State == Idle {
		return false, "", fmt.Errorf("no active conversation")
	}

	c.Answers[c.CurrentQ] = answer
	c.CurrentQ++

	if c.CurrentQ >= len(c.Questions) {
		c.State = Idle
		return true, "", nil
	}

	return false, c.Questions[c.CurrentQ], nil
}

func (m *Manager) GetProgress(userID string) (current int, total int, answers []string, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok2 := m.conv[userID]
	if !ok2 || c.State == Idle {
		return 0, 0, nil, false
	}

	answersCopy := make([]string, len(c.Answers))
	copy(answersCopy, c.Answers)
	return c.CurrentQ, len(c.Questions), answersCopy, true
}

func (m *Manager) EndConversation(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.conv[userID]; ok {
		c.State = Idle
	}
}

func (m *Manager) Cancel(userID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.conv[userID]
	if !ok || c.State == Idle {
		return "", fmt.Errorf("no active conversation")
	}

	answers := strings.Join(c.Answers[:c.CurrentQ], "\n")
	c.State = Idle
	return answers, nil
}
