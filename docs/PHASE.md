# Cometline Phase Plan

This is the living task board and roadmap for Cometline frontend shell work. Keep this file plus `COMETLINE_ARCHITECTURE.md` as the only long-lived docs in `cometline/docs/`.

## Current Snapshot

- **Phase 0: Desktop scaffold**: done.
- **Phase 1: Chat/session MVP**: mostly done.
- **Phase 2: Permission UX**: next major backend+frontend phase.
- **Phase 3: Provider/secrets hardening**: partially prototyped in Electron, not production-safe.
- **Phase 4+: Memory, artifacts, scheduler, subagents, browser/computer use**: not started. Skills have a read-only first slice.

## Phase Task Board

When a task ships, move it to `Done`, update the relevant phase section below, and add any newly discovered follow-up to `Next` or `Backlog`.

### Current

- [ ] Add visible stop/abort button during active streaming. Phase: 1. Owner area: Cometline UI + existing CometMind abort API.

### Next

- [ ] Improve session title generation and sidebar update timing. Phase: 1. Owner area: CometMind title logic + Cometline session refresh.
- [ ] Add retry/reconnect UX for dropped SSE streams. Phase: 1. Owner area: Cometline chat store and chat UI.
- [ ] Add focused tests for first-message queue, route remounting, delete confirmation, and chat error states. Phase: 1. Owner area: Cometline test suite.
- [ ] Move provider model discovery behind a CometMind provider/config API. Phase: 3. Owner area: CometMind API + Cometline settings.
- [ ] Move API keys out of `~/.cometmind/cometline-settings.json`. Phase: 3. Owner area: CometMind secrets + Electron IPC.

### Backlog

- [ ] Add CometMind permission events for impactful tool calls. Phase: 2. Owner area: CometMind agent/server/session persistence.
- [ ] Build Cometline approval UI for pending tool calls. Phase: 2. Owner area: Cometline chat/tool cards.
- [ ] Add provider test flow with success, invalid key, rate limit, and network error states. Phase: 3. Owner area: CometMind provider API + Cometline settings.
- [ ] Add memory schema/retrieval and memory management UI. Phase: 4. Owner area: CometMind memory + Cometline settings/chat indicators.
- [ ] Add artifact cards and sandboxed previews. Phase: 6. Owner area: Cometline UI + Electron-safe preview/open actions.

### Done

- [x] Scaffold SvelteKit + Electron desktop shell. Phase: 0.
- [x] Add root `Makefile` for install/dev/check/build helpers. Phase: 0.
- [x] Start and health-check CometMind from Electron. Phase: 0.
- [x] Implement session list, creation, transcript loading, and SSE streaming. Phase: 1.
- [x] Add `/` centered hero composer and `/session/[id]` active chat route. Phase: 1.
- [x] Add first-message queue and flight transition into the chat thread. Phase: 1.
- [x] Render assistant text, reasoning, tool calls/results, status, and normalized errors. Phase: 1.
- [x] Add session deletion, hover trash control, and in-app confirmation sheet. Phase: 1.
- [x] Add provider settings modal, model fetch prototype, and composer model selector. Phase: 3 prototype.
- [x] Add read-only skills settings, current skill list, and sync action. Phase: 5 first slice.
- [x] Generate rounded desktop app icon and use `project_icon.png` for assistant/project avatar. Phase: 0/1 polish.

## Phase 0: Desktop Scaffold

Status: done.

Implemented:

- SvelteKit + Electron app under `cometline/`.
- Electron main/preload with `contextIsolation: true` and `nodeIntegration: false`.
- CometMind process manager and health polling.
- Static SvelteKit build via `@sveltejs/adapter-static`.
- Root `Makefile` targets for install, dev, check, build, port, and log cleanup.
- Rounded app icon assets for development and packaged macOS app metadata.

Exit criteria: met.

## Phase 1: Chat And Session MVP

Status: mostly done.

Implemented:

