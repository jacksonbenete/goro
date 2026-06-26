# goro

`goro` is a Go Ragnarok Online client foundation.

The current runtime uses GoGPU/wgpu for the window and presentation path. The
first migration pass renders into a compatibility canvas and uploads the frame to
GoGPU each frame; hot paths can move to native GPU pipelines incrementally.
The default launcher uses `CGO_ENABLED=0`; audio is currently silent through the
no-op audio stub on this path.

## Run

```sh
CGO_ENABLED=0 go run .
```

For the local OldRO + rAthena test setup:

```sh
./scripts/run-oldro.sh
```

Useful toggles:

```sh
GORO_NET_TRACE=1 ./scripts/run-oldro.sh
GORO_PACKET_CLIENT_DATE=20211103 ./scripts/run-oldro.sh # only when rAthena is rebuilt for that packetver
GORO_FULLSCREEN=1 ./scripts/run-oldro.sh
GORO_BGM=0 ./scripts/run-oldro.sh
GORO_BGM_VOLUME=0.35 ./scripts/run-oldro.sh
GOGPU_GRAPHICS_API=gles ./scripts/run-oldro.sh # fallback if Vulkan is unavailable
```

Runtime data is discovered from, in order:

- `GORO_DATA_DIR`
- current working directory

The resource manager currently looks for loose files such as:

- `data/clientinfo.xml`
- `data/sclientinfo.xml`
- `clientinfo.xml`
- `sclientinfo.xml`
- `System/clientinfo.xml`
- `System/sclientinfo.xml`

## Current Scope

This first pass establishes the same broad subsystem boundaries used by
OpenMidgard:

- `core` startup configuration
- `res` runtime data discovery and `clientinfo.xml` parsing
- `network` TCP connection and RO packet framing
- `session` account/character/session state
- `world` map and actor state
- `gamemode` boot, login/server selection, and world modes
- `render` GoGPU backend
- `input` per-frame input snapshot

It is not yet a complete RO implementation. The next substantial steps are GRF
loading, packet serializers for account/char/map login, and map asset parsers.
