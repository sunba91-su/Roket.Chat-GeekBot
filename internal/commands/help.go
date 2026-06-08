package commands

import "fmt"

func handleHelp(ctx *Context) error {
	base := "Available commands:\n"

	cmds := ""
	if ctx.Username == ctx.Config.MainAdmin {
		cmds += "  /standup team create <name> --channel #ch --schedule <cron>\n"
		cmds += "  /standup team delete <name>\n"
		cmds += "  /standup team set-lead <team> @user\n"
	}

	cmds += "  /standup team add @user\n"
	cmds += "  /standup team remove @user\n"
	cmds += "  /standup team members\n"
	cmds += "  /standup team set schedule <cron>\n"
	cmds += "  /standup team set channel #ch\n"
	cmds += "  /standup team set questions <q1|q2|q3>\n"
	cmds += "  /standup team set timezone <tz>\n"
	cmds += "  /standup submit — start daily standup\n"
	cmds += "  /standup status — check submission status\n"
	cmds += "  /standup cancel — cancel in-progress standup\n"
	cmds += "  /standup list — list your teams\n"
	cmds += "  /standup report — view latest report\n"
	cmds += "  /standup help — show this message"

	return send(ctx.Messenger, ctx.RoomID, base+cmds)
}

func sendHelpForCommand(ctx *Context, cmd string) error {
	help := map[string]string{
		"submit": "Start your daily standup. The bot will DM you with questions.",
		"status": "Check if you've submitted your standup today.",
		"cancel": "Cancel an in-progress standup submission.",
		"list":   "List teams you belong to.",
		"report": "Post the latest standup report to the team channel.",
	}

	msg, ok := help[cmd]
	if !ok {
		msg = fmt.Sprintf("No help available for `/standup %s`.", cmd)
	}
	return send(ctx.Messenger, ctx.RoomID, msg)
}
