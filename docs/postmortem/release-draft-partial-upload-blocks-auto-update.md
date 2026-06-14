# Release stuck as draft with partial assets blocks auto-update

**Date:** 2026-06-14
**Components:** `.github/workflows/release.yml` (parent repo `cometline-release`), `electron/main.cjs`, `electron/preload.cjs`

## Symptom

The auto-updater never found new versions. On GitHub, the releases for
`v0.0.1` and `v0.0.2` were stuck as **drafts** (`v0.0.2` had two duplicate
draft entries). The `v0.0.2` draft contained only a single tiny asset
(`Cometline-0.0.2-arm64.dmg.blockmap`, ~153 KB) — the `.dmg`, `.zip`, their
other blockmaps, and `latest-mac.yml` were all missing.

## Root cause

The release job ran `electron-builder --mac --publish always`, which uploads
to GitHub in a fire-and-forget manner and creates the release as a **draft**.

1. **Drafts are invisible to electron-updater.** The client reads
   `latest-mac.yml` from the _published_ release for a tag; a draft release is
   never served, so updates were never detected.
2. **The large uploads did not complete.** CI logs (run 27491065573, marked
   "ok") showed electron-builder start uploading the dmg/zip at 06:51:35, but
   the next workflow step was already running by 06:51:53. A 147 MB dmg and
   144 MB zip cannot upload in ~18 s — the process exited before the uploads
   finished, leaving a partial draft with only the small blockmap.
3. **Stale-draft cleanup missed them.** The cleanup step filtered releases by
   `.tag_name == TAG`, but electron-builder drafts have no tag attached yet
   (they appear as `untagged-*`), so orphaned partial drafts accumulated.

## Fix

Stop relying on electron-builder's `--publish always`. Make publishing an
explicit, synchronous step:

- Build with **`build:mac`** (`electron-builder --mac --publish never`), which
  still produces every artifact plus `latest-mac.yml` in `cometline/dist/`.
- Verify the dmg/zip exist and are non-empty before publishing.
- Publish with **`gh release create "$TAG" --target "$GITHUB_SHA"
--generate-notes <assets…>`**, which creates a **non-draft** release and
  waits for each asset upload to finish.
- Clean up both the tagged release for this tag **and** any `untagged-*`
  drafts whose assets match this version before building.
- Add a post-publish verification step that asserts `latest-mac.yml`, a
  `.dmg`, and a `-mac.zip` are attached to the release.

Renderer/updater wiring added alongside the CI fix so users can act on
updates:

- `electron/main.cjs`: emit `cometline:update-state`
  (`checking`/`downloading`/`ready`/`error` with version + percent) to the
  renderer, plus IPC handlers `get-update-state`, `check-for-updates`, and
  `install-update` (calls `autoUpdater.quitAndInstall()`).
- `electron/preload.cjs`: expose `getUpdateState`, `checkForUpdates`,
  `installUpdate`, `onUpdateState`.
- `UpdateButton.svelte`: bottom-left dock showing download progress and a
  "Restart to update" button when an update is ready.

## How to avoid regressions

- For electron-updater on GitHub Releases, the release **must be published
  (non-draft)** and contain `latest-mac.yml`, the `.zip`, and the
  `.blockmap`. Treat draft releases as broken for auto-update.
- Prefer an explicit `gh release create/upload` (which blocks until uploads
  complete) over tools that upload asynchronously and may exit early.
- When matching releases for cleanup, remember electron-builder drafts are
  `untagged-*` until published — match on assets/version, not just `tag_name`.
- The "Verify release assets uploaded" step fails the run if required assets
  are missing, so a partial release can no longer pass silently.
