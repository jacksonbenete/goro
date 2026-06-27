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

- **CPU-side geometry construction**
  - Current state: GND, RSM, sprites, billboards, water, fog, and lighting are
    assembled mostly on the CPU before being submitted through GoGPU.
  - Ugly part: much of this still reflects the old 2D-engine workaround history.
  - Improved: render commands now keep compact `uint16` indices until final GPU
    batching, and RSM node matrices are cached per loaded model instead of being
    rebuilt every frame.
  - Improved: hot geometry builders can now transfer owned vertex/index slices
    into render commands, avoiding a second copy for RSM batches, lightmapped GND
    patches, sprite billboards, and common quads.
  - Improved: GPU frame assembly now pre-sizes its transient command maps,
    vertex buffers, index buffers, and batch lists from queued draw commands.
  - Better fix: move more work into normal GPU pipelines with stable vertex/index
    buffers, shader-side fog/lighting, and fewer per-frame allocations.

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
