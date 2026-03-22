// Package platformsdk provides a simple SDK for integrating messaging platforms
// (Feishu/Lark, etc.) into your Go application.
//
// Example usage:
//
//	package main
//
//	import (
//	    "context"
//	    "log"
//
//	    sdk "github.com/chenhg5/cc-connect/platform-sdk"
//	    "github.com/chenhg5/cc-connect/platform-sdk/core"
//	)
//
//	func main() {
//	    // Create a Feishu bot
//	    bot, err := sdk.NewFeishuBot(sdk.FeishuConfig{
//	        AppID:     "cli_xxx",
//	        AppSecret: "xxx",
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Set message handler
//	    bot.OnMessage(func(ctx context.Context, msg *sdk.Message) error {
//	        // Echo back the message
//	        return msg.Reply("You said: " + msg.Content)
//	    })
//
//	    // Start the bot
//	    if err := bot.Start(); err != nil {
//	        log.Fatal(err)
//	    }
//	    defer bot.Stop()
//
//	    // Wait for interrupt signal...
//	    select {}
//	}
package platformsdk

import (
	"context"
	"fmt"

	"github.com/chenhg5/cc-connect/platform-sdk/core"
	"github.com/chenhg5/cc-connect/platform-sdk/dingtalk"
	"github.com/chenhg5/cc-connect/platform-sdk/discord"
	"github.com/chenhg5/cc-connect/platform-sdk/feishu"
	"github.com/chenhg5/cc-connect/platform-sdk/slack"
	"github.com/chenhg5/cc-connect/platform-sdk/telegram"
	"github.com/chenhg5/cc-connect/platform-sdk/wecom"
)

// Message wraps a core.Message with convenient methods for replying.
type Message struct {
	*core.Message
	platform core.Platform
}

// Reply sends a reply message to the same chat as a reply to the original message.
func (m *Message) Reply(content string) error {
	return m.platform.Reply(context.Background(), m.ReplyCtx, content)
}

// Replyf sends a formatted reply message.
func (m *Message) Replyf(format string, args ...interface{}) error {
	return m.Reply(fmt.Sprintf(format, args...))
}

// Send sends a new message to the same chat (not a reply to original message).
func (m *Message) Send(content string) error {
	return m.platform.Send(context.Background(), m.ReplyCtx, content)
}

// Sendf sends a formatted new message.
func (m *Message) Sendf(format string, args ...interface{}) error {
	return m.Send(fmt.Sprintf(format, args...))
}

// SendCard sends a rich card message as a reply.
func (m *Message) SendCard(card *core.Card) error {
	if sender, ok := m.platform.(core.CardSender); ok {
		return sender.SendCard(context.Background(), m.ReplyCtx, card)
	}
	return m.Reply(card.RenderText())
}

// ReplyCard sends a rich card message as a reply.
func (m *Message) ReplyCard(card *core.Card) error {
	if sender, ok := m.platform.(core.CardSender); ok {
		return sender.ReplyCard(context.Background(), m.ReplyCtx, card)
	}
	return m.Reply(card.RenderText())
}

// MessageHandler is the type for message handling functions.
type MessageHandler func(ctx context.Context, msg *Message) error

// Bot represents a messaging bot.
type Bot struct {
	platform core.Platform
	handler  MessageHandler
}

// FeishuConfig contains configuration for a Feishu/Lark bot.
type FeishuConfig struct {
	AppID                 string // required: Feishu/Lark App ID
	AppSecret             string // required: Feishu/Lark App Secret
	IsLark                bool   // optional: set to true for Lark (international version)
	ReactionEmoji         string // optional: emoji for reaction indicator, default "OnIt"
	AllowFrom             string // optional: comma-separated allowed user IDs, default "*"
	GroupReplyAll         bool   // optional: reply to all group messages, default false
	ShareSessionInChannel bool // optional: share session across channel, default false
	ReplyInThread         bool   // optional: reply in thread, default false
	ThreadIsolation       bool   // optional: isolate sessions by thread, default false
	// Webhook mode (for Lark international version)
	UseWebhook     bool   // optional: use webhook mode instead of WebSocket
	Port           string // optional: webhook server port, default "8080"
	CallbackPath   string // optional: webhook callback path, default "/feishu/webhook"
	EncryptKey     string // optional: webhook encrypt key
}

