package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/commands"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/config"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/convstate"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/rocket"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/store"
)

type userProviderAdapter struct {
	client *rocket.Client
}

func (a *userProviderAdapter) UserInfo(username string) (*commands.UserInfo, error) {
	user, err := a.client.UserInfo(username)
	if err != nil {
		return nil, err
	}
	return &commands.UserInfo{
		ID:       user.ID,
		Username: user.UserName,
		Name:     user.Name,
	}, nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-health" {
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbPath := os.Getenv("STANDUP_DB_PATH")
	if dbPath == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Cannot determine home dir: %v", err)
		}
		dbPath = filepath.Join(dir, "standup-bot.db")
	}

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer st.Close()
	log.Printf("Database initialized at %s", dbPath)

	client, err := rocket.New(cfg.ServerURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	cmdReg := commands.New()
	commands.RegisterTeamCommands(cmdReg)
	commands.RegisterStandupCommands(cmdReg)

	convMgr := convstate.NewManager()

	client.OnMessage(func(msg rocket.IncomingMessage) {
		isDM, _ := client.IsDMRoom(msg.RoomID)

		ctx := &commands.Context{
			UserID:       msg.UserID,
			Username:     msg.Username,
			RoomID:       msg.RoomID,
			RawText:      msg.Text,
			Store:        st,
			Messenger:    client,
			UserProvider: &userProviderAdapter{client: client},
			Config:       cfg,
			ConvState:    convMgr,
			IsDM:         isDM,
		}

		handled, err := cmdReg.Dispatch(ctx)
		if err != nil {
			log.Printf("Command error: %v", err)
		}
		if handled {
			log.Printf("[%s] %s: %s", msg.RoomID, msg.Username, msg.Text)
			return
		}

		if !isDM {
			return
		}

		if err := handleStandupReply(ctx); err != nil {
			log.Printf("Standup reply error: %v", err)
		}
	})

	if err := client.Connect(cfg.BotUser, cfg.BotPass); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	if err := client.SubscribeToMyMessages(); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Println("Bot is running. Press Ctrl+C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	fmt.Println()
	log.Println("Shutting down...")
}

func handleStandupReply(ctx *commands.Context) error {
	conv, ok := ctx.ConvState.GetConversation(ctx.UserID)
	if !ok {
		return nil
	}
	if conv.RoomID != ctx.RoomID {
		return nil
	}

	finished, nextQ, err := ctx.ConvState.RecordAnswer(ctx.UserID, ctx.RawText)
	if err != nil {
		return ctx.SendMessage(ctx.RoomID, "Error recording your answer.")
	}

	if finished {
		conv, ok := ctx.ConvState.GetConversation(ctx.UserID)
		if !ok {
			return ctx.SendMessage(ctx.RoomID, "Error: conversation lost.")
		}
		answersJoined := strings.Join(conv.Answers, "|")
		resp := &store.StandupResponse{
			ID:          fmt.Sprintf("resp-%d", time.Now().UnixMilli()),
			SessionID:   conv.SessionID,
			UserID:      conv.UserID,
			Username:    conv.Username,
			Answers:     answersJoined,
			SubmittedAt: time.Now(),
		}
		if err := ctx.Store.SubmitResponse(resp); err != nil {
			return ctx.SendMessage(ctx.RoomID,
				fmt.Sprintf("Error saving standup: %v", err))
		}

		ctx.ConvState.EndConversation(ctx.UserID)
		return ctx.SendMessage(ctx.RoomID,
			"✅ *Standup submitted!* Thank you. Have a great day!")
	}

	return ctx.SendMessage(ctx.RoomID,
		fmt.Sprintf("**Q%d:** %s\n\nReply with your answer.", conv.CurrentQ+1, nextQ))
}
