# goro

`goro` is a Go Ragnarok Online client foundation.

The current runtime uses GoGPU/wgpu for the window and presentation path. The
first migration pass renders into a compatibility canvas and uploads the frame to
GoGPU each frame; hot paths can move to native GPU pipelines incrementally.
Build and run with `CGO_ENABLED=0` and `-tags nofakecgo`. The tag lets GoGPU's
goffi use Oto/purego's fake cgo runtime symbols instead of defining a second
copy, which keeps pure-Go audio enabled.

BGM playback uses `libmpg123` at runtime when available. This is needed for old
22 kHz Ragnarok MP3 tracks; without it, the client falls back to the pure-Go MP3
decoder, which is less accurate for those files.

## Run

```sh
CGO_ENABLED=0 go run -tags nofakecgo .
```

Configuration is loaded from `goro.ini` in the current directory when the file
exists. Pass another file with `--config`:

```ini
data_dir = /home/kivutar/Téléchargements/OldRO

[window]
width = 1280
height = 720
fullscreen = false

[packet]
client_date = 20080910
profile = 23

[audio]
bgm = true
bgm_volume = 0.55

[render]
graphics_api = vulkan
vsync = true

[network]
trace = false
```

Command-line options override the ini file:

```sh
CGO_ENABLED=0 go run -tags nofakecgo . --data-dir /home/kivutar/Téléchargements/OldRO --fullscreen
CGO_ENABLED=0 go run -tags nofakecgo . --config ./oldro.ini --bgm=false --graphics-api gles
```

Useful options:

```sh
CGO_ENABLED=0 go run -tags nofakecgo . --net-trace
CGO_ENABLED=0 go run -tags nofakecgo . --packet-client-date 20211103 # only when rAthena is rebuilt for that packetver
CGO_ENABLED=0 go run -tags nofakecgo . --fullscreen
CGO_ENABLED=0 go run -tags nofakecgo . --bgm=false
CGO_ENABLED=0 go run -tags nofakecgo . --bgm-volume 0.35
CGO_ENABLED=0 go run -tags nofakecgo . --graphics-api gles # fallback if Vulkan is unavailable
```

Runtime data is discovered from, in order:

- `--data-dir`
- `data_dir` in `goro.ini`
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
- `gamemode` login/server selection and world modes
- `render` GoGPU backend
- `input` per-frame input snapshot

It is not yet a complete RO implementation. The next substantial steps are GRF
loading, packet serializers for account/char/map login, and map asset parsers.
