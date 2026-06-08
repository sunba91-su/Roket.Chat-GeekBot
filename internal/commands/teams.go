package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/store"
)

func RegisterTeamCommands(r *Registry) {
	r.Register("team", handleTeam, PermissionAny)
}

func handleTeam(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return send(ctx.Messenger, ctx.RoomID,
			"Usage: /standup team <create|delete|set-lead|add|remove|members|set> ...")
	}

	sub := strings.ToLower(ctx.Args[0])
	args := ctx.Args[1:]

	switch sub {
	case "create":
		return handleTeamCreate(ctx, args)
	case "delete":
		return handleTeamDelete(ctx, args)
	case "set-lead":
		return handleTeamSetLead(ctx, args)
	case "add":
		return handleTeamAdd(ctx, args)
	case "remove":
		return handleTeamRemove(ctx, args)
	case "members":
		return handleTeamMembers(ctx, args)
	case "set":
		return handleTeamSet(ctx, args)
	default:
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Unknown subcommand: %s. Try /standup help.", sub))
	}
}

func handleTeamCreate(ctx *Context, args []string) error {
	if ctx.Username != ctx.Config.MainAdmin {
		return send(ctx.Messenger, ctx.RoomID,
			"Only the main admin can create teams.")
	}

	if len(args) < 1 {
		return send(ctx.Messenger, ctx.RoomID,
			"Usage: /standup team create <name> [--channel #ch]")
	}

	name := args[0]
	channelID := ""
	for i, a := range args {
		if a == "--channel" && i+1 < len(args) {
			ch := args[i+1]
			channelID = strings.TrimPrefix(ch, "#")
		}
	}

	if channelID == "" {
		channelID = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	team := &store.Team{
		ID:        fmt.Sprintf("team-%d", time.Now().UnixMilli()),
		Name:      name,
		ChannelID: channelID,
		Questions: "How are you feeling?|What did you do yesterday?|What are you doing today?|Any blockers?",
		Timezone:  "UTC",
	}

	if err := ctx.Store.CreateTeam(team); err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to create team: %v", err))
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("Team %q created (channel: #%s). Set a lead with `/standup team set-lead %s @user`.",
			name, channelID, name))
}

func handleTeamDelete(ctx *Context, args []string) error {
	if ctx.Username != ctx.Config.MainAdmin {
		return send(ctx.Messenger, ctx.RoomID,
			"Only the main admin can delete teams.")
	}
	if len(args) < 1 {
		return send(ctx.Messenger, ctx.RoomID,
			"Usage: /standup team delete <name>")
	}

	team, err := ctx.Store.GetTeamByName(args[0])
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Team %q not found.", args[0]))
	}

	if err := ctx.Store.DeleteTeam(team.ID); err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to delete team: %v", err))
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("Team %q deleted.", args[0]))
}

func handleTeamSetLead(ctx *Context, args []string) error {
	if ctx.Username != ctx.Config.MainAdmin {
		return send(ctx.Messenger, ctx.RoomID,
			"Only the main admin can set team leads.")
	}
	if len(args) < 2 {
		return send(ctx.Messenger, ctx.RoomID,
			"Usage: /standup team set-lead <team> @user")
	}

	teamName := args[0]
	username := strings.TrimPrefix(args[1], "@")

	team, err := ctx.Store.GetTeamByName(teamName)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Team %q not found.", teamName))
	}

	user, err := ctx.UserProvider.UserInfo(username)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("User %q not found on the server.", username))
	}

	isMember, _ := ctx.Store.IsMember(team.ID, user.ID)
	if !isMember {
		if err := ctx.Store.AddMember(&store.Member{
			ID:       fmt.Sprintf("mem-%d", time.Now().UnixMilli()),
			TeamID:   team.ID,
			UserID:   user.ID,
			Username: user.Username,
			Role:     "lead",
		}); err != nil {
			return send(ctx.Messenger, ctx.RoomID,
				fmt.Sprintf("Failed to add lead: %v", err))
		}
	} else {
		if err := ctx.Store.SetRole(team.ID, user.ID, "lead"); err != nil {
			return send(ctx.Messenger, ctx.RoomID,
				fmt.Sprintf("Failed to set role: %v", err))
		}
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("@%s is now the lead of %s.", user.Username, teamName))
}

