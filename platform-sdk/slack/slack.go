package slack

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chenhg5/cc-connect/platform-sdk/core"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func init() {
	core.RegisterPlatform("slack", New)
}

type replyContext struct {
	channel   string
	timestamp string
}

type Platform struct {
	botToken              string
	appToken              string
	allowFrom             string
	shareSessionInChannel bool
	client                *slack.Client
	socket                *socketmode.Client
	handler               core.MessageHandler
	cancel                context.CancelFunc
	channelNameCache      map[string]string
	channelCacheMu        sync.RWMutex
	userNameCache         sync.Map
}

func New(opts map[string]any) (core.Platform, error) {
	botToken, _ := opts["bot_token"].(string)
	appToken, _ := opts["app_token"].(string)
	allowFrom, _ := opts["allow_from"].(string)
	core.CheckAllowFrom("slack", allowFrom)
	shareSessionInChannel, _ := opts["share_session_in_channel"].(bool)
	if botToken == "" || appToken == "" {
		return nil, fmt.Errorf("slack: bot_token and app_token are required")
	}
	return &Platform{
		botToken:              botToken,
		appToken:              appToken,
		allowFrom:             allowFrom,
		shareSessionInChannel: shareSessionInChannel,
		channelNameCache:      make(map[string]string),
	}, nil
}

func (p *Platform) Name() string { return "slack" }

func (p *Platform) Start(handler core.MessageHandler) error {
	p.handler = handler

	p.client = slack.New(p.botToken, slack.OptionAppLevelToken(p.appToken))
	p.socket = socketmode.New(p.client)

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-p.socket.Events:
				p.handleEvent(evt)
			}
		}
	}()

	go func() {
		if err := p.socket.RunContext(ctx); err != nil {
			slog.Error("slack: socket mode error", "error", err)
		}
	}()

	slog.Info("slack: socket mode connected")
	return nil
}

func (p *Platform) handleEvent(evt socketmode.Event) {
	slog.Debug("slack: raw event received", "type", evt.Type)
	switch evt.Type {
	case socketmode.EventTypeEventsAPI:
		data, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return
		}
		if evt.Request != nil {
			p.socket.Ack(*evt.Request)
		}

		if data.Type == slackevents.CallbackEvent {
			switch ev := data.InnerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				if ev.BotID != "" || ev.User == "" {
					return
				}
				p.handleMessage(ev.User, ev.Channel, ev.TimeStamp, stripAppMentionText(ev.Text))
			case *slackevents.MessageEvent:
				if ev.BotID != "" || ev.User == "" {
					return
				}
				if ev.ChannelType == "im" {
					p.handleMessage(ev.User, ev.Channel, ev.TimeStamp, ev.Text)
				}
			}
		}
	}
}

func (p *Platform) handleMessage(userID, channelID, ts, text string) {
	if ts != "" {
		if dotIdx := strings.IndexByte(ts, '.'); dotIdx > 0 {
			if sec, err := strconv.ParseInt(ts[:dotIdx], 10, 64); err == nil {
				if core.IsOldMessage(time.Unix(sec, 0)) {
					slog.Debug("slack: ignoring old message after restart", "ts", ts)
					return
				}
			}
		}
	}

	if !core.AllowList(p.allowFrom, userID) {
		slog.Debug("slack: message from unauthorized user", "user", userID)
		return
	}

	var sessionKey string
	if p.shareSessionInChannel {
		sessionKey = fmt.Sprintf("slack:%s", channelID)
	} else {
		sessionKey = fmt.Sprintf("slack:%s:%s", channelID, userID)
	}

	msg := &core.Message{
		SessionKey: sessionKey, Platform: "slack",
		UserID:    userID,
		UserName:  p.resolveUserName(userID),
		Content:   text,
		MessageID: ts,
		ReplyCtx:  replyContext{channel: channelID, timestamp: ts},
	}
	if msg.Content == "" {
		return
	}
	p.handler(p, msg)
}

func (p *Platform) Reply(ctx context.Context, rctx any, content string) error {
	rc, ok := rctx.(replyContext)
	if !ok {
		return fmt.Errorf("slack: invalid reply context type %T", rctx)
	}
	opts := []slack.MsgOption{
		slack.MsgOptionText(content, false),
		slack.MsgOptionTS(rc.timestamp),
	}
	_, _, err := p.client.PostMessage(rc.channel, opts...)
	if err != nil {
		return fmt.Errorf("slack: reply: %w", err)
	}
	return nil
}

func (p *Platform) Send(ctx context.Context, rctx any, content string) error {
	rc, ok := rctx.(replyContext)
	if !ok {
		return fmt.Errorf("slack: invalid reply context type %T", rctx)
	}
	opts := []slack.MsgOption{
		slack.MsgOptionText(content, false),
	}
	_, _, err := p.client.PostMessage(rc.channel, opts...)
	if err != nil {
		return fmt.Errorf("slack: send: %w", err)
	}
	return nil
}

func (p *Platform) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *Platform) resolveUserName(userID string) string {
	if v, ok := p.userNameCache.Load(userID); ok {
		return v.(string)
	}
	user, err := p.client.GetUserInfo(userID)
	if err != nil {
		return userID
	}
	name := user.Name
	if user.Profile.DisplayName != "" {
		name = user.Profile.DisplayName
	}
	p.userNameCache.Store(userID, name)
	return name
}

func stripAppMentionText(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "<@") {
		if idx := strings.Index(text, ">"); idx > 0 {
			text = strings.TrimSpace(text[idx+1:])
		}
	}
	return text
}

var _ core.Platform = (*Platform)(nil)
