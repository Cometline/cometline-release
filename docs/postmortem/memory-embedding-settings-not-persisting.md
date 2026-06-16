# Memory embedding model resets after save

**Date:** 2026-06-16  
**Components:** `cometline/src/lib/components/SettingsMemoryPanel.svelte`, `cometline/src/lib/components/SettingsPanel.svelte`, `cometline/src/lib/embedding-models.ts`, `cometline/electron/main.cjs`, `cometmind/server/server.go`, `cometmind/server/memory_handlers.go`

## Symptom

In **Settings → Memory**, the user selected an embedding model (e.g.
`OpenAI Compatible · text-embedding-3-small`), clicked save, and reopened the
panel — but the dropdown showed **Select embedding model…** again.

Memory **write/search could still work** (manual add, auto-retrieve, compaction),
which made the bug confusing: runtime memory was alive, but the UI looked
unconfigured.

In dev (`make dev`, renderer on `http://127.0.0.1:5173`), saving sometimes
surfaced **Failed to fetch** in the settings footer. The browser console showed:

```text
Access to fetch at 'http://127.0.0.1:7700/api/v1/memory/settings' from origin
'http://127.0.0.1:5173' has been blocked by CORS policy: Method PUT is not
allowed by Access-Control-Allow-Methods in preflight response.
```

## Root cause

Three separate issues stacked:

### 1. UI reconciliation was too strict

The dropdown bound to `selectedEmbeddingKey`, which had to exactly match
`${providerId}:${model}` in **currently enabled** provider models
(`listEmbeddingModelOptions`). Saved values from the CometMind API or local
settings that were missing `provider_id`, not in `enabledModels`, or only present
in one store were dropped — HTML `<select>` then fell back to the placeholder
option.

Memory could still run because CometMind `resolveMemoryEmbedding()` falls back
to provider config when `memory.embedding` is empty in `config.toml`. The UI
only read explicit saved embedding fields.

### 2. Split persistence and split Save buttons

Saving embedding touched three places:

| Store | How |
| ----- | --- |
| CometMind in-memory | `PUT /api/v1/memory/settings` |
| `~/.cometmind/cometline-settings.json` | `persistMemoryEmbedding()` |
| `~/.cometmind/config.toml` | `writeCometMindConfig()` |

On reload the panel called **GET** from CometMind only. The footer **Save**
button did not call `PUT /memory/settings`; only a separate Memory-panel
**Save settings** button did. Users who saved via the footer thought embedding
was persisted locally, but runtime/API state and dropdown reconciliation could
still be empty.

### 3. CORS blocked PUT in dev

CometMind `localCORS()` allowed `GET, POST, PATCH, DELETE, OPTIONS` but not
**PUT**. Memory settings use `PUT /api/v1/memory/settings`. Packaged Cometline
(`app://`) often same-origin with the sidecar; dev Vite on port **5173** is
cross-origin and hit preflight failure after unifying save to the footer button.

## Fix process

### 1. Harden embedding resolution (`embedding-models.ts`)

- `mergeEmbeddingFields()` — merge API + local saved embedding
- `resolveEmbeddingSelection()` — fall back to model-only match; synthesize
  **orphan** option when saved model is not in enabled providers
- `buildEmbeddingDropdownOptions()` — include orphan entry labeled
  `(enable in Providers)`
- Pass `savedEmbedding` from `draft.cometmind.memory.embedding` into
  `SettingsMemoryPanel`

### 2. Unify Save UX

- Removed Memory-panel **Save settings** button
- Footer **Save** on the Memory section calls `memoryPanel.saveMemorySettings()`
  (PUT + local persist via `persistMemoryEmbedding`)
- Footer hint: *Embedding and memory behavior save with Save below*

### 3. Keep providers in sync on save

`persistMemoryEmbedding()` adds the chosen embedding model to the provider's
`enabledModels` before writing local settings, so future reloads find a match.

### 4. Allow PUT in CometMind CORS

```go
// server.go localCORS()
c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
```

Added `TestLocalCORSAllowsMemorySettingsPut` in `server/server_test.go`.

## How to avoid regressions

- Any new CometMind endpoint used from the **Vite dev renderer** must be allowed
  in `localCORS()` — especially non-GET methods. If save works packaged but
  fails in dev with CORS, check `Access-Control-Allow-Methods` first.
- Memory UI must not assume `GET /memory/settings` is the only source of truth;
  reconcile with `cometmind.memory.embedding` in local settings and show orphan
  options when enabled providers drift.
- Do not reintroduce a second save path for Memory-only fields without wiring
  the footer **Save** to the same `putMemorySettings` + `persistMemoryEmbedding`
  flow.
- **Memory working ≠ embedding saved in UI** — provider fallback can embed
  without explicit `[memory.embedding]` in config.

## Verification

1. Rebuild/restart CometMind after CORS change (`go build` / Save Providers in
   Settings to restart sidecar).
2. Dev: `make dev` — open Settings → Memory, select embedding, footer **Save**
   — no CORS error in console; message **Memory settings saved.**
3. Close and reopen Settings → Memory — dropdown shows saved model.
4. Quit and relaunch Cometline — selection persists.
5. `cd cometmind && go test -run TestLocalCORS ./server`
6. `cd cometline && pnpm test embedding-models`

## Relation to other postmortems

- [memory-extract-wrong-provider.md](./memory-extract-wrong-provider.md) —
  runtime memory LLM calls; separate from embedding **settings** persistence.
- [fetch-models-data-clone-error.md](./fetch-models-data-clone-error.md) —
  Settings → Providers IPC; embedding models must be enabled there before they
  appear in the Memory dropdown.
