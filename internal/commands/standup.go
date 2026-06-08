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
	r.Register("report", handleReport, PermissionAny)
	r.Register("cancel", handleCancel, PermissionMember)
	r.Register("list", handleList, PermissionAny)
}

func resolveSubmitTeam(ctx *Context, teams []*store.Team) (*store.Team, string, error) {
	if len(teams) == 1 {
		return teams[0], "", nil
	}

	for i, a := range ctx.Args {
		if (a == "--team" || a == "-t") && i+1 < len(ctx.Args) {
			name := ctx.Args[i+1]
			for _, t := range teams {
				if strings.EqualFold(t.Name, name) {
					return t, "", nil
				}
			}
			return nil, "", fmt.Errorf("You are not a member of team %q.", name)
		}
	}

	var names []string
	for _, t := range teams {
		names = append(names, t.Name)
	}
	return nil, "", fmt.Errorf(
		"You are in multiple teams. Use `--team \"<name>\"`: %s",
		strings.Join(names, ", "),
	)
}

func activeOrNewSession(ctx *Context, teamID, teamName string) (string, error) {
	date := time.Now().Format("2006-01-02")

	session, err := ctx.Store.GetActiveSession(teamID, date)
	if err == nil {
		submitted, _ := ctx.Store.HasSubmitted(session.ID, ctx.UserID)
		if submitted {
			return "", fmt.Errorf("You have already submitted your standup for %s today.", teamName)
		}
		return session.ID, nil
	}

	sessionID := fmt.Sprintf("sess-%d-%s", time.Now().UnixMilli(), teamID)
	session = &store.StandupSession{
		ID:     sessionID,
		TeamID: teamID,
		Date:   date,
		Status: "open",
	}
	if err := ctx.Store.CreateSession(session); err != nil {
		return "", fmt.Errorf("Failed to create session: %v", err)
	}
	return sessionID, nil
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

	team, _, err := resolveSubmitTeam(ctx, teams)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
	}

	dmRoomID, err := ensureDMRoom(ctx)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Could not start DM: %v", err))
	}

	sessionID, err := activeOrNewSession(ctx, team.ID, team.Name)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
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

	team, _, err := resolveSubmitTeam(ctx, teams)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
	}

	date := time.Now().Format("2006-01-02")

	sessID := findSessionID(ctx, team.ID, date)
	submitted, total, err := ctx.Store.GetSessionStatus(team.ID, date)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, "No active standup for today.")
	}

	hasSubmitted, _ := ctx.Store.HasSubmitted(sessID, ctx.UserID)
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

func handleCancel(ctx *Context) error {
	conv, ok := ctx.ConvState.GetConversation(ctx.UserID)
	if !ok {
		return send(ctx.Messenger, ctx.RoomID,
			"You don't have an active standup submission.")
	}

	cancelTo := conv.RoomID
	if !ctx.IsDM {
		cancelTo = ctx.RoomID
	}

	partial, err := ctx.ConvState.Cancel(ctx.UserID)
	if err != nil {
		return send(ctx.Messenger, cancelTo,
			"You don't have an active standup submission.")
	}

	if partial == "" {
		partial = "No answers recorded."
	}

	return send(ctx.Messenger, cancelTo,
		fmt.Sprintf("❌ *Standup cancelled.* Your partial answers:\n%s", partial))
}

func handleList(ctx *Context) error {
	var teams []*store.Team
	var err error

	if ctx.Username == ctx.Config.MainAdmin {
		teams, err = ctx.Store.ListTeams()
	} else {
		teams, err = ctx.Store.GetTeamsForUser(ctx.UserID)
	}
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			"Error looking up teams.")
	}
	if len(teams) == 0 {
		return send(ctx.Messenger, ctx.RoomID, "No teams found.")
	}

	var lines []string
	for _, t := range teams {
		isLead, _ := ctx.Store.IsTeamLead(t.ID, ctx.UserID)
		role := ""
		if isLead {
			role = " (lead)"
		}
		lines = append(lines, fmt.Sprintf("  • **%s**%s", t.Name, role))
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("Teams:\n%s", strings.Join(lines, "\n")))
}

func handleReport(ctx *Context) error {
	team, err := resolveTeam(ctx)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
	}

	if ctx.Username != ctx.Config.MainAdmin {
		isLead, _ := ctx.Store.IsTeamLead(team.ID, ctx.UserID)
		if !isLead {
			return send(ctx.Messenger, ctx.RoomID,
				"Only team leads and the main admin can post reports.")
		}
	}

	date := time.Now().Format("2006-01-02")
	sessID := findSessionID(ctx, team.ID, date)
	if sessID == "" {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("No standup session found for %s today.", team.Name))
	}

	responses, err := ctx.Store.GetResponses(sessID)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to get responses: %v", err))
	}

	questions := strings.Split(team.Questions, "|")

	report := buildReport(team.Name, date, questions, responses)

	submitted, total, _ := ctx.Store.GetSessionStatus(team.ID, date)

	sendTo := team.ChannelID
	if len(ctx.Args) > 0 && ctx.Args[0] == "--here" {
		sendTo = ctx.RoomID
	}

	if err := send(ctx.Messenger, sendTo, report); err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to post report: %v", err))
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("✅ Report posted to #%s (%d/%d submitted).",
			sendTo, submitted, total))
}

func buildReport(teamName, date string, questions []string, responses []*store.StandupResponse) string {
	emojis := []string{"✅", "💻", "⚠️", "📌", "🔧", "🎯"}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 *Daily Standup — %s* (%s)\n\n", teamName, date))

	for _, r := range responses {
		sb.WriteString(fmt.Sprintf("👤 @%s\n", r.Username))
		answers := strings.Split(r.Answers, "|")
		for i, a := range answers {
			q := ""
			if i < len(questions) {
				q = questions[i]
			}
			emoji := emojis[i%len(emojis)]
			if q != "" {
				sb.WriteString(fmt.Sprintf("%s *%s* %s\n", emoji, q, a))
			} else {
				sb.WriteString(fmt.Sprintf("%s %s\n", emoji, a))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func findSessionID(ctx *Context, teamID, date string) string {
	session, err := ctx.Store.GetActiveSession(teamID, date)
	if err == nil {
		return session.ID
	}
	sessions, err := ctx.Store.ListSessions(teamID, 1)
	if err != nil || len(sessions) == 0 {
		return ""
	}
	return sessions[0].ID
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
