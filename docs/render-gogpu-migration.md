# Render Gogpu Migration

Goal: remove the old Ebiten-shaped `render.Image` screen/render-target abstraction and make `render/` model the gogpu pipeline directly.

The final design should have these responsibilities separated:

- `Image`: CPU-backed RGBA texture data only.
- `Frame`: per-frame command buffer and clear/camera state.
- `Texture`: GPU resource/cache entry owned by the renderer.
- `WorldMesh`: retained world geometry buffer source.
- CPU raster code: only for offscreen texture generation and tests.

## Rules

- Keep each phase compiling and passing `go test ./...`.
- Keep behavior unchanged unless a phase explicitly says otherwise.
- Do not mix game, UI, or input policy into `render/`.
- Prefer small mechanical commits over one large rewrite.
- Remove compatibility helpers once their callers are migrated.

## Phase 0: Make The Existing Responsibilities Visible

- [x] Split the old monolithic `render/image.go` into:
  - `types.go`
  - `image.go`
  - `commands.go`
  - `world_mesh.go`
  - `cpu_raster.go`
- [x] Verify with `go test ./...`.
- [x] Verify with `staticcheck .`.

## Phase 1: Introduce `Frame` As The Screen API

- [x] Add `render.Frame` as the public screen target type.
- [x] Change backend `Game.Draw`, overlay drawing, and GPU renderer entry points to accept `*render.Frame`.
- [x] Change game/UI draw signatures from `*render.Image` to `*render.Frame` where the value is the main screen.
- [x] Keep `Frame` temporarily backed by the old implementation so this phase is mechanical.
- [x] Verify no backend caller creates the screen with `NewImage`.
- [x] Verify with `go test ./...` and `staticcheck .`.

Acceptance: screen rendering APIs mention `Frame`, not `Image`, even if `Frame` still aliases the legacy structure internally.

## Phase 2: Physically Split `Frame` From `Image`

- [x] Replace the temporary alias with a real `Frame` struct.
- [x] Move these fields out of `Image` and into `Frame`:
  - `commands`
  - `worldCommands`
  - `worldMeshes`
  - `worldBillboards`
  - `uiRects`
  - `uiTextBoxes`
  - `uiTextLabels`
  - `clear`
  - `camera`
  - `screenScaleX`
  - `screenScaleY`
  - width/height, replacing `pix.Bounds()` for frame size.
- [x] Remove `screen bool` from `Image`.
- [x] Make `Frame.Bounds()` derive from width/height.
- [x] Move frame-only methods to `frame.go`.
- [x] Keep CPU `Image` methods image-only.
- [x] Verify with `go test ./...` and `staticcheck .`.

Acceptance: `Image` no longer knows whether it is a screen.

## Phase 3: Split Drawing APIs By Target

- [x] Move frame draw submission methods to `Frame`:
  - `DrawImage`
  - `DrawTriangles`
  - `DrawTrianglesOwned`
  - `DrawTriangles3D`
  - `DrawTriangles3DOwned`
  - `DrawWorldMesh`
  - `DrawWorldBillboard`
- [x] Keep CPU/offscreen composition methods on `Image`:
  - `DrawImage`
  - `DrawTriangles`
  - `Fill`
- [x] Replace transitional primitive helpers with frame-specific and image-specific APIs.
- [x] Replace `DrawRect(*Image, ...)` style APIs with either:
  - `DrawFrameRect(*Frame, ...)` for frame commands
  - `DrawImageRect(*Image, ...)` for CPU composition
- [x] Remove `if dst.screen` branches.
- [x] Verify with `go test ./...` and `staticcheck .`.

Acceptance: no render function branches on target kind.

## Phase 4: Make Text And UI Overlay Explicit

- [x] Move `uiRects`, `uiTextBoxes`, and `uiTextLabels` to a `FrameOverlay` or explicit `Frame` fields.
- [x] Keep `DrawUI*` functions as explicit frame overlay enqueue helpers.
- [x] Ensure UI overlay uses the same gogpu/ui text path as FPS/speech/tooltips.
- [x] Keep cached text images as `Image` texture data.
- [x] Verify with `go test ./...` and `staticcheck .`.

Acceptance: text overlay is a frame command path, not hidden inside `Image`.

## Phase 5: Separate CPU Texture Generation

- [x] Audit all `NewImage` + `RGBA()` writers.
- [x] Group CPU texture generation helpers near their owners:
  - sprite billboard generation
  - item marker generation
  - minimap generation
  - text cache generation
- [x] Keep `render.Image` only as the common CPU texture container.
- [x] Avoid adding new CPU composition paths for per-frame drawing.
- [x] Verify with `go test ./...` and `staticcheck .`.

Acceptance: CPU image writes happen only for asset/cache generation, not ordinary frame rendering.

## Phase 6: Make GPU Resources Explicit

- [x] Rename internal `gpuTexture` cache entries to distinguish them from CPU `Image`.
- [x] Ensure `Image.version` is only used for CPU texture upload invalidation.
- [x] Keep retained meshes as retained GPU-ish resources:
  - CPU-side vertex/index source in `WorldMesh`
  - GPU buffers in renderer cache
- [x] Remove unnecessary `RGBA()` exposure from hot render paths where possible.
- [x] Verify with `go test ./...` and `staticcheck .`.

Acceptance: CPU image data and GPU texture resources are clearly different concepts.

## Phase 7: Clean Compatibility Names

- [x] Remove `NewScreenImage`.
- [x] Remove any `screen` field or screen-only method from `Image`.
- [x] Rename files if needed:
  - `commands.go` -> `frame_commands.go` and `image_draw.go`
  - `cpu_raster.go` stays CPU-only.
- [x] Update docs/comments in `render/doc.go`.
- [x] Verify with `go test ./...` and `staticcheck .`.

Acceptance: `render/` no longer reads like an Ebiten compatibility layer.

## Phase 8: Final Audit

- [x] `rg "screen" render` only finds framebuffer/surface concepts, not `Image` mode switches.
- [x] `rg "RGBA\\(\\)"` has no main-frame hot path usage except texture upload/cache generation.
- [x] `go test ./...`
- [x] `staticcheck .`
- [x] Manual smoke test:
  - [x] login screen
  - [x] map rendering
  - [x] sprites/effects
  - [x] UI windows
  - [x] FPS meter/speech/tooltips
