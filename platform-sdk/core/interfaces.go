// Package core defines the core interfaces and types for the platform SDK.
package core

import (
	"context"
	"errors"
)

// Platform abstracts a messaging platform (Feishu, DingTalk, Slack, etc.).
type Platform interface {
	Name() string
	Start(handler MessageHandler) error
	Reply(ctx context.Context, replyCtx any, content string) error
	Send(ctx context.Context, replyCtx any, content string) error
	Stop() error
}

// ErrNotSupported indicates a platform doesn't support a particular operation.
var ErrNotSupported = errors.New("operation not supported by this platform")

// MessageHandler is called by platforms when a new message arrives.
type MessageHandler func(p Platform, msg *Message)

// ReplyContextReconstructor is an optional interface for platforms that can
// recreate a reply context from a session key.
type ReplyContextReconstructor interface {
	ReconstructReplyCtx(sessionKey string) (any, error)
}

// TypingIndicator is an optional interface for platforms that can show a
// "processing" indicator (typing bubble, emoji reaction, etc.).
type TypingIndicator interface {
	StartTyping(ctx context.Context, replyCtx any) (stop func())
}

// ImageSender is an optional interface for platforms that support sending images.
type ImageSender interface {
	SendImage(ctx context.Context, replyCtx any, img ImageAttachment) error
}

// FileSender is an optional interface for platforms that support sending files.
type FileSender interface {
	SendFile(ctx context.Context, replyCtx any, file FileAttachment) error
}

// MessageUpdater is an optional interface for platforms that support updating messages.
type MessageUpdater interface {
	UpdateMessage(ctx context.Context, replyCtx any, content string) error
}

// ButtonOption represents a clickable inline button.
type ButtonOption struct {
	Text string // display text on the button
	Data string // callback data returned when clicked
}

// InlineButtonSender is an optional interface for platforms that support
// sending messages with clickable inline buttons.
type InlineButtonSender interface {
	SendWithButtons(ctx context.Context, replyCtx any, content string, buttons [][]ButtonOption) error
}

// CardSender is an optional interface for platforms that support sending
// structured rich cards.
type CardSender interface {
	SendCard(ctx context.Context, replyCtx any, card *Card) error
	ReplyCard(ctx context.Context, replyCtx any, card *Card) error
}

// CardNavigationHandler is called by platforms to render a card for in-place
// card updates.
type CardNavigationHandler func(action string, sessionKey string) *Card

// CardNavigable is an optional interface for platforms that support in-place
// card navigation.
type CardNavigable interface {
	SetCardNavigationHandler(h CardNavigationHandler)
}

// PreviewMessageSender is an optional interface for platforms that support
// sending a preview message that can be updated later (for streaming).
type PreviewMessageSender interface {
	SendPreviewStart(ctx context.Context, replyCtx any, content string) (any, error)
	UpdateMessage(ctx context.Context, previewHandle any, content string) error
	DeletePreviewMessage(ctx context.Context, previewHandle any) error
}

// CommandRegistrar is an optional interface for platforms that support
// registering commands to the platform's native menu.
type CommandRegistrar interface {
	RegisterCommands(commands []BotCommandInfo) error
}

