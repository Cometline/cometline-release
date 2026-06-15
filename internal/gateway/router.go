package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cometline/cometmind/internal/config"
	"github.com/cometline/cometmind/internal/event"
	"github.com/cometline/cometmind/internal/session"
)

// Runner executes agent turns for gateway inbound messages.
type Runner interface {
	RunTurn(ctx context.Context, sess session.Session, workspacePath, text string, onEvent func(event.Event)) error
}

// Router maps platform identities to CometMind sessions and runs turns.
type Router struct {
	Sessions  *session.Service
	Config    *config.Config
	Runner    Runner
	Typing    TypingIndicator
	onReply   func(context.Context, OutboundMessage) error
}

// SetReplyHandler registers the callback used to deliver outbound messages.
func (r *Router) SetReplyHandler(fn func(context.Context, OutboundMessage) error) {
	r.onReply = fn
}

// HandleInbound routes one external message through the CometMind runtime.
func (r *Router) HandleInbound(ctx context.Context, msg InboundMessage) error {
	if r == nil || r.Sessions == nil || r.Runner == nil {
		return fmt.Errorf("gateway router is not configured")
	}
	if !r.allowed(msg) {
		if reason := r.blockReason(msg); reason != "" {
			log.Printf("discord: ignoring message from user=%s channel=%s: %s", msg.UserID, msg.ChannelID, reason)
		}
		return nil
	}

	wsPath := r.Config.Gateway.Discord.WorkspacePath
	if wsPath == "" {
		return fmt.Errorf("gateway workspace_path is not configured")
	}
	ws, err := r.Sessions.EnsureWorkspace(ctx, wsPath)
	if err != nil {
		return err
	}

	sessID, err := r.resolveSession(ctx, msg, ws)
	if err != nil {
		return err
	}
	sess, err := r.Sessions.GetSession(ctx, sessID)
	if err != nil {
		return err
	}

	if _, err := r.Sessions.AppendUserMessageAndMaybeTitle(ctx, sess.ID, msg.Text); err != nil {
		return err
	}

	if r.Typing != nil {
		stopTyping := r.Typing.KeepTyping(ctx, msg.ChannelID)
		defer stopTyping()
	}

	log.Printf("discord: running agent turn session=%s", sess.ID)
	var reply strings.Builder
	err = r.Runner.RunTurn(ctx, sess, ws.Path, msg.Text, func(ev event.Event) {
		switch ev.Kind {
		case event.KindTextDelta:
			reply.WriteString(ev.Delta)
		case event.KindError:
			if ev.Message != "" {
				reply.WriteString("\n[error] ")
				reply.WriteString(ev.Message)
				reply.WriteByte('\n')
			}
		case event.KindSubagentProgress:
			if ev.ProgressText != "" {
				reply.WriteString("\n[subagent] ")
				reply.WriteString(ev.ProgressText)
				reply.WriteByte('\n')
			}
		case event.KindSubagentFinished:
			if ev.Summary != "" {
				reply.WriteString("\n[subagent done] ")
				reply.WriteString(ev.Summary)
				reply.WriteByte('\n')
			}
		}
	})
	var text string
	if err != nil {
		text = fmt.Sprintf("Error: %v", err)
		log.Printf("discord: agent turn failed user=%s channel=%s: %v", msg.UserID, msg.ChannelID, err)
	} else {
		text = strings.TrimSpace(reply.String())
		if text == "" {
			text = "(no response)"
		}
	}
	if r.onReply != nil {
		log.Printf("discord: replying to channel=%s (%d bytes)", msg.ChannelID, len(text))
		return r.onReply(ctx, OutboundMessage{
			Platform:  msg.Platform,
			UserID:    msg.UserID,
			ChannelID: msg.ChannelID,
			ThreadID:  msg.ThreadID,
			Text:      text,
		})
	}
	return nil
}

func (r *Router) resolveSession(ctx context.Context, msg InboundMessage, ws session.Workspace) (string, error) {
	mapped, err := r.Sessions.LookupGatewaySession(ctx, msg.Platform, msg.UserID, msg.ChannelID, msg.ThreadID)
	if err == nil {
		return mapped.CometmindSessionID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	modelID, providerID := r.discordSessionModel()
	sess, err := r.Sessions.NewSession(ctx, ws.ID, modelID, providerID)
	if err != nil {
		return "", err
	}
	if _, err := r.Sessions.UpsertGatewaySession(ctx, msg.Platform, msg.UserID, msg.ChannelID, msg.ThreadID, sess.ID, ws.ID); err != nil {
		return "", err
	}
	return sess.ID, nil
}

func (r *Router) allowed(msg InboundMessage) bool {
	return r.blockReason(msg) == ""
}

func (r *Router) blockReason(msg InboundMessage) string {
	cfg := r.Config.Gateway.Discord
	if msg.Platform == "discord" && cfg.RequireMention && !msg.Mentioned {
		return "mention required"
	}
	if len(cfg.AllowedUsers) > 0 && !contains(cfg.AllowedUsers, msg.UserID) {
		return "user not in allowed_users"
	}
	// Guild channel allowlist only; DMs use per-user channel IDs that won't match guild channels.
	// Thread channels inherit access from their parent channel ID.
	if len(cfg.AllowedChannels) > 0 && msg.GuildID != "" {
		if !contains(cfg.AllowedChannels, msg.ChannelID) &&
			!contains(cfg.AllowedChannels, msg.ParentChannelID) {
			return "channel not in allowed_channels"
		}
	}
	return ""
}

func (r *Router) discordSessionModel() (modelID, providerID string) {
	cfg := r.Config.Gateway.Discord
	modelID = strings.TrimSpace(cfg.Model)
	providerID = strings.TrimSpace(cfg.Provider)
	if modelID == "" {
		modelID = r.Config.Model
	}
	if providerID == "" {
		providerID = r.Config.Provider
	}
	return modelID, providerID
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