- Sidebar session list scoped to workspace path.
- New session creation with selected provider/model.
- `/` home route with centered hero composer.
- `/session/[id]` chat route keyed by session id.
- First-message queue from home route into session route.
- First-message flight transition before real user bubble insertion.
- Transcript loading.
- SSE message streaming.
- User, assistant, reasoning, tool, status, and error rendering.
- Throttled streaming auto-scroll.
- Session deletion with backend `DELETE /api/v1/sessions/{id}`.
- In-app delete confirmation with `Don't ask again` persisted in localStorage.
- Clean missing-API-key error presentation.

Remaining before Phase 1 is fully closed:

- Add visible stop/abort button during streaming.
- Improve session title generation and update timing.
- Add retry/reconnect UX for dropped SSE.
- Add focused tests for first-message queue, route remounting, and delete confirmation.

## Phase 2: Permission UX

Status: not started.

Goal: make local tool execution understandable and controllable.

Backend work needed first:

- Tool risk metadata.
- Permission request SSE event.
- Approve/deny endpoint.
- Run-loop pause/resume while waiting for approval.
- Persisted tool-call permission state.

Frontend work:

- Render pending tool approval cards.
- Show exact command/path/input requested.
- Approve once, approve for session, deny.
- Render denied and resumed states.
- Add tests around pending permission state.

Exit criteria:

- User can approve or deny impactful tool calls without leaving chat.
- Cometline never executes or simulates approval locally; every decision goes through CometMind.

## Phase 3: Provider, Model, And Secrets Hardening

Status: MVP prototype exists, production version not started.

Implemented prototype:

- `Command+,` settings modal.
- Provider/base URL/API key fields.
- Electron IPC model fetch from `<baseURL>/models`.
- Save settings and restart CometMind with `COMETMIND_*` env vars.
- Composer model selector populated from fetched models.

Production work still needed:

- Move secrets out of `~/.cometmind/cometline-settings.json`.
- Put provider config and model discovery behind CometMind API.
- Use OS keychain or another encrypted local secret store.
- Add provider test flow.
- Redact secrets in logs, UI errors, screenshots, and diagnostics.

Exit criteria:

- User can configure/test providers from the app without raw secrets living in renderer-accessible state or plaintext Electron JSON.

## Phase 4: Memory UI

Status: not started.

Goal: make memory visible and editable before automatic learning becomes prominent.

Needed:

- Backend memory schema and retrieval.
- Memory list/search/edit/delete UI.
- Memory-injected chips in chat.
- Proposed memory write review UI.

## Phase 5: Skills UI

Status: not started.

Goal: expose procedural memory as manageable user assets.

Needed:

- Skills index and detail views.
- Enable/disable flows.
- Slash command or invocation autocomplete.
- Diff/review UI for skill changes.

## Phase 6: Artifacts And Rich Output

Status: not started.

Goal: render generated outputs beyond plain chat text.

Needed:

- Artifact cards for files, markdown, HTML, SVG, Mermaid, images, and logs.
- Sandboxed HTML preview.
- Open/download actions through Electron IPC.

## Phase 7: Scheduler UI

Status: not started.

Goal: let users inspect and control background automations safely.

Needed:

- Jobs list.
- Job creation/edit form.
- Run history.
- Pause/resume/run-now/delete actions.

## Phase 8: Subagents And Coding Delegation UI

Status: not started.

Goal: make delegated workstreams understandable.

Needed:

- Child task cards.
- Nested transcript/progress details.
- Cancellation controls.

## Phase 9: Browser And Computer-Use UX

Status: not started.

Goal: make browser/computer automation visible and permissioned.

Needed:

- Browser session cards.
- Requested action preview.
- Domain allow/block settings.
- Evidence trail.

## Phase 10: Cloud Readiness

Status: not started.

Goal: keep the renderer reusable outside Electron.

Needed:

- Centralized API base URL/mode detection.
- Desktop-only capability checks.
- Hosted auth boundary.
- Responsive web layout review.

## Recommended Next Order

1. Add stop/abort button in the current chat UI.
2. Add CometMind permission events and Cometline approval UI.
3. Move provider settings/secrets into a safer CometMind-owned path.
4. Improve title generation and session grouping.
5. Add tests around chat transitions and route/session state.
