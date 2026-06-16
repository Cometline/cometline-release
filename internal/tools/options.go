package tools

import (
	"github.com/cometline/cometmind/internal/acp"
	"github.com/cometline/cometmind/internal/session"
)

// RegistryOptions configures optional registry capabilities.
type RegistryOptions struct {
	Sessions *session.Service
	ACP      acp.Config
	ACPMgr   *acp.SessionManager
}
