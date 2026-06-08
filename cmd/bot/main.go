package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/config"
	"github.com/sunba91-su/Roket.Chat-GeekBot/internal/rocket"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := rocket.New(cfg.ServerURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	client.OnMessage(func(msg rocket.IncomingMessage) {
		log.Printf("[%s] %s: %s", msg.RoomID, msg.Username, msg.Text)
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