func handleTeamAdd(ctx *Context, args []string) error {
	if len(args) < 1 {
		return send(ctx.Messenger, ctx.RoomID,
			"Usage: /standup team add @user")
	}

	username := strings.TrimPrefix(args[0], "@")

	team, err := resolveTeam(ctx)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
	}

	user, err := ctx.UserProvider.UserInfo(username)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("User %q not found on the server.", username))
	}

	if err := ctx.Store.AddMember(&store.Member{
		ID:       fmt.Sprintf("mem-%d", time.Now().UnixMilli()),
		TeamID:   team.ID,
		UserID:   user.ID,
		Username: user.Username,
		Role:     "member",
	}); err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to add member: %v", err))
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("@%s added to %s.", user.Username, team.Name))
}

func handleTeamRemove(ctx *Context, args []string) error {
	if len(args) < 1 {
		return send(ctx.Messenger, ctx.RoomID,
			"Usage: /standup team remove @user")
	}

	username := strings.TrimPrefix(args[0], "@")

	team, err := resolveTeam(ctx)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
	}

	user, err := ctx.UserProvider.UserInfo(username)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("User %q not found on the server.", username))
	}

	if err := ctx.Store.RemoveMember(team.ID, user.ID); err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to remove member: %v", err))
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("@%s removed from %s.", user.Username, team.Name))
}

func handleTeamMembers(ctx *Context, args []string) error {
	team, err := resolveTeam(ctx)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
	}

	members, err := ctx.Store.GetMembers(team.ID)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to get members: %v", err))
	}

	if len(members) == 0 {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("No members in %s.", team.Name))
	}

	var lines []string
	for _, m := range members {
		if m.Role == "lead" {
			lines = append(lines, fmt.Sprintf("  @%s (lead)", m.Username))
		} else {
			lines = append(lines, fmt.Sprintf("  @%s", m.Username))
		}
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("Members of %s:\n%s", team.Name, strings.Join(lines, "\n")))
}

func handleTeamSet(ctx *Context, args []string) error {
	if len(args) < 2 {
		return send(ctx.Messenger, ctx.RoomID,
			"Usage: /standup team set <schedule|channel|questions|timezone> <value>")
	}

	field := strings.ToLower(args[0])
	value := strings.Join(args[1:], " ")

	team, err := resolveTeam(ctx)
	if err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
	}

	switch field {
	case "schedule":
		team.ScheduleCron = value
	case "channel":
		team.ChannelID = strings.TrimPrefix(value, "#")
	case "questions":
		if !strings.Contains(value, "|") {
			return send(ctx.Messenger, ctx.RoomID,
				"Questions must be pipe-separated, e.g. \"Q1|Q2|Q3\"")
		}
		team.Questions = value
	case "timezone":
		team.Timezone = value
	default:
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Unknown field: %s. Use schedule, channel, questions, or timezone.", field))
	}

	if err := ctx.Store.UpdateTeam(team); err != nil {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Failed to update: %v", err))
	}

	return send(ctx.Messenger, ctx.RoomID,
		fmt.Sprintf("%s team %s updated to %q.", team.Name, field, value))
}

func resolveTeam(ctx *Context) (*store.Team, error) {
	if ctx.Username == ctx.Config.MainAdmin {
		if len(ctx.Args) < 2 {
			return nil, fmt.Errorf("As admin, specify the team name: /standup team <cmd> <team> ...")
		}
		return ctx.Store.GetTeamByName(ctx.Args[1])
	}

	teams, err := ctx.Store.GetTeamsForUser(ctx.UserID)
	if err != nil {
		return nil, fmt.Errorf("Error looking up your teams.")
	}

	var leadTeams []*store.Team
	for _, t := range teams {
		isLead, _ := ctx.Store.IsTeamLead(t.ID, ctx.UserID)
		if isLead {
			leadTeams = append(leadTeams, t)
		}
	}

	if len(leadTeams) == 0 {
		return nil, fmt.Errorf("You are not a team lead for any team.")
	}
	if len(leadTeams) > 1 {
		return nil, fmt.Errorf("You lead multiple teams. As admin, specify the team name.")
	}

	return leadTeams[0], nil
}
