package dingtalk

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/chenhg5/cc-connect/platform-sdk/core"
)

func init() {
	core.RegisterPlatform("dingtalk", New)
}

type replyContext struct {
	conversationID string
	senderStaffID  string
}

type Platform struct {
	appKey              string
	appSecret           string
	allowFrom           string
	robotCode           string
	handler             core.MessageHandler
	cancel              context.CancelFunc
}

func New(opts map[string]any) (core.Platform, error) {
	appKey, _ := opts["app_key"].(string)
	appSecret, _ := opts["app_secret"].(string)
	allowFrom, _ := opts["allow_from"].(string)
	robotCode, _ := opts["robot_code"].(string)
	core.CheckAllowFrom("dingtalk", allowFrom)
	if appKey == "" || appSecret == "" {
		return nil, fmt.Errorf("dingtalk: app_key and app_secret are required")
	}
	return &Platform{
		appKey:     appKey,
		appSecret:  appSecret,
		allowFrom:  allowFrom,
		robotCode:  robotCode,
	}, nil
}

func (p *Platform) Name() string { return "dingtalk" }

func (p *Platform) Start(handler core.MessageHandler) error {
	p.handler = handler
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	slog.Info("dingtalk: platform initialized (webhook mode)")

	go func() {
		<-ctx.Done()
	}()

	return nil
}

func (p *Platform) Reply(ctx context.Context, rctx any, content string) error {
	rc, ok := rctx.(replyContext)
	if !ok {
		return fmt.Errorf("dingtalk: invalid reply context type %T", rctx)
	}
	slog.Debug("dingtalk: reply (stub)", "conversation", rc.conversationID, "content_len", len(content))
	return nil
}

func (p *Platform) Send(ctx context.Context, rctx any, content string) error {
	rc, ok := rctx.(replyContext)
	if !ok {
		return fmt.Errorf("dingtalk: invalid reply context type %T", rctx)
	}
	slog.Debug("dingtalk: send (stub)", "conversation", rc.conversationID, "content_len", len(content))
	return nil
}

func (p *Platform) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

var _ core.Platform = (*Platform)(nil)
