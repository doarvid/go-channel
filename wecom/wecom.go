package wecom

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/chenhg5/cc-connect/platform-sdk/core"
)

func init() {
	core.RegisterPlatform("wecom", New)
}

type replyContext struct {
	chatType string
	userID   string
	chatID   string
}

type Platform struct {
	corpID              string
	agentID             int
	secret              string
	allowFrom           string
	handler             core.MessageHandler
	cancel              context.CancelFunc
}

func New(opts map[string]any) (core.Platform, error) {
	corpID, _ := opts["corp_id"].(string)
	secret, _ := opts["secret"].(string)
	allowFrom, _ := opts["allow_from"].(string)
	agentID, _ := opts["agent_id"].(int)
	if agentID == 0 {
		if a, ok := opts["agent_id"].(int64); ok {
			agentID = int(a)
		}
		if a, ok := opts["agent_id"].(string); ok {
			fmt.Sscanf(a, "%d", &agentID)
		}
	}
	core.CheckAllowFrom("wecom", allowFrom)
	if corpID == "" || secret == "" {
		return nil, fmt.Errorf("wecom: corp_id and secret are required")
	}
	return &Platform{
		corpID:    corpID,
		agentID:   agentID,
		secret:    secret,
		allowFrom: allowFrom,
	}, nil
}

func (p *Platform) Name() string { return "wecom" }

func (p *Platform) Start(handler core.MessageHandler) error {
	p.handler = handler
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	slog.Info("wecom: platform initialized (webhook mode)")

	go func() {
		<-ctx.Done()
	}()

	return nil
}

func (p *Platform) Reply(ctx context.Context, rctx any, content string) error {
	rc, ok := rctx.(replyContext)
	if !ok {
		return fmt.Errorf("wecom: invalid reply context type %T", rctx)
	}
	slog.Debug("wecom: reply (stub)", "chat_type", rc.chatType, "content_len", len(content))
	return nil
}

func (p *Platform) Send(ctx context.Context, rctx any, content string) error {
	rc, ok := rctx.(replyContext)
	if !ok {
		return fmt.Errorf("wecom: invalid reply context type %T", rctx)
	}
	slog.Debug("wecom: send (stub)", "chat_type", rc.chatType, "content_len", len(content))
	return nil
}

func (p *Platform) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

var _ core.Platform = (*Platform)(nil)
