// Rich card example for Feishu/Lark.
// This bot demonstrates how to send rich interactive cards.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	sdk "github.com/doarvid/go-channel"
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
		AppID:         appID,
		AppSecret:     appSecret,
		ReactionEmoji: "OnIt", // Show a reaction when processing
	})
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Set message handler
	bot.OnMessage(func(ctx context.Context, msg *sdk.Message) error {
		log.Printf("Received: %s", msg.Content)

		// Handle different commands
		switch msg.Content {
		case "/card":
			return sendSampleCard(msg)
		case "/buttons":
			return sendButtonCard(msg)
		case "/help":
			return sendHelpCard(msg)
		default:
			return sendWelcomeCard(msg)
		}
	})

	// Start the bot
	log.Println("Starting bot...")
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
	log.Println("Bot started! Try sending: /card, /buttons, or /help")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	bot.Stop()
}

func sendWelcomeCard(msg *sdk.Message) error {
	card := sdk.NewCard().
		Title("👋 Welcome!", "blue").
		Markdown("Hello! I'm a demo bot showcasing rich cards.\n\nTry these commands:").
		Divider().
		Buttons(
			sdk.PrimaryBtn("Show Sample Card", "cmd:/card"),
			sdk.DefaultBtn("Show Buttons", "cmd:/buttons"),
		)

	return msg.SendCard(card.Build())
}

func sendSampleCard(msg *sdk.Message) error {
	card := sdk.NewCard().
		Title("📋 Sample Card", "green").
		Markdown("This is a **rich card** with markdown support!\n\nFeatures:").
		Divider().
		Markdown("- Bold text: **bold**\n- Italic: *italic*\n- Code: `inline code`").
		Divider().
		Markdown("```go\n// Code blocks work too!\nfunc hello() {\n    fmt.Println(\"Hello!\")\n}\n```").
		Note("This is a note at the bottom")

	return msg.SendCard(card.Build())
}

func sendButtonCard(msg *sdk.Message) error {
	card := sdk.NewCard().
		Title("🔘 Button Demo", "orange").
		Markdown("Click any button below:").
		Divider().
		Buttons(
			sdk.PrimaryBtn("Primary", "act:primary"),
			sdk.DefaultBtn("Default", "act:default"),
			sdk.DangerBtn("Danger", "act:danger"),
		).
		Divider().
		Markdown("Equal width buttons:").
		ButtonsEqual(
			sdk.DefaultBtn("Option 1", "act:opt1"),
			sdk.DefaultBtn("Option 2", "act:opt2"),
			sdk.DefaultBtn("Option 3", "act:opt3"),
		)

	return msg.SendCard(card.Build())
}

func sendHelpCard(msg *sdk.Message) error {
	card := sdk.NewCard().
		Title("❓ Help", "purple").
		Markdown("Available commands:\n\n").
		ListItem("Show a sample card with markdown", "/card", "cmd:/card").
		ListItem("Show button examples", "/buttons", "cmd:/buttons").
		ListItem("Show this help", "/help", "cmd:/help")

	return msg.SendCard(card.Build())
}