// NewFeishuBot creates a new Feishu/Lark bot.
func NewFeishuBot(cfg FeishuConfig) (*Bot, error) {
	opts := make(map[string]any)
	opts["app_id"] = cfg.AppID
	opts["app_secret"] = cfg.AppSecret
	if cfg.ReactionEmoji != "" {
		opts["reaction_emoji"] = cfg.ReactionEmoji
	}
	if cfg.AllowFrom != "" {
		opts["allow_from"] = cfg.AllowFrom
	}
	opts["group_reply_all"] = cfg.GroupReplyAll
	opts["share_session_in_channel"] = cfg.ShareSessionInChannel
	opts["reply_in_thread"] = cfg.ReplyInThread
	opts["thread_isolation"] = cfg.ThreadIsolation
	if cfg.Port != "" {
		opts["port"] = cfg.Port
	}
	if cfg.CallbackPath != "" {
		opts["callback_path"] = cfg.CallbackPath
	}
	if cfg.EncryptKey != "" {
		opts["encrypt_key"] = cfg.EncryptKey
	}

	var platform core.Platform
	var err error
	if cfg.IsLark {
		platform, err = feishu.NewLark(opts)
	} else {
		platform, err = feishu.New(opts)
	}
	if err != nil {
		return nil, err
	}

	return &Bot{
		platform: platform,
	}, nil
}

// TelegramConfig contains configuration for a Telegram bot.
type TelegramConfig struct {
	Token                 string // required: Telegram bot token
	AllowFrom             string // optional: comma-separated allowed user IDs, default "*"
	GroupReplyAll         bool   // optional: reply to all group messages, default false
	ShareSessionInChannel bool   // optional: share session across channel, default false
	Proxy                 string // optional: proxy URL
	ProxyUsername         string // optional: proxy username
	ProxyPassword         string // optional: proxy password
}

// NewTelegramBot creates a new Telegram bot.
func NewTelegramBot(cfg TelegramConfig) (*Bot, error) {
	opts := make(map[string]any)
	opts["token"] = cfg.Token
	opts["allow_from"] = cfg.AllowFrom
	opts["group_reply_all"] = cfg.GroupReplyAll
	opts["share_session_in_channel"] = cfg.ShareSessionInChannel
	if cfg.Proxy != "" {
		opts["proxy"] = cfg.Proxy
	}
	if cfg.ProxyUsername != "" {
		opts["proxy_username"] = cfg.ProxyUsername
	}
	if cfg.ProxyPassword != "" {
		opts["proxy_password"] = cfg.ProxyPassword
	}

	platform, err := telegram.New(opts)
	if err != nil {
		return nil, err
	}

	return &Bot{
		platform: platform,
	}, nil
}

// DiscordConfig contains configuration for a Discord bot.
type DiscordConfig struct {
	Token                 string // required: Discord bot token
	AllowFrom             string // optional: comma-separated allowed user IDs, default "*"
	GuildID               string // optional: guild ID for instant command registration
	GroupReplyAll         bool   // optional: reply to all group messages, default false
	ShareSessionInChannel bool   // optional: share session across channel, default false
	ThreadIsolation       bool   // optional: isolate sessions by thread, default false
}

// NewDiscordBot creates a new Discord bot.
func NewDiscordBot(cfg DiscordConfig) (*Bot, error) {
	opts := make(map[string]any)
	opts["token"] = cfg.Token
	opts["allow_from"] = cfg.AllowFrom
	opts["guild_id"] = cfg.GuildID
	opts["group_reply_all"] = cfg.GroupReplyAll
	opts["share_session_in_channel"] = cfg.ShareSessionInChannel
	opts["thread_isolation"] = cfg.ThreadIsolation

	platform, err := discord.New(opts)
	if err != nil {
		return nil, err
	}

	return &Bot{
		platform: platform,
	}, nil
}

// SlackConfig contains configuration for a Slack bot.
type SlackConfig struct {
	BotToken              string // required: Slack bot token (xoxb-...)
	AppToken              string // required: Slack app token (xapp-...)
	AllowFrom             string // optional: comma-separated allowed user IDs, default "*"
	ShareSessionInChannel bool   // optional: share session across channel, default false
}

