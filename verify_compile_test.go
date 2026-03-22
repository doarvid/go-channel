//go:build ignore
// +build ignore

package main

import (
	"log"

	sdk "github.com/chenhg5/cc-connect/platform-sdk"
)

func main() {
	log.Println("Testing SDK compilation...")

	log.Println("Testing Feishu bot...")
	_, err := sdk.NewFeishuBot(sdk.FeishuConfig{
		AppID:     "test_app_id",
		AppSecret: "test_app_secret",
	})
	if err != nil {
		log.Printf("NewFeishuBot (expected error for invalid credentials): %v", err)
	}

	log.Println("Testing Telegram bot...")
	_, err = sdk.NewTelegramBot(sdk.TelegramConfig{
		Token: "test_token",
	})
	if err != nil {
		log.Printf("NewTelegramBot (expected error for invalid credentials): %v", err)
	}

	log.Println("Testing Discord bot...")
	_, err = sdk.NewDiscordBot(sdk.DiscordConfig{
		Token: "test_token",
	})
	if err != nil {
		log.Printf("NewDiscordBot (expected error for invalid credentials): %v", err)
	}

	log.Println("Testing Slack bot...")
	_, err = sdk.NewSlackBot(sdk.SlackConfig{
		BotToken: "test_bot_token",
		AppToken: "test_app_token",
	})
	if err != nil {
		log.Printf("NewSlackBot (expected error for invalid credentials): %v", err)
	}

	log.Println("Testing DingTalk bot...")
	_, err = sdk.NewDingTalkBot(sdk.DingTalkConfig{
		AppKey:    "test_app_key",
		AppSecret: "test_app_secret",
	})
	if err != nil {
		log.Printf("NewDingTalkBot (expected error for invalid credentials): %v", err)
	}

	log.Println("Testing WeCom bot...")
	_, err = sdk.NewWeComBot(sdk.WeComConfig{
		CorpID:  "test_corp_id",
		AgentID: 1000001,
		Secret:  "test_secret",
	})
	if err != nil {
		log.Printf("NewWeComBot (expected error for invalid credentials): %v", err)
	}

	// Test card builder
	card := sdk.NewCard().
		Title("Test", "blue").
		Markdown("Hello **world**").
		Buttons(
			sdk.PrimaryBtn("OK", "act:ok"),
		).
		Build()

	log.Printf("Card created with %d elements", len(card.Elements))

	// Test button helpers
	_ = sdk.DefaultBtn("Cancel", "act:cancel")
	_ = sdk.DangerBtn("Delete", "act:delete")

	// Test listing platforms
	platforms := sdk.ListPlatforms()
	log.Printf("Available platforms: %v", platforms)

	log.Println("SDK compiled successfully! All platforms are supported!")
}
