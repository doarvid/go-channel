package discord

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chenhg5/cc-connect/platform-sdk/core"

	"github.com/bwmarrin/discordgo"
)

func init() {
	core.RegisterPlatform("discord", New)
}

const maxDiscordLen = 2000

type replyContext struct {
	channelID string
	messageID string
	threadID  string
}

type interactionReplyCtx struct {
	interaction *discordgo.Interaction
	channelID   string
	mu          sync.Mutex
	firstDone   bool
}

type Platform struct {
	token                 string
	allowFrom             string
	guildID               string
	groupReplyAll         bool
	shareSessionInChannel bool
	threadIsolation       bool
	session               *discordgo.Session
	handler               core.MessageHandler
	botID                 string
	appID                 string
	channelNameCache      sync.Map
	botRoleIDs            sync.Map
	readyCh               chan struct{}
	seenMsgs              sync.Map
}

func New(opts map[string]any) (core.Platform, error) {
	token, _ := opts["token"].(string)
	if token == "" {
		return nil, fmt.Errorf("discord: token is required")
	}
	allowFrom, _ := opts["allow_from"].(string)
	core.CheckAllowFrom("discord", allowFrom)
	guildID, _ := opts["guild_id"].(string)
	groupReplyAll, _ := opts["group_reply_all"].(bool)
	shareSessionInChannel, _ := opts["share_session_in_channel"].(bool)
	threadIsolation, _ := opts["thread_isolation"].(bool)
	return &Platform{
		token:                 token,
		allowFrom:             allowFrom,
		guildID:               guildID,
		groupReplyAll:         groupReplyAll,
		shareSessionInChannel: shareSessionInChannel,
		readyCh:               make(chan struct{}),
		threadIsolation:       threadIsolation,
	}, nil
}

func (p *Platform) Name() string { return "discord" }

func (p *Platform) Start(handler core.MessageHandler) error {
	p.handler = handler

	session, err := discordgo.New("Bot " + p.token)
	if err != nil {
		return fmt.Errorf("discord: create session: %w", err)
	}
	p.session = session

	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentMessageContent

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		p.botID = r.User.ID
		p.appID = r.User.ID
		slog.Info("discord: connected", "bot", r.User.Username+"#"+r.User.Discriminator)
		select {
		case <-p.readyCh:
		default:
			close(p.readyCh)
		}
	})

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if _, loaded := p.seenMsgs.LoadOrStore(m.ID, struct{}{}); loaded {
			slog.Debug("discord: ignoring duplicate message", "msg_id", m.ID)
			return
		}
		time.AfterFunc(2*time.Minute, func() { p.seenMsgs.Delete(m.ID) })

		if m.Author.Bot || m.Author.ID == p.botID {
			return
		}
		if core.IsOldMessage(m.Timestamp) {
			slog.Debug("discord: ignoring old message after restart", "timestamp", m.Timestamp)
			return
		}
		if !core.AllowList(p.allowFrom, m.Author.ID) {
			slog.Debug("discord: message from unauthorized user", "user", m.Author.ID)
			return
		}

		slog.Debug("discord: message received", "user", m.Author.Username, "channel", m.ChannelID)

		sessionKey := p.makeSessionKey(m.ChannelID, m.Author.ID)
		rctx := replyContext{channelID: m.ChannelID, messageID: m.ID}

		var images []core.ImageAttachment
		var audio *core.AudioAttachment
		for _, att := range m.Attachments {
			ct := strings.ToLower(att.ContentType)
			if strings.HasPrefix(ct, "audio/") {
				data, err := downloadURL(att.URL)
				if err != nil {
					slog.Error("discord: download audio failed", "url", att.URL, "error", err)
					continue
				}
				format := "ogg"
				if parts := strings.SplitN(ct, "/", 2); len(parts) == 2 {
					format = parts[1]
				}
				audio = &core.AudioAttachment{
					MimeType: ct, Data: data, Format: format,
				}
			} else if att.Width > 0 && att.Height > 0 {
				data, err := downloadURL(att.URL)
				if err != nil {
					slog.Error("discord: download attachment failed", "url", att.URL, "error", err)
					continue
				}
				images = append(images, core.ImageAttachment{
					MimeType: att.ContentType, Data: data, FileName: att.Filename,
				})
			}
		}

		if m.Content == "" && len(images) == 0 && audio == nil {
			return
		}

		msg := &core.Message{
			SessionKey: sessionKey, Platform: "discord",
			MessageID: m.ID,
			UserID:    m.Author.ID, UserName: m.Author.Username,
			Content: m.Content, Images: images, Audio: audio, ReplyCtx: rctx,
		}
		p.handler(p, msg)
	})

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		p.handleInteraction(s, i)
	})

	if err := session.Open(); err != nil {
		return fmt.Errorf("discord: open gateway: %w", err)
	}

	return nil
}

func (p *Platform) makeSessionKey(channelID string, userID string) string {
	if p.shareSessionInChannel {
		return fmt.Sprintf("discord:%s", channelID)
	}
	return fmt.Sprintf("discord:%s:%s", channelID, userID)
}

