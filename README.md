# goro

`goro` is an open Ragnarok Online client recreation implemented in Go.

The runtime uses GoGPU/wgpu for the window and presentation path, with a modern
GPU pipeline and Vulkan support. Built 100% in Go without CGO, it is fully statically
compiled and can be easily deployed.

## Goals

- Faithfully reimplement the Ragnarok Online client.
- Focus on the pre-renewal 2008 experience first.
- Stay pure Go, without CGO, so cross-compilation and deployment stay simple on
  many platforms.
- Use a modern GPU pipeline through GoGPU, including Vulkan and Wayland support.
- Deliver good performance, including support for high-refresh-rate displays.
- Provide a modernized, neat themeable UI built with `gogpu/ui`.
- Keep the engine reusable for creating new MMORPGs.
- Become a drop-in replacement for `Ragexe` and `Sakexe`.

### Stretch goals:

- Provide GRF tooling.
- Provide map, sprite, and model viewers.
- Support optional Lua scripting for autoplay and automation experiments.
- Support more Ragnarok Online client versions.
- Support optional anti-cheat and security features.

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

Mostly done:

 * Login
 * Character selection
 * Character creation
 * Maps display
   * Water
   * Map sounds
   * Lightmaps
   * Fog (innacurate)
   * Animated models
   * Weather effects
 * Camera, zoom, rotation
 * Battle
   * Enemies
   * Drops
   * Jobs
     * Novice
     * 1-1
       * Swordman
       * Magician
       * Archer
       * Acolyte
       * Thief
   * Skill effects
 * UI
   * Basic information
   * Button bar
   * Shortcuts bar
   * Console
   * Minimap
   * Items
   * Equipment
   * Option
     * Settings
   * Friends
   * Party
   * Skills (flat version)
   * Cart Storage
   * Kafra Storage
   * Teleport skill modal
   * Warp skill modal
   * Cart appearance modal
 * Emotes
