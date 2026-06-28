# Technical Debt

This file tracks intentionally ugly or temporary implementation choices that should
not become invisible project assumptions.

## Runtime And Portability

- **Lua 5.1 `.lub` execution**
  - Current state: RO `.lub` bytecode is parsed in `res/lub.go`, translated into
    GopherLua bytecode, then executed by GopherLua.
  - Ugly part: GopherLua requires a private `FunctionProto.stringConstants`
    field for `GETGLOBAL` and `SETGLOBAL`, so we currently set that field with
    `reflect` and `unsafe`.
  - Better fix: upstream or fork a small public API in GopherLua for externally
    built prototypes, or own a minimal Lua 5.1 bytecode runner for RO data files.
  - Also revisit whether the old fallback evaluator is still needed after enough
    real `.lub` files are covered by tests.

- **MP3 playback through runtime `libmpg123`**
  - Current state: BGM uses `go-mp3` where possible and falls back to `libmpg123`
    through `purego` for old low-rate RO MP3 files.
  - Ugly part: this keeps `CGO_ENABLED=0`, but still needs a platform library at
    runtime and manually registers mpg123 symbols.
  - Better fix: find or write a pure-Go MP3 decoder that correctly handles the
    old 22 kHz files, or convert/cache decoded audio in a controlled asset
    pipeline.

- **`nofakecgo` build tag requirement**
  - Current state: builds are expected to use `CGO_ENABLED=0 -tags nofakecgo`.
  - Ugly part: the tag exists to avoid duplicate fake-cgo symbol providers across
    GoGPU/goffi and Oto/purego.
  - Better fix: remove the tag requirement once the dependency stack no longer
    needs competing fake-cgo shims.

- **Dynamic library loading paths**
  - Current state: `audio/dlopen_*.go` probes common mpg123 library names.
  - Ugly part: this is best-effort and may fail silently into lower-quality or no
    audio depending on the platform install.
  - Better fix: explicit dependency diagnostics in the launcher, packaged
    per-platform libraries, or no runtime library dependency.

## Rendering

- **Dynamic world geometry**
  - Current state: static GND surfaces, lightmapped GND subdivisions, and RSM
    placements are retained `render.WorldMesh` objects. They are built once for a
    loaded map/model placement, uploaded into persistent GPU buffers, and reused.
  - Current state: actor/NPC/mob/item/effect sprite billboards use the dedicated
    shared-quad billboard pipeline with per-instance data instead of per-frame
    quad vertex slices.
  - Current state: shader-side camera projection and fog are used for retained
    meshes and billboards; render stats expose `world_mesh_commands`,
    `retained_world_meshes`, `world_billboards`, and remaining dynamic
    `world_vertices`.
  - Remaining dynamic geometry is intentional: animated water waves, warp/effect
    cylinders/rings, STR effects with animated per-corner coordinates, tile
    cursor geometry, and debug/fallback paths.
  - Better fix: move water and common effect primitives to specialized
    instanced/procedural GPU pipelines so `world_vertices` mostly measures only
    rare debug/fallback geometry.

- **Billboard depth and clipping**
  - Current state: sprites use 3D-ish billboard placement with custom depth
    handling to avoid being wrongly hidden by terrain/buildings.
  - Ugly part: RO sprites are not physically normal 3D quads; correct occlusion
    still needs policy knobs and reference-client comparison.
  - Better fix: document and implement the exact roBrowser/OpenMidgard approach
    for sprite depth bias, top/head clipping, and object interaction.

- **Fog and lighting model**
  - Current state: fog and lighting were tuned visually against RO clients.
  - Ugly part: some constants are empirical.
  - Improved: renderer-level camera fog is now applied in the world shader
    instead of mutating every submitted 3D vertex on the CPU.
  - Better fix: derive formulas and default parameters directly from RSW/GND data
    and reference-client code, then add screenshot-style regression fixtures.

## Gameplay State

- **Local death as held animation**
  - Current state: local player death uses the transient actor animation path,
    with the final death frame held until positive HP or map change.
  - Ugly part: roBrowser models death as persistent entity action/state, not as
    an expiring combat-style animation.
  - Better fix: split persistent actor state such as idle, walk, sit, and dead
    from transient overlays such as attack, hurt, and pickup, then clear dead
    state only on resurrection, respawn, or map transition.

## Data And Protocol

- **Fallback resource tables**
  - Current state: some NPC/monster/job/resource names still have hardcoded
    fallback maps.
  - Ugly part: these can hide `.lub` or DB parsing bugs and can become stale.
  - Better fix: rely on client data tables first, keep fallbacks only as explicit
    diagnostics or emergency compatibility.

- **Packet-version assumptions**
  - Current state: packet client date/profile defaults target the current rAthena
    setup.
  - Ugly part: several protocol choices are still date/profile-sensitive.
  - Better fix: centralize packetver negotiation/configuration and document which
    server builds are supported.

- **Real-data tests**
  - Current state: real RO data tests are gated by `GORO_DATA_DIR`.
  - Ugly part: CI cannot validate the most important compatibility paths without
    proprietary data.
  - Better fix: keep small synthetic fixtures for parsers and maintain a local
    real-data test checklist for releases.
