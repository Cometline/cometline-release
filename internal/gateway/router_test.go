package gateway

import (
	"testing"

	"github.com/cometline/cometmind/internal/config"
)

func TestRouterAllowed(t *testing.T) {
	t.Parallel()
	r := &Router{
		Config: &config.Config{
			Gateway: config.GatewayConfig{
				Discord: config.DiscordGatewayConfig{
					AllowedUsers:    []string{"user-1"},
					AllowedChannels: []string{"chan-1"},
					RequireMention:  true,
				},
			},
		},
	}

	if r.allowed(InboundMessage{Platform: "discord", GuildID: "guild-1", UserID: "user-1", ChannelID: "chan-1", Mentioned: true}) != true {
		t.Fatal("expected allowed mention")
	}
	if r.allowed(InboundMessage{Platform: "discord", GuildID: "guild-1", UserID: "user-1", ChannelID: "thread-1", ParentChannelID: "chan-1", Mentioned: true}) != true {
		t.Fatal("expected thread allowed via parent channel")
	}
	if r.allowed(InboundMessage{Platform: "discord", GuildID: "guild-1", UserID: "user-1", ChannelID: "chan-1", Mentioned: false}) != false {
		t.Fatal("expected blocked without mention")
	}
	if r.allowed(InboundMessage{Platform: "discord", GuildID: "", UserID: "user-1", ChannelID: "dm-chan", Mentioned: true}) != true {
		t.Fatal("expected DM allowed without channel allowlist match")
	}
	if r.allowed(InboundMessage{Platform: "discord", GuildID: "guild-1", UserID: "other", ChannelID: "chan-1", Mentioned: true}) != false {
		t.Fatal("expected blocked user")
	}
}
