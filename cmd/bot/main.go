package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/commands"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/config"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/rocket"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/store"
)

func main() {
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

	client.OnMessage(func(msg rocket.IncomingMessage) {
		ctx := &commands.Context{
			UserID:    msg.UserID,
			Username:  msg.Username,
			RoomID:    msg.RoomID,
			RawText:   msg.Text,
			Store:     st,
			Messenger: client,
			Config:    cfg,
		}

		handled, err := cmdReg.Dispatch(ctx)
		if err != nil {
			log.Printf("Command error: %v", err)
		}
		if handled {
			log.Printf("[%s] %s: %s", msg.RoomID, msg.Username, msg.Text)
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
