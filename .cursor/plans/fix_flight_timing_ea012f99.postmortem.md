# First-Turn Flight Postmortem

**Status:** Solved
**Originally proposed in:** `.cursor/plans/fix_flight_timing_ea012f99.plan.md`
**Affected files:**

- `cometline/src/lib/conversation/conversation-controller.ts`
- `cometline/src/lib/conversation/conversation-controller.test.ts`
- `cometline/src/lib/components/ChatView.svelte`
- `cometline/src/lib/components/chat/ChatThread.svelte`
- `cometline/src/lib/components/chat/ThreadAvatar.svelte`
- `cometline/src/lib/components/FirstTurnFlight.svelte`

## Goal

Empty-state first turn should look like this, in order:

1. Overlay avatar animates from the empty-state center to the assistant row.
2. The real destination avatar and "Thinking..." indicator only appear **after** the overlay arrives — no double avatar, no premature spinner.
3. Follow-up turns: user bubble flies in, assistant pending row + indicator appear immediately (already correct).

The first-turn flight must also be **resilient to mid-flight state churn**: staging the user message and appending the pending assistant row must not un-hide the destination.

## What the original plan got right

- Active first turn should let the DOM flight own staging/revealing, instead of pre-staging in the controller.
- Active follow-up should start the backend send in parallel with the bubble flight, not after awaiting the whole 560 ms flight.
- Background turns should stage+reveal in the controller and skip DOM flight entirely.
- Adapter context must carry session-stable `stageUser` / `revealStagedUser` callbacks bound to the enqueued `turnSessionId`, not the live `getSessionId()`, so a session switch during flight does not retarget the reveal.

## What the original plan got wrong (and what we changed)

1. **Adapter context shape.** It described `{ firstTurn, visualOnly, ... }` but did not thread `sessionId` or the staging callbacks. Without `sessionId`, the adapter could not bind per-run reveal targets; without per-run `stageUser`/`revealStagedUser`, the controller could not safely stage before send and reveal after flight on the correct session. **Fix:** `ConversationFlightAdapter.onUserMessageFlight` ctx is now `{ firstTurn, sessionId, stageUser, revealStagedUser }`. `FirstTurnFlight.run/runAsync` accepts these as per-run `RunOptions`; the Svelte component keeps the no-op fallbacks so it still works when called from a different mount.

2. **Controller sequencing.** The original `stage → reveal → flight → send` order for follow-up was broken: `revealStagedUser` removes `data-flight-target="user"` before `flyUserBubble` can measure the follow-up target, so the flight silently fails. **Fix:** stage hidden, start the flight promise, immediately call `send(..., { skipUser: true })`, then `await` the send and the flight promise, with the reveal bound to the flight's `finally` (so it always fires even if the flight rejects). For active first turn the controller still awaits the whole flight but the flight itself owns staging and reveal.

3. **The "missing piece" — parallel-send for the active first turn.** Active first turn previously awaited the full flight (560 ms) before `send()` was called, so the assistant pending row and indicator did not appear until after flight completion. **Fix:** introduce `sendPromise` / `startSend()` inside `runTurn` (start exactly once) and make the per-run `stageUser` callback in the active first-turn branch also call `void startSend()`. This stages the hidden user, then immediately appends the pending assistant row, so the destination row exists for `FirstTurnFlight` to measure and the indicator/avatar handoff is well-defined. `runTurn` then does `await startSend(); if (flightPromise) await flightPromise;` for the rest of the path.

4. **isViewing for fresh sessions.** The original gate `deps.getSessionId() === turnSessionId && chatStore.sessionID === turnSessionId` could classify a fresh new session as "background" (route updated first, `chatStore.bindSession()` lagged), which would skip the DOM `FirstTurnFlight` entirely. **Fix:** relax to `deps.getSessionId() === turnSessionId` and have `firstTurn` use the cached item count only for the background-flight staging path (`usesFlight && !isViewing`). This keeps the active-flight branch route-driven.