func (p *Platform) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	userID, userName := "", ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
		userName = i.Member.User.Username
	} else if i.User != nil {
		userID = i.User.ID
		userName = i.User.Username
	}

	if !core.AllowList(p.allowFrom, userID) {
		slog.Debug("discord: interaction from unauthorized user", "user", userID)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You are not authorized to use this bot.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		slog.Error("discord: defer interaction failed", "error", err)
		return
	}

	data := i.ApplicationCommandData()
	cmdText := "/" + data.Name
	for _, opt := range data.Options {
		cmdText += " " + opt.StringValue()
	}
	channelID := i.ChannelID

	slog.Debug("discord: slash command", "user", userName, "command", cmdText, "channel", channelID)

	sessionKey := p.makeSessionKey(channelID, userID)
	ictx := &interactionReplyCtx{
		interaction: i.Interaction,
		channelID:   channelID,
	}

	msg := &core.Message{
		SessionKey: sessionKey, Platform: "discord",
		MessageID: i.ID,
		UserID:    userID, UserName: userName,
		Content: cmdText, ReplyCtx: ictx,
	}
	p.handler(p, msg)
}

func (p *Platform) Reply(ctx context.Context, rctx any, content string) error {
	switch rc := rctx.(type) {
	case *interactionReplyCtx:
		return p.sendInteraction(rc, content)
	case replyContext:
		return p.sendChannelReply(rc, content)
	default:
		return fmt.Errorf("discord: invalid reply context type %T", rctx)
	}
}

func (p *Platform) Send(ctx context.Context, rctx any, content string) error {
	switch rc := rctx.(type) {
	case *interactionReplyCtx:
		return p.sendInteraction(rc, content)
	case replyContext:
		return p.sendChannel(rc, content)
	default:
		return fmt.Errorf("discord: invalid reply context type %T", rctx)
	}
}

func (p *Platform) sendInteraction(ictx *interactionReplyCtx, content string) error {
	chunks := core.SplitMessageCodeFenceAware(content, maxDiscordLen)
	for _, chunk := range chunks {
		ictx.mu.Lock()
		first := !ictx.firstDone
		if first {
			ictx.firstDone = true
		}
		ictx.mu.Unlock()

		var err error
		if first {
			c := chunk
			_, err = p.session.InteractionResponseEdit(ictx.interaction, &discordgo.WebhookEdit{Content: &c})
		} else {
			_, err = p.session.FollowupMessageCreate(ictx.interaction, true, &discordgo.WebhookParams{Content: chunk})
		}

		if err != nil {
			slog.Warn("discord: interaction response failed, falling back to channel message", "error", err)
			_, err = p.session.ChannelMessageSend(ictx.channelID, chunk)
			if err != nil {
				return fmt.Errorf("discord: send fallback: %w", err)
			}
		}
	}
	return nil
}

func (p *Platform) sendChannelReply(rc replyContext, content string) error {
	chunks := core.SplitMessageCodeFenceAware(content, maxDiscordLen)
	for _, chunk := range chunks {
		var err error
		ref := &discordgo.MessageReference{MessageID: rc.messageID}
		_, err = p.session.ChannelMessageSendReply(rc.channelID, chunk, ref)
		if err != nil {
			return fmt.Errorf("discord: send: %w", err)
		}
	}
	return nil
}

func (p *Platform) sendChannel(rc replyContext, content string) error {
	chunks := core.SplitMessageCodeFenceAware(content, maxDiscordLen)
	for _, chunk := range chunks {
		_, err := p.session.ChannelMessageSend(rc.channelID, chunk)
		if err != nil {
			return fmt.Errorf("discord: send: %w", err)
		}
	}
	return nil
}

func (p *Platform) SendImage(ctx context.Context, rctx any, img core.ImageAttachment) error {
	name := img.FileName
	if name == "" {
		name = "image.png"
	}

	newFile := func() *discordgo.File {
		return &discordgo.File{
			Name:        name,
			ContentType: img.MimeType,
			Reader:      bytes.NewReader(img.Data),
		}
	}

	switch rc := rctx.(type) {
	case *interactionReplyCtx:
		rc.mu.Lock()
		first := !rc.firstDone
		if first {
			rc.firstDone = true
		}
		rc.mu.Unlock()

		var err error
		if first {
			_, err = p.session.InteractionResponseEdit(rc.interaction, &discordgo.WebhookEdit{
				Files: []*discordgo.File{newFile()},
			})
		} else {
			_, err = p.session.FollowupMessageCreate(rc.interaction, true, &discordgo.WebhookParams{
				Files: []*discordgo.File{newFile()},
			})
		}
		if err != nil {
			slog.Warn("discord: interaction image failed, falling back to channel message", "error", err)
			_, err = p.session.ChannelMessageSendComplex(rc.channelID, &discordgo.MessageSend{
				Files: []*discordgo.File{newFile()},
			})
			if err != nil {
				return fmt.Errorf("discord: send image fallback: %w", err)
			}
		}
		return nil
	case replyContext:
		_, err := p.session.ChannelMessageSendComplex(rc.channelID, &discordgo.MessageSend{
			Files: []*discordgo.File{newFile()},
		})
		if err != nil {
			return fmt.Errorf("discord: send image: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("discord: SendImage: invalid reply context type %T", rctx)
	}
}

func (p *Platform) Stop() error {
	if p.session != nil {
		return p.session.Close()
	}
	return nil
}

func downloadURL(u string) ([]byte, error) {
	resp, err := core.HTTPClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d", u, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

var _ core.Platform = (*Platform)(nil)
var _ core.ImageSender = (*Platform)(nil)
