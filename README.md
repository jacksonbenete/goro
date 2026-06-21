# goro

`goro` is a Go Ragnarok Online client foundation.

The current runtime uses Ebitengine with its OpenGL backend. Ebitengine `v2.9.x`
does not expose Vulkan as a public graphics-library choice, so the code keeps the
renderer isolated behind a small backend boundary and favors a working OpenGL
path first.

## Run

```sh
go run ./cmd/goro
```

For the local OldRO + rAthena test setup:

```sh
./scripts/run-oldro.sh
```

Useful toggles:

```sh
GORO_NET_TRACE=1 ./scripts/run-oldro.sh
GORO_DEBUG_RSW_MARKERS=1 ./scripts/run-oldro.sh
GORO_RENDER_RSM=0 ./scripts/run-oldro.sh
GORO_RSM_MAX_FACES=1500 ./scripts/run-oldro.sh
GORO_RSM_RENDER_RADIUS=60 ./scripts/run-oldro.sh
GORO_DEBUG_RSM_TRANSFORMS=1 ./scripts/run-oldro.sh
GORO_SCENE_PROJECTION=flat ./scripts/run-oldro.sh
GORO_SCENE_HEIGHT_SCALE=2.8 ./scripts/run-oldro.sh
GORO_CAMERA_PITCH=230 GORO_CAMERA_YAW=0 GORO_CAMERA_ZOOM=420 GORO_CAMERA_FOV=15 ./scripts/run-oldro.sh
GORO_CAMERA_TARGET_Z=8 ./scripts/run-oldro.sh
GORO_CAMERA_FOLLOW_TERRAIN_HEIGHT=1 ./scripts/run-oldro.sh
GORO_PACKET_CLIENT_DATE=20080910 ./scripts/run-oldro.sh
GORO_BUILD=0 ./scripts/run-oldro.sh
```

Runtime data is discovered from, in order:

- `GORO_DATA_DIR`
- `OPEN_MIDGARD_DATA_DIR`
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
- `render` Ebitengine/OpenGL backend
- `input` per-frame input snapshot

It is not yet a complete RO implementation. The next substantial steps are GRF
loading, packet serializers for account/char/map login, and map asset parsers.
