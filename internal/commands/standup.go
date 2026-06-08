package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/store"
)

func RegisterStandupCommands(r *Registry) {
	r.Register("submit", handleSubmit, PermissionMember)
	r.Register("status", handleStatus, PermissionMember)
}

func handleSubmit(ctx *Context) error {
	teams, err := ctx.Store.GetTeamsForUser(ctx.UserID)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			"Error looking up your teams.")
	}
	if len(teams) == 0 {
		return send(ctx.Messenger, ctx.RoomID,
			"You are not a member of any team.")
	}

	dmRoomID, err := ensureDMRoom(ctx)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Could not start DM: %v", err))
	}

	team := teams[0]
	date := time.Now().Format("2006-01-02")

	sessionID := fmt.Sprintf("sess-%d-%s", time.Now().UnixMilli(), team.ID)

	session := &store.StandupSession{
		ID:     sessionID,
		TeamID: team.ID,
		Date:   date,
		Status: "open",
	}
	if err := ctx.Store.CreateSession(session); err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to create session: %v", err))
	}

	questions := strings.Split(team.Questions, "|")

	ctx.ConvState.StartConversation(
		ctx.UserID, ctx.Username, team.ID,
		dmRoomID, questions, sessionID,
	)

	firstQ := questions[0]
	_ = send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("Starting your standup for **%s**. Check your DMs!", team.Name))

	return send(ctx.Messenger, dmRoomID,
		fmt.Sprintf("📋 *Daily Standup — %s*\n\n**Q1:** %s\n\nReply with your answer.", team.Name, firstQ))
}

func handleStatus(ctx *Context) error {
	teams, err := ctx.Store.GetTeamsForUser(ctx.UserID)
	if err != nil || len(teams) == 0 {
		return send(ctx.Messenger, ctx.RoomID, "You are not a member of any team.")
	}

	date := time.Now().Format("2006-01-02")
	team := teams[0]

	submitted, total, err := ctx.Store.GetSessionStatus(team.ID, date)
	if err != nil {
		session, err2 := ctx.Store.GetActiveSession(team.ID, date)
		if err2 != nil {
			return send(ctx.Messenger, ctx.RoomID,
				"No active standup for today.")
		}
		submitted, total, _ = ctx.Store.GetSessionStatus(team.ID, session.Date)
	}

	hasSubmitted, _ := ctx.Store.HasSubmitted(sessionID(ctx, team.ID), ctx.UserID)
	statusLine := ""
	if hasSubmitted {
		statusLine = "✅ You have submitted."
	} else {
		statusLine = "⏳ You have not submitted yet."
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("*Standup Status — %s*\n%s\nTeam progress: %d/%d submitted.",
			team.Name, statusLine, submitted, total))
}

func sessionID(ctx *Context, teamID string) string {
	date := time.Now().Format("2006-01-02")
	session, err := ctx.Store.GetActiveSession(teamID, date)
	if err != nil {
		return ""
	}
	return session.ID
}

func ensureDMRoom(ctx *Context) (string, error) {
	if ctx.IsDM && ctx.RoomID != "" {
		return ctx.RoomID, nil
	}

	dmProvider, ok := ctx.Messenger.(interface{ CreateDM(string) (string, error) })
	if !ok {
		return "", fmt.Errorf("DM creation not available")
	}
	return dmProvider.CreateDM(ctx.Username)
}
