package core

import (
	"log/slog"
	"strings"
)

// ImageAttachment represents an image sent by the user.
type ImageAttachment struct {
	MimeType string // e.g. "image/png", "image/jpeg"
	Data     []byte // raw image bytes
	FileName string // original filename (optional)
}

// FileAttachment represents a file (PDF, doc, spreadsheet, etc.) sent by the user.
type FileAttachment struct {
	MimeType string // e.g. "application/pdf", "text/plain"
	Data     []byte // raw file bytes
	FileName string // original filename
}

// AudioAttachment represents a voice/audio message sent by the user.
type AudioAttachment struct {
	MimeType string // e.g. "audio/amr", "audio/ogg", "audio/mp4"
	Data     []byte // raw audio bytes
	Format   string // short format hint: "amr", "ogg", "m4a", "mp3", "wav", etc.
	Duration int    // duration in seconds (if known)
}

// Message represents a unified incoming message from any platform.
type Message struct {
	SessionKey string          // unique key for user context, e.g. "feishu:{chatID}:{userID}"
	Platform   string          // platform name
	MessageID  string          // platform message ID for tracing
	UserID     string          // user ID
	UserName   string          // user display name
	ChatName   string          // human-readable chat/group name (optional)
	Content    string          // message content
	Images     []ImageAttachment // attached images (if any)
	Files      []FileAttachment  // attached files (if any)
	Audio      *AudioAttachment  // voice message (if any)
	ReplyCtx   any             // platform-specific context needed for replying
	FromVoice  bool            // true if message originated from voice transcription
}

// BotCommandInfo represents a command for bot menu registration.
type BotCommandInfo struct {
	Command     string // command name without leading "/"
	Description string // short description for the menu
}

// CheckAllowFrom logs a security warning when allow_from is not configured.
func CheckAllowFrom(platform, allowFrom string) {
	if strings.TrimSpace(allowFrom) == "" {
		slog.Warn("allow_from is not set — all users are permitted. "+
			"Set allow_from to restrict access.",
			"platform", platform)
	}
}

// AllowList checks whether a user ID is permitted based on a comma-separated
// allow_from string. Returns true if allowFrom is empty or "*" (allow all),
// or if the userID is in the list. Comparison is case-insensitive.
func AllowList(allowFrom, userID string) bool {
	allowFrom = strings.TrimSpace(allowFrom)
	if allowFrom == "" || allowFrom == "*" {
		return true
	}
	for _, id := range strings.Split(allowFrom, ",") {
		if strings.EqualFold(strings.TrimSpace(id), userID) {
			return true
		}
	}
	return false
}

// RedactToken replaces a secret token in text with [REDACTED].
func RedactToken(text, token string) string {
	if token == "" || text == "" {
		return text
	}
	return strings.ReplaceAll(text, token, "[REDACTED]")
}