// NewSlackBot creates a new Slack bot.
func NewSlackBot(cfg SlackConfig) (*Bot, error) {
	opts := make(map[string]any)
	opts["bot_token"] = cfg.BotToken
	opts["app_token"] = cfg.AppToken
	opts["allow_from"] = cfg.AllowFrom
	opts["share_session_in_channel"] = cfg.ShareSessionInChannel

	platform, err := slack.New(opts)
	if err != nil {
		return nil, err
	}

	return &Bot{
		platform: platform,
	}, nil
}

// DingTalkConfig contains configuration for a DingTalk bot.
type DingTalkConfig struct {
	AppKey    string // required: DingTalk App Key
	AppSecret string // required: DingTalk App Secret
	AllowFrom string // optional: comma-separated allowed user IDs, default "*"
	RobotCode string // optional: robot code
}

// NewDingTalkBot creates a new DingTalk bot.
func NewDingTalkBot(cfg DingTalkConfig) (*Bot, error) {
	opts := make(map[string]any)
	opts["app_key"] = cfg.AppKey
	opts["app_secret"] = cfg.AppSecret
	opts["allow_from"] = cfg.AllowFrom
	opts["robot_code"] = cfg.RobotCode

	platform, err := dingtalk.New(opts)
	if err != nil {
		return nil, err
	}

	return &Bot{
		platform: platform,
	}, nil
}

// WeComConfig contains configuration for a WeCom (WeChat Work) bot.
type WeComConfig struct {
	CorpID    string // required: WeCom Corp ID
	AgentID   int    // required: WeCom Agent ID
	Secret    string // required: WeCom Secret
	AllowFrom string // optional: comma-separated allowed user IDs, default "*"
}

// NewWeComBot creates a new WeCom (WeChat Work) bot.
func NewWeComBot(cfg WeComConfig) (*Bot, error) {
	opts := make(map[string]any)
	opts["corp_id"] = cfg.CorpID
	opts["agent_id"] = cfg.AgentID
	opts["secret"] = cfg.Secret
	opts["allow_from"] = cfg.AllowFrom

	platform, err := wecom.New(opts)
	if err != nil {
		return nil, err
	}

	return &Bot{
		platform: platform,
	}, nil
}

// OnMessage sets the message handler for the bot.
func (b *Bot) OnMessage(handler MessageHandler) {
	b.handler = handler
}

// Start starts the bot and begins receiving messages.
func (b *Bot) Start() error {
	if b.handler == nil {
		return fmt.Errorf("message handler not set, use OnMessage() to set one")
	}

	return b.platform.Start(func(p core.Platform, msg *core.Message) {
		wrapped := &Message{
			Message:  msg,
			platform: p,
		}
		if err := b.handler(context.Background(), wrapped); err != nil {
			// Log the error but don't propagate it back to the platform
		}
	})
}

// Stop stops the bot.
func (b *Bot) Stop() error {
	return b.platform.Stop()
}

// Name returns the name of the platform.
func (b *Bot) Name() string {
	return b.platform.Name()
}

// Platform returns the underlying platform instance for advanced usage.
func (b *Bot) Platform() core.Platform {
	return b.platform
}

// NewCard creates a new card builder for rich messages.
func NewCard() *core.CardBuilder {
	return core.NewCard()
}

// PrimaryBtn creates a primary-styled button.
func PrimaryBtn(text, value string) core.CardButton {
	return core.PrimaryBtn(text, value)
}

// DefaultBtn creates a default-styled button.
func DefaultBtn(text, value string) core.CardButton {
	return core.DefaultBtn(text, value)
}

// DangerBtn creates a danger-styled button.
func DangerBtn(text, value string) core.CardButton {
	return core.DangerBtn(text, value)
}

// Btn creates a button with custom style.
func Btn(text, typ, value string) core.CardButton {
	return core.Btn(text, typ, value)
}

// CardSelectOption creates an option for a select dropdown.
func CardSelectOption(text, value string) core.CardSelectOption {
	return core.CardSelectOption{Text: text, Value: value}
}

// ListPlatforms returns the names of all registered platforms.
func ListPlatforms() []string {
	return core.ListPlatforms()
}