5. **The actual root cause of the remaining "double avatar" bug.** Even with an explicit `firstTurnHandoffPending` state in `ChatView.svelte` (set true at flight start, set false on `onFlightDoneChange(true)`), the destination avatar and "Thinking..." indicator still appeared mid-flight. The reason: the per-session reset `$effect` in `ChatView.svelte` (intended to fire only on `sessionId` change) also read `chatStore.isAwaitingFirstAssistant(sessionId)` and `chatStore.getCachedItemCount(sessionId)`. Both are reactive. When the flight ran, staging the user and appending the pending assistant row mutated the store, which **re-triggered this effect mid-flight**. The re-run executed `firstTurnHandoffPending = false` and recomputed `firstTurnFlightDone = true`, un-hiding the destination row before the overlay arrived. This also produced the `[svelte] derived_inert` warning (store-derived state read inside an effect whose lifetime is being repeatedly torn down and recreated).

   **Fix:** keep `void sessionId;` as the effect's only tracked dependency and wrap the entire reset body in `untrack(() => { ... })`. Now the reset only fires on an actual session change. This is the change that finally made the destination stay hidden through the full overlay flight.

6. **Robust destination hiding.** Three render paths in `ChatThread.svelte` can render the first assistant row:

   - the `awaitingFirstAssistant && !firstUserId` placeholder row
   - the `showFirstTurnAvatarSlot() && item.id === firstUserId` slot rendered right after the first user row
   - the normal-list `{#each}` path (`item.type === 'assistant' && showAssistantRow(item) && firstAssistantInNormalList(item)`)

   All three now gate on `firstTurnHandoffPending` — the avatar via `flightHidden` on `ThreadAvatar` (which also sets inline `style:visibility={flightHidden ? 'hidden' : undefined}` to defeat style-scoping / specificity issues), and the assistant stack / spinner via a `.first-turn-destination-hidden` class (`visibility: hidden`). A new `firstAssistantRowId` `$derived` (the first assistant row's id, including pending-only rows) is used to match the normal-list path, since the previous `firstAssistantId` only matched assistants with text/reasoning.

7. **Avatar handoff timing inside `FirstTurnFlight`.** The overlay avatar's CSS animation is now the source of truth for handoff: `waitForAvatarFlightEnd()` listens for the overlay `<div class="flight-particle avatar-flight">` `animationend` (with a `FLIGHT_MS + 160` ms fallback), and `waitForStableRect()` samples the target rect across two consecutive stable frames to avoid measuring a stale target while thread scroll/layout settles. Both the normal path and the `!userFlew` fallback await `avatarFlightEnd` before revealing the user and setting `firstTurnFlightDone = true`, and an extra `afterPaint()` is awaited before tearing down the flight particle so the real avatar slot is un-hidden first and the avatar never blinks out between overlay end and the thread slot.

## Tests added/updated

`cometline/src/lib/conversation/conversation-controller.test.ts`:

- `lets active first-turn flight own staging before sending with skipUser` — assert controller does not stage/reveal; flight ctx includes `firstTurn: true, sessionId: 'sess-1'`; send called with `skipUser: true`.
- `starts active first-turn send as soon as the flight stages the user` — flight stages then waits on a gate before revealing; assert `send` is already called before reveal, then release and assert reveal targets `sess-1`.
- `stages and reveals a queued background turn without running DOM flight` — second queued turn in a switched-away session is staged/revealed without invoking DOM flight again; send with `skipUser: true`.
- `uses the enqueued session callbacks if getSessionId changes during active first-turn flight` — adapter is invoked with the enqueued `turnSessionId` even after `getSessionId` switches; stage and reveal are bound to the original session.
- `runs active first-turn flight even before the chat store binds the session` — fresh-session route with `chatStore.bindSession()` lagging still uses DOM flight.
- Follow-up test updated to assert parallel-send ordering: `send` is called before the flight promise resolves, then reveal runs in the flight's `finally`.

## Verification

- `rtk pnpm run check` → 0 errors, 0 warnings.
- `rtk pnpm run test` → 30 files, 264 tests, all passing.
- Manual: empty-state first turn renders the overlay avatar sliding from the empty-state center to the assistant row with no double avatar, no premature "Thinking..." indicator; follow-up turns render the user bubble flight with the assistant row visible from the start.
