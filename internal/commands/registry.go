package commands

import (
	"fmt"
	"strings"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/config"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/convstate"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/store"
)

type Messenger interface {
	SendMessage(roomID, text string) error
}

type UserInfo struct {
	ID       string
	Username string
	Name     string
}

type UserProvider interface {
	UserInfo(username string) (*UserInfo, error)
}

type Permission int

const (
	PermissionAny Permission = iota
	PermissionMember
	PermissionTeamLead
	PermissionAdmin
)

type Context struct {
	UserID   string
	Username string
	RoomID   string
	RawText  string
	Args     []string
	Store    *store.Store
	Messenger
	UserProvider
	Config    *config.Config
	ConvState *convstate.Manager
	CmdName   string
	IsDM      bool
}

type Handler func(ctx *Context) error

type entry struct {
	handler    Handler
	permission Permission
}

type Registry struct {
	entries map[string]entry
}

func New() *Registry {
	r := &Registry{entries: make(map[string]entry)}
	r.Register("help", handleHelp, PermissionAny)
	return r
}

func (r *Registry) Register(name string, handler Handler, perm Permission) {
	r.entries[name] = entry{handler: handler, permission: perm}
}

func (r *Registry) Dispatch(ctx *Context) (bool, error) {
	if !strings.HasPrefix(ctx.RawText, "/standup") {
		return false, nil
	}

	parts := strings.Fields(ctx.RawText)
	if len(parts) < 2 {
		ctx.CmdName = "help"
		ctx.Args = nil
		return true, r.execute(ctx, "help")
	}

	ctx.CmdName = strings.ToLower(parts[1])
	ctx.Args = parts[2:]
	return true, r.execute(ctx, ctx.CmdName)
}

func (r *Registry) execute(ctx *Context, name string) error {
	entry, ok := r.entries[name]
	if !ok {
		return send(ctx.Messenger, ctx.RoomID,
			fmt.Sprintf("Unknown command `/standup %s`. Try `/standup help`.", name))
	}

	if err := r.checkPermission(ctx, entry.permission); err != nil {
		return send(ctx.Messenger, ctx.RoomID, err.Error())
	}

	return entry.handler(ctx)
}

func (r *Registry) checkPermission(ctx *Context, required Permission) error {
	switch required {
	case PermissionAny:
		return nil
	case PermissionAdmin:
		if ctx.Username != ctx.Config.MainAdmin {
			return fmt.Errorf("Only the main admin can use this command.")
		}
		return nil
	case PermissionTeamLead:
		if ctx.Username == ctx.Config.MainAdmin {
			return nil
		}
		for _, teamID := range ctx.Args {
			isLead, err := ctx.Store.IsTeamLead(teamID, ctx.UserID)
			if err != nil {
				return fmt.Errorf("Error checking permissions.")
			}
			if !isLead {
				return fmt.Errorf("You are not a team lead for this team.")
			}
		}
		return nil
	case PermissionMember:
		if ctx.Username == ctx.Config.MainAdmin {
			return nil
		}
		teams, err := ctx.Store.GetTeamsForUser(ctx.UserID)
		if err != nil {
			return fmt.Errorf("Error checking membership.")
		}
		if len(teams) == 0 {
			return fmt.Errorf("You are not a member of any team.")
		}
		return nil
	default:
		return nil
	}
}

func send(m Messenger, roomID, text string) error {
	if text == "" {
		return nil
	}
	return m.SendMessage(roomID, text)
}
