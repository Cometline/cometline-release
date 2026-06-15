package gateway

import "context"

// InboundMessage is a normalized message from an external chat platform.
type InboundMessage struct {
	Platform         string
	GuildID          string
	ParentChannelID  string
	UserID           string
	ChannelID        string
	ThreadID         string
	Text             string
	Mentioned        bool
}

// TypingIndicator can show a platform-specific "typing" state while a turn runs.
type TypingIndicator interface {
	KeepTyping(ctx context.Context, channelID string) (stop func())
}

// OutboundMessage is a reply destined for an external chat platform.
type OutboundMessage struct {
	Platform  string
	UserID    string
	ChannelID string
	ThreadID  string
	Text      string
}

// PlatformAdapter connects CometMind to one messaging surface.
type PlatformAdapter interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Deliver(ctx context.Context, msg OutboundMessage) error
	SetInboundHandler(fn func(context.Context, InboundMessage))
}
