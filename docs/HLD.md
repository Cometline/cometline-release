# Cometline — High-Level Design Document

> **CometCode** (Go runtime & CLI) + **Lunar** (Svelte/TypeScript GUI)

| Field       | Value                        |
|-------------|------------------------------|
| Version     | 0.1.0 (Draft)                |
| Authors     | Cometline Team               |
| Status      | Proposed                     |
| Last Updated| 2026-02-17                   |

---

## Table of Contents

1. [Background](#1-background)
2. [Functional Requirements](#2-functional-requirements)
3. [Non-Functional Requirements](#3-non-functional-requirements)
4. [Tech Stack](#4-tech-stack)

---

## 1. Background

### 1.1 Problem Statement

Existing AI coding assistants are predominantly cloud-hosted SaaS products or tightly coupled IDE extensions. Developers who value privacy, offline capability, and full control over their workflow lack a credible local-first alternative that:

- Runs entirely on the developer's machine with no mandatory cloud dependency.
- Separates the **AI runtime** from the **user interface**, allowing terminal-native and GUI workflows to coexist.
- Treats sessions, context, and tool execution as first-class primitives rather than afterthoughts.
- Supports multiple LLM providers (OpenAI, Anthropic, local models) without vendor lock-in.

### 1.2 Project Vision

**Cometline** is a local-first AI coding agent platform. It is composed of two independently deployable projects:

| Project       | Role                                                                 |
|---------------|----------------------------------------------------------------------|
| **CometCode** | Go-based daemon + CLI + TUI — the runtime that manages sessions, executes tools, and communicates with LLMs. |
| **Lunar**     | Svelte 5 / TypeScript frontend — a rich GUI served as a web app (SvelteKit) or desktop app (Electron). |

The daemon runs in the background, exposes a REST + WebSocket API, and any number of clients (CLI, TUI, Lunar, third-party) can connect to it.

### 1.3 Target Users

- **Professional developers** who want an AI pair-programmer that respects their local environment.
- **Privacy-conscious teams** who cannot send code to third-party cloud services.
- **Power users** who prefer terminal workflows but want the option of a GUI when useful.
- **Open-source contributors** who want to extend or self-host the agent.

### 1.4 Design Principles

| Principle               | Description                                                                                  |
|--------------------------|----------------------------------------------------------------------------------------------|
| **Local-first**          | All data (sessions, config, embeddings) lives on disk. Cloud is opt-in, never required.     |
| **Daemon-based**         | A long-running process holds state, manages connections, and survives client disconnects.    |
| **Separation of concerns** | Runtime logic (CometCode) is fully decoupled from presentation (Lunar, CLI, TUI).         |
| **Extensibility**        | Custom tools, new LLM providers, and plugins can be added without forking core code.        |
| **Convention over configuration** | Sensible defaults; zero config to get started, deep config when needed.              |

### 1.5 Architecture Overview

```
┌────────────────────────────────────────────────────────────┐
│                        Developer                           │
│                                                            │
│   ┌──────────┐   ┌──────────┐   ┌────────────────────┐    │
│   │  CLI     │   │  TUI     │   │  Lunar (GUI)       │    │
│   │ (Cobra)  │   │(BubbleTea│   │  Web / Electron    │    │
│   └────┬─────┘   └────┬─────┘   └────────┬───────────┘    │
│        │              │                   │                │
│        └──────────────┼───────────────────┘                │
│                       │                                    │
│              REST + WebSocket API                          │
│                       │                                    │
│        ┌──────────────┴──────────────────┐                 │
│        │       CometCode Daemon          │                 │
│        │                                 │                 │
│        │  ┌───────────┐  ┌────────────┐  │                 │
│        │  │  Session   │  │   Tool     │  │                 │
│        │  │  Manager   │  │  Executor  │  │                 │
│        │  └───────────┘  └────────────┘  │                 │
│        │  ┌───────────┐  ┌────────────┐  │                 │
│        │  │ LLM Router│  │ Workspace  │  │                 │
│        │  │ (Provider) │  │  Manager   │  │                 │
│        │  └─────┬─────┘  └────────────┘  │                 │
│        │        │                        │                 │
│        │  ┌─────┴─────────────────────┐  │                 │
│        │  │     SQLite (sqlc)         │  │                 │
│        │  └───────────────────────────┘  │                 │
│        └─────────────────────────────────┘                 │
│                       │                                    │
│            ┌──────────┴──────────┐                         │
│            │   LLM Providers     │                         │
│            │ OpenAI · Anthropic  │                         │
│            │ Ollama · Custom     │                         │
│            └─────────────────────┘                         │
└────────────────────────────────────────────────────────────┘
```

### 1.6 Scope

#### In-Scope (Phase 1)

- CometCode daemon with REST + WebSocket API.
- CLI commands for workspace, session, and daemon management.
- TUI chat interface (Bubble Tea).
- Lunar web GUI with session, chat, code, and settings views.
- Lunar Electron wrapper for desktop distribution.
- SQLite persistence for sessions, messages, and config.
- Built-in tools: file read/write, shell exec, search, diff.
- Multi-provider LLM support (OpenAI, Anthropic).

#### Out-of-Scope (Phase 1)

- Collaborative / multi-user sessions.
- Cloud sync or remote daemon hosting.
- Fine-tuning or training pipelines.
- Marketplace or registry for community tools.
- Mobile clients.

---

## 2. Functional Requirements

### 2.1 CometCode (Go Runtime)

#### 2.1.1 Daemon Lifecycle

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| D-01 | The daemon starts as a background process, binding to a configurable host and port (default `127.0.0.1:7700`). |
| D-02 | `cometcode daemon start` launches the daemon; `cometcode daemon stop` sends a graceful shutdown signal. |
| D-03 | `cometcode daemon status` reports PID, uptime, connected clients, and active sessions. |
| D-04 | On startup the daemon writes a PID file to `~/.cometcode/daemon.pid` and removes it on shutdown. |
| D-05 | If the port is already in use, the daemon exits with a clear error and suggests `--port`. |

#### 2.1.2 Session Management

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| S-01 | A **session** encapsulates a conversation between the user and the agent, scoped to a workspace. |
| S-02 | Sessions can be **created**, **loaded**, **resumed**, **archived**, and **exported** (JSON/Markdown). |
| S-03 | Each session has a unique ID (ULID), a human-readable title, and a creation timestamp. |
| S-04 | Session state includes: messages, tool call history, token usage, model config, and workspace reference. |
| S-05 | Archived sessions are read-only and can be searched by title, date range, or full-text content. |
| S-06 | `cometcode session list` shows active and recent sessions; `cometcode session resume <id>` reconnects. |

#### 2.1.3 Workspace Management

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| W-01 | A **workspace** maps to a directory on the filesystem. It is the scope for file operations and tool execution. |
| W-02 | `cometcode init` initializes a workspace in the current directory, creating `.cometcode/` with config and DB. |
| W-03 | File operations (read, write, list, search) are sandboxed to the workspace root unless explicitly overridden. |
| W-04 | Git integration: the daemon detects `.git` and exposes status, diff, log, and commit through tools. |
| W-05 | Workspace config (`.cometcode/config.toml`) stores per-project model preferences, tool allow-lists, and ignore patterns. |

#### 2.1.4 LLM Integration

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| L-01 | The LLM router accepts a provider name + model ID and dispatches requests to the correct backend. |
| L-02 | Supported providers at launch: **OpenAI** (GPT-4o, o3), **Anthropic** (Claude Sonnet 4, Opus 4.5). |
| L-03 | Providers are configured via `~/.cometcode/providers.toml` with API keys, base URLs, and default models. |
| L-04 | All LLM responses are **streamed** token-by-token over WebSocket to connected clients. |
| L-05 | Token usage (prompt + completion) is tracked per message and aggregated per session. |
| L-06 | A **system prompt** template is configurable per workspace and per session, with variable interpolation (`{{workspace}}`, `{{os}}`, `{{datetime}}`). |
| L-07 | Support for local models via **Ollama**-compatible API as an additional provider. |

#### 2.1.5 Tool Execution

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| T-01 | Tools are functions the agent can invoke. Each tool has a name, description, JSON Schema for parameters, and an executor function. |
| T-02 | **Built-in tools:** `read_file`, `write_file`, `list_directory`, `search_files`, `run_command`, `git_diff`, `git_status`, `git_commit`. |
| T-03 | Tool calls are logged with input, output, duration, and exit code (for shell commands). |
| T-04 | Shell commands (`run_command`) execute in the workspace directory with a configurable timeout (default 120s). |
| T-05 | Tools can be restricted per workspace via an allow-list in `.cometcode/config.toml`. |
| T-06 | **Custom tools** can be registered as external executables or scripts; the daemon discovers them from `~/.cometcode/tools/`. |
| T-07 | Tool execution is **sandboxed** — file writes outside the workspace require explicit user confirmation via the client. |

#### 2.1.6 Code Analysis

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| C-01 | File reading returns content with line numbers and supports offset + limit for large files. |
| C-02 | Diff generation produces unified diffs for proposed edits, shown to the user before application. |
| C-03 | Regex and literal search across workspace files, respecting `.gitignore` and `.cometcode/ignore`. |
| C-04 | Glob-based file discovery for pattern matching (e.g., `**/*.go`). |

#### 2.1.7 API Layer

##### REST Endpoints

| Method | Path                              | Description                          |
|--------|-----------------------------------|--------------------------------------|
| GET    | `/api/v1/health`                  | Daemon health check                  |
| GET    | `/api/v1/daemon/status`           | Daemon status (uptime, clients, etc.)|
| POST   | `/api/v1/sessions`                | Create a new session                 |
| GET    | `/api/v1/sessions`                | List sessions (with filters)         |
| GET    | `/api/v1/sessions/:id`            | Get session detail                   |
| DELETE | `/api/v1/sessions/:id`            | Archive a session                    |
| POST   | `/api/v1/sessions/:id/export`     | Export session (JSON/Markdown)       |
| GET    | `/api/v1/workspaces`              | List known workspaces                |
| GET    | `/api/v1/workspaces/:id`          | Get workspace detail                 |
| GET    | `/api/v1/providers`               | List configured LLM providers        |
| GET    | `/api/v1/tools`                   | List available tools                 |
| GET    | `/api/v1/config`                  | Get daemon config                    |
| PUT    | `/api/v1/config`                  | Update daemon config                 |

##### WebSocket

| Path                              | Description                          |
|-----------------------------------|--------------------------------------|
| `/ws/v1/sessions/:id`             | Bi-directional session channel       |

**WebSocket Message Format:**

```json
{
  "type": "user_message | assistant_chunk | assistant_done | tool_call | tool_result | error | status",
  "session_id": "01HX...",
  "payload": { ... },
  "timestamp": "2026-02-17T12:00:00Z"
}
```

| Message Type       | Direction      | Payload                                             |
|--------------------|----------------|------------------------------------------------------|
| `user_message`     | Client → Server| `{ "content": "...", "attachments": [...] }`        |
| `assistant_chunk`  | Server → Client| `{ "delta": "...", "token_index": 42 }`             |
| `assistant_done`   | Server → Client| `{ "message_id": "...", "usage": {...} }`           |
| `tool_call`        | Server → Client| `{ "tool": "read_file", "args": {...}, "call_id": "..." }` |
| `tool_result`      | Server → Client| `{ "call_id": "...", "output": "...", "duration_ms": 150 }` |
| `tool_confirm`     | Server → Client| `{ "call_id": "...", "tool": "write_file", "args": {...} }` |
| `tool_approve`     | Client → Server| `{ "call_id": "...", "approved": true }`            |
| `error`            | Server → Client| `{ "code": "PROVIDER_ERROR", "message": "..." }`   |
| `status`           | Server → Client| `{ "state": "thinking | tool_executing | idle" }`  |

#### 2.1.8 CLI Commands (Cobra)

```
cometcode
├── init                          # Initialize workspace
├── daemon
│   ├── start [--port] [--host]   # Start the daemon
│   ├── stop                      # Stop the daemon
│   └── status                    # Show daemon status
├── session
│   ├── new [--model] [--title]   # Create session
│   ├── list [--all]              # List sessions
│   ├── resume <id>               # Resume session in TUI
│   ├── archive <id>              # Archive session
│   └── export <id> [--format]    # Export to JSON/MD
├── chat <message>                # One-shot message (starts daemon if needed)
├── config
│   ├── show                      # Show current config
│   ├── set <key> <value>         # Set config value
│   └── provider add <name>       # Add LLM provider
├── tool
│   ├── list                      # List available tools
│   └── run <name> [args]         # Manually run a tool
└── version                       # Print version info
```

#### 2.1.9 TUI Interface (Bubble Tea)

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| U-01 | The TUI presents a full-screen chat interface with input area, message history, and status bar. |
| U-02 | Streaming tokens render character-by-character in the assistant message bubble. |
| U-03 | Tool calls display inline with name, arguments (collapsed), and result (expandable). |
| U-04 | Markdown in assistant messages is rendered with syntax highlighting for code blocks. |
| U-05 | Keyboard shortcuts: `Ctrl+C` cancel generation, `Ctrl+D` exit, `Ctrl+N` new session, `Ctrl+L` clear screen. |
| U-06 | A status bar shows: current model, token usage, session ID, and daemon connection state. |

#### 2.1.10 Data Model

```
┌──────────────────┐       ┌──────────────────┐
│    Workspace     │       │  ModelProvider    │
│──────────────────│       │──────────────────│
│ id          TEXT │       │ id          TEXT │
│ name        TEXT │       │ name        TEXT │
│ root_path   TEXT │       │ api_key     TEXT │
│ config      JSON │       │ base_url    TEXT │
│ created_at  TIME │       │ models      JSON │
│ updated_at  TIME │       │ is_default  BOOL │
└────────┬─────────┘       └──────────────────┘
         │
         │ 1:N
         ▼
┌──────────────────┐
│     Session      │
│──────────────────│
│ id          TEXT │──────────────────────────┐
│ workspace_id TEXT│                          │
│ title       TEXT │                          │
│ model_id    TEXT │       ┌──────────────────┐
│ system_prompt TEXT│      │    ToolCall      │
│ status      TEXT │       │──────────────────│
│ token_usage JSON │       │ id          TEXT │
│ created_at  TIME │       │ message_id  TEXT │
│ updated_at  TIME │       │ tool_name   TEXT │
│ archived_at TIME │       │ arguments   JSON │
└────────┬─────────┘       │ result      TEXT │
         │                 │ duration_ms INT  │
         │ 1:N             │ exit_code   INT  │
         ▼                 │ created_at  TIME │
┌──────────────────┐       └──────────────────┘
│     Message      │              ▲
│──────────────────│              │ 1:N
│ id          TEXT │──────────────┘
│ session_id  TEXT │
│ role        TEXT │  (user | assistant | system | tool)
│ content     TEXT │
│ token_count INT  │
│ created_at  TIME │
└──────────────────┘
```

#### 2.1.11 Database Schema (SQLite via sqlc)

```sql
CREATE TABLE workspaces (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    root_path   TEXT NOT NULL UNIQUE,
    config      TEXT NOT NULL DEFAULT '{}',  -- JSON
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE sessions (
    id            TEXT PRIMARY KEY,
    workspace_id  TEXT NOT NULL REFERENCES workspaces(id),
    title         TEXT NOT NULL,
    model_id      TEXT NOT NULL,
    system_prompt TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'active'
                  CHECK (status IN ('active', 'archived')),
    token_usage   TEXT NOT NULL DEFAULT '{}',  -- JSON
    created_at    DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at    DATETIME NOT NULL DEFAULT (datetime('now')),
    archived_at   DATETIME
);

CREATE TABLE messages (
    id          TEXT PRIMARY KEY,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    role        TEXT NOT NULL
                CHECK (role IN ('user', 'assistant', 'system', 'tool')),
    content     TEXT NOT NULL,
    token_count INTEGER NOT NULL DEFAULT 0,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE tool_calls (
    id          TEXT PRIMARY KEY,
    message_id  TEXT NOT NULL REFERENCES messages(id),
    tool_name   TEXT NOT NULL,
    arguments   TEXT NOT NULL DEFAULT '{}',  -- JSON
    result      TEXT NOT NULL DEFAULT '',
    duration_ms INTEGER NOT NULL DEFAULT 0,
    exit_code   INTEGER,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE model_providers (
    id        TEXT PRIMARY KEY,
    name      TEXT NOT NULL UNIQUE,
    api_key   TEXT NOT NULL DEFAULT '',
    base_url  TEXT NOT NULL DEFAULT '',
    models    TEXT NOT NULL DEFAULT '[]',  -- JSON
    is_default INTEGER NOT NULL DEFAULT 0
);

-- Indexes
CREATE INDEX idx_sessions_workspace ON sessions(workspace_id);
CREATE INDEX idx_sessions_status    ON sessions(status);
CREATE INDEX idx_messages_session   ON messages(session_id);
CREATE INDEX idx_messages_role      ON messages(session_id, role);
CREATE INDEX idx_tool_calls_message ON tool_calls(message_id);
```

---

### 2.2 Lunar (Svelte Frontend)

#### 2.2.1 Web GUI Views

| View               | Route                  | Description                                                   |
|--------------------|------------------------|---------------------------------------------------------------|
| **Dashboard**      | `/`                    | Overview of active sessions, recent activity, daemon status.  |
| **Session List**   | `/sessions`            | Filterable/searchable list of sessions with status badges.    |
| **Chat**           | `/sessions/:id`        | Primary interaction view — message thread with streaming.     |
| **Code View**      | `/sessions/:id/code`   | Side panel or tab showing workspace files with syntax highlighting. |
| **Tool Logs**      | `/sessions/:id/tools`  | Chronological log of tool calls with expandable I/O detail.   |
| **Settings**       | `/settings`            | Daemon connection, LLM provider config, appearance, keybindings. |
| **Workspace Browser** | `/workspaces`       | List of workspaces with quick-launch into a session.          |

#### 2.2.2 Desktop Application (Electron)

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| E-01 | Lunar ships as an Electron app wrapping the SvelteKit static build.      |
| E-02 | The Electron main process can optionally **auto-start** the CometCode daemon if not already running. |
| E-03 | A **system tray** icon shows daemon status (connected / disconnected / error) and provides quick actions. |
| E-04 | Deep links (`cometline://session/<id>`) open the app to a specific session. |
| E-05 | Native OS notifications for: session events, tool confirmations, errors. |
| E-06 | IPC bridge between Electron main and Svelte renderer for OS-level APIs (file dialogs, clipboard, notifications). |

#### 2.2.3 API Communication

| ID   | Requirement                                                              |
|------|--------------------------------------------------------------------------|
| A-01 | An **HTTP client** module wraps all REST calls with typed request/response interfaces. |
| A-02 | A **WebSocket client** manages connection lifecycle (connect, reconnect, heartbeat) and dispatches incoming messages to Svelte stores. |
| A-03 | **Daemon discovery**: on startup, Lunar checks `~/.cometcode/daemon.pid` for host:port, or falls back to `127.0.0.1:7700`. |
| A-04 | Connection state is reactive and surfaced in the UI (connected / reconnecting / disconnected). |
| A-05 | All API errors are caught, categorized, and displayed as toast notifications. |

#### 2.2.4 UI Components

| Component          | Purpose                                                                  |
|--------------------|--------------------------------------------------------------------------|
| `ChatMessage`      | Renders a single message (user or assistant) with Markdown and code blocks. |
| `ChatInput`        | Multi-line input with send button, attachment support, and keyboard shortcuts. |
| `CodeBlock`        | Syntax-highlighted code block with copy button and language label.        |
| `DiffViewer`       | Side-by-side or inline unified diff display for file changes.            |
| `FileTree`         | Collapsible file/directory tree for workspace browsing.                  |
| `ToolCallCard`     | Compact card showing tool name, args summary, result preview, duration.  |
| `SessionCard`      | Session preview card with title, model, message count, last active time. |
| `StatusBar`        | Persistent bar showing daemon status, model, token usage.                |
| `Toast`            | Non-blocking notification component for errors, confirmations, info.     |
| `Modal`            | Generic modal container for settings panels and confirmations.           |
| `ProviderConfig`   | Form for adding/editing LLM provider credentials and model selection.    |
| `StreamingText`    | Component that renders tokens as they arrive with a blinking cursor.     |

#### 2.2.5 State Management

Lunar uses **Svelte 5 runes** (`$state`, `$derived`, `$effect`) for component-local state and a set of **shared stores** for cross-component state:

| Store              | Contents                                                                 |
|--------------------|--------------------------------------------------------------------------|
| `daemon`           | Connection status, host, port, health check interval.                   |
| `sessions`         | Session list, active session ID, loading state.                         |
| `chat`             | Messages for the active session, streaming buffer, scroll position.     |
| `tools`            | Tool call log for the active session, expanded/collapsed state.         |
| `workspace`        | Active workspace metadata, file tree cache.                             |
| `settings`         | User preferences (theme, keybindings, default model).                   |
| `notifications`    | Toast queue.                                                            |

---

## 3. Non-Functional Requirements

### 3.1 Performance

| ID    | Requirement                                                             |
|-------|-------------------------------------------------------------------------|
| P-01  | REST API responses (non-streaming) return within **200ms** at p95.      |
| P-02  | Time-to-first-token for streaming responses is bounded by the upstream LLM provider; the daemon adds no more than **50ms** of overhead. |
| P-03  | The daemon idle memory footprint stays below **100 MB**.                |
| P-04  | The daemon supports at least **10 concurrent WebSocket connections** without degradation. |
| P-05  | Lunar initial page load (cached) completes in under **1 second**.      |
| P-06  | SQLite queries for session listing (1000 sessions) complete in under **50ms**. |
| P-07  | File search across a 10,000-file workspace completes in under **2 seconds**. |

### 3.2 Security

| ID    | Requirement                                                             |
|-------|-------------------------------------------------------------------------|
| SE-01 | The daemon binds to `127.0.0.1` by default; binding to `0.0.0.0` requires explicit `--host` flag and prints a warning. |
| SE-02 | API keys for LLM providers are stored in `~/.cometcode/providers.toml` with file permissions `0600`. |
| SE-03 | Tool execution is sandboxed to the workspace directory. File writes outside the workspace require explicit client-side approval. |
| SE-04 | Shell command execution (`run_command` tool) logs the full command and prevents execution of commands matching a configurable deny-list. |
| SE-05 | WebSocket connections are authenticated via a session token generated on daemon start and stored in the PID file. |
| SE-06 | Lunar's Electron build enables `contextIsolation` and disables `nodeIntegration` in the renderer process. |
| SE-07 | User input in tool arguments is sanitized to prevent command injection. |
| SE-08 | No telemetry or analytics data is collected or transmitted. |

### 3.3 Reliability & Fault Tolerance

| ID    | Requirement                                                             |
|-------|-------------------------------------------------------------------------|
| R-01  | If the LLM provider returns a transient error (429, 500, 503), the daemon retries up to 3 times with exponential backoff. |
| R-02  | If a WebSocket connection drops, the client automatically reconnects with the full session state preserved server-side. |
| R-03  | Session state is persisted to SQLite after every message; a daemon crash loses at most the in-flight streaming response. |
| R-04  | Tool execution timeouts (default 120s) prevent runaway processes from blocking the daemon. |
| R-05  | The daemon handles `SIGTERM` and `SIGINT` gracefully: flushes pending writes, closes connections, removes PID file. |
| R-06  | If the SQLite database is corrupted, the daemon logs the error and attempts to recover from the WAL, or creates a new database with a backup of the old file. |

### 3.4 Extensibility

| ID    | Requirement                                                             |
|-------|-------------------------------------------------------------------------|
| X-01  | New LLM providers can be added by implementing a Go interface (`Provider`) with `Chat`, `StreamChat`, and `ListModels` methods. |
| X-02  | Custom tools are executable files in `~/.cometcode/tools/` that follow a JSON-in/JSON-out protocol via stdin/stdout. |
| X-03  | The system prompt is a Go template that can be customized per workspace or per session. |
| X-04  | Lunar's component library is designed for composition — new views can be added by creating a route and composing existing components. |

### 3.5 Offline Capability

| ID    | Requirement                                                             |
|-------|-------------------------------------------------------------------------|
| O-01  | The daemon, CLI, and TUI are fully functional without internet access when using a local LLM provider (Ollama). |
| O-02  | Session history, workspace state, and configuration are always available offline. |
| O-03  | When the configured LLM provider is unreachable, the daemon returns a clear error and the UI shows a "provider unavailable" state rather than hanging. |

### 3.6 Portability

| ID    | Requirement                                                             |
|-------|-------------------------------------------------------------------------|
| PT-01 | CometCode compiles to a single static binary for **Linux** (amd64, arm64), **macOS** (amd64, arm64), and **Windows** (amd64). |
| PT-02 | Lunar's Electron build targets macOS (universal), Windows (x64), and Linux (x64, AppImage + deb). |
| PT-03 | Lunar's web build is a static SvelteKit site that can be served by any HTTP server or opened as a local file. |
| PT-04 | All filesystem paths use platform-appropriate separators and respect `XDG_CONFIG_HOME` on Linux, `~/Library/Application Support` on macOS, and `%APPDATA%` on Windows. |

### 3.7 Developer Experience

| ID    | Requirement                                                             |
|-------|-------------------------------------------------------------------------|
| DX-01 | The daemon exposes a `/api/v1/health` endpoint returning `200 OK` for liveness probes and integration tests. |
| DX-02 | Structured JSON logging via `zap` with configurable log levels (`debug`, `info`, `warn`, `error`). |
| DX-03 | `cometcode daemon start --dev` enables verbose logging and CORS `*` for local Lunar development. |
| DX-04 | Lunar runs `vite dev` with hot module replacement against a running daemon — no build step during development. |
| DX-05 | Both projects include CI pipelines (GitHub Actions) for lint, test, and build on every push. |
| DX-06 | Go tests use `testing` + `testify`; Svelte tests use `vitest` + `@testing-library/svelte`. |

---

## 4. Tech Stack

### 4.1 CometCode (Go Runtime)

| Component         | Choice          | Rationale                                                              |
|-------------------|-----------------|------------------------------------------------------------------------|
| Language          | **Go 1.23+**    | Fast compilation, single binary output, excellent concurrency, strong standard library. |
| HTTP Framework    | **Gin**         | High-performance, battle-tested router with middleware ecosystem.      |
| WebSocket         | **gorilla/websocket** | Mature, well-documented WebSocket implementation for Go.         |
| CLI Framework     | **Cobra**       | De facto standard for Go CLIs; subcommand structure, flag parsing, shell completions. |
| TUI Framework     | **Bubble Tea**  | Elm-architecture TUI framework with composable components (Bubbles).  |
| TUI Styling       | **Lip Gloss**   | Declarative terminal styling companion to Bubble Tea.                 |
| Database          | **SQLite**      | Zero-config embedded database; single file, ACID-compliant, perfect for local-first. |
| SQL Codegen       | **sqlc**        | Generates type-safe Go code from SQL queries; no ORM overhead.        |
| Configuration     | **Viper**       | Reads TOML/YAML/env config with layered precedence.                   |
| Logging           | **zap**         | High-performance structured logger with zero-allocation fast path.    |
| ID Generation     | **oklog/ulid**  | Sortable, URL-safe unique IDs; better than UUIDv4 for time-ordered data. |
| Testing           | **testify**     | Assertions and mocking for Go tests.                                  |
| Build             | **GoReleaser**  | Cross-compilation, checksums, changelogs, and release automation.     |

### 4.2 Lunar (Svelte Frontend)

| Component         | Choice             | Rationale                                                           |
|-------------------|--------------------|---------------------------------------------------------------------|
| Framework         | **SvelteKit**      | Full-stack Svelte framework with file-based routing, SSR/SSG, and adapter ecosystem. |
| Language          | **TypeScript**     | Type safety across the frontend; catches API contract mismatches at compile time. |
| UI Reactivity     | **Svelte 5 (Runes)** | `$state`, `$derived`, `$effect` — fine-grained reactivity without virtual DOM overhead. |
| Styling           | **Tailwind CSS v4** | Utility-first CSS with JIT compilation; consistent design system without custom CSS. |
| Desktop Wrapper   | **Electron**       | Proven cross-platform desktop framework; reuses web codebase.       |
| Electron Tooling  | **electron-vite**  | Unified Vite config for main, preload, and renderer processes.      |
| Build Tool        | **Vite**           | Fast HMR, ESM-native bundling, plugin ecosystem.                   |
| Testing           | **Vitest**         | Vite-native test runner; fast, ESM-first, compatible with Jest API. |
| Component Testing | **@testing-library/svelte** | DOM-testing utilities for Svelte components.              |
| Code Quality      | **ESLint + Prettier** | Consistent code style and error prevention.                      |
| Package Manager   | **pnpm**           | Fast, disk-efficient, strict dependency resolution.                 |
| Distribution      | **electron-builder** | App packaging, code signing, auto-update, and notarization.      |

### 4.3 Communication

| Component           | Choice            | Rationale                                                          |
|---------------------|-------------------|--------------------------------------------------------------------|
| API Protocol        | **REST (JSON)**   | Simple, debuggable, universally supported for CRUD operations.     |
| Real-time Protocol  | **WebSocket**     | Bi-directional streaming for chat messages and tool call events.   |
| Serialization       | **JSON**          | Human-readable, native to both Go and TypeScript, debugger-friendly. |
| Future Consideration| **gRPC**          | Potential addition for high-throughput internal communication or CLI-to-daemon calls. |

### 4.4 Deployment & Distribution

| Component           | Choice                 | Rationale                                                       |
|---------------------|------------------------|-----------------------------------------------------------------|
| CometCode Binary    | **GoReleaser**         | Single static binary per platform; no runtime dependencies.     |
| Lunar Web           | **SvelteKit static adapter** | Pre-rendered static site; deployable anywhere or bundled with daemon. |
| Lunar Desktop       | **electron-builder**   | DMG (macOS), NSIS (Windows), AppImage + deb (Linux).           |
| CI/CD               | **GitHub Actions**     | Lint, test, build, and release pipelines for both projects.     |
| Container (optional)| **Docker**             | Single container running daemon + static web UI for self-hosting. |

### 4.5 Development Tools

| Tool                | Purpose                                                                  |
|---------------------|--------------------------------------------------------------------------|
| **Task (taskfile.dev)** | Polyglot task runner for build, test, lint commands across both projects. |
| **Air**             | Go live-reload during development.                                       |
| **golangci-lint**   | Comprehensive Go linter aggregator.                                      |
| **svelte-check**    | Svelte-specific type checking and diagnostics.                           |

---

*End of document.*
