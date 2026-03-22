// Simple echo bot example for Feishu/Lark.
// This bot echoes back any text message it receives.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	sdk "github.com/chenhg5/cc-connect/platform-sdk"
)

func main() {
	// Get credentials from environment variables
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")

	if appID == "" || appSecret == "" {
		log.Fatal("Please set FEISHU_APP_ID and FEISHU_APP_SECRET environment variables")
	}

	// Create a Feishu bot
	bot, err := sdk.NewFeishuBot(sdk.FeishuConfig{
		AppID:     appID,
		AppSecret: appSecret,
		// IsLark: true, // Uncomment for Lark (international version)
	})
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Set message handler
	bot.OnMessage(func(ctx context.Context, msg *sdk.Message) error {
		log.Printf("Received message from %s (%s): %s", msg.UserName, msg.UserID, msg.Content)

		// Echo back the message
		reply := "You said: " + msg.Content
		if err := msg.Reply(reply); err != nil {
			log.Printf("Failed to reply: %v", err)
			return err
		}

		log.Printf("Replied: %s", reply)
		return nil
	})

	// Start the bot
	log.Println("Starting bot...")
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
	log.Println("Bot started successfully!")
	log.Printf("Bot name: %s", bot.Name())

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down bot...")
	if err := bot.Stop(); err != nil {
		log.Printf("Error stopping bot: %v", err)
	}
	log.Println("Bot stopped")
}
