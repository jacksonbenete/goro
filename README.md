# goro

`goro` is an open Ragnarok Online client recreation implemented in Go.

The runtime uses GoGPU/wgpu for the window and presentation path, with a modern
GPU pipeline and Vulkan support. Built 100% in Go without CGO, it is fully statically
compiled and can be easily deployed.

This project wouldn't be possible without the existence of other open source clients
like ROBrowser Legacy and Open Midgard and their reverse engineering efforts.

## Project Goals

- Faithfully reimplement the original Ragnarok Online client.
- Focus on the pre-renewal 2008 experience first.
- Stay pure Go, without CGO, so cross-compilation and deployment stay simple on
  many platforms.
- Aim for simple, readable, hackable codebase.
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
CGO_ENABLED=0 go run -tags nofakecgo . --data-dir ~/kRO --fullscreen
CGO_ENABLED=0 go run -tags nofakecgo . --config ./goro.ini --bgm=false --graphics-api vulkan
```

Useful options:

```sh
--net-trace
--packet-client-date 20211103 # only when rAthena is rebuilt for that packetver
--fullscreen
--bgm=false
--bgm-volume 0.35
--graphics-api gles # fallback if Vulkan is unavailable
```

When using the default 2008 packet profile with local rAthena, set
`char_del_delay: 0` in `conf/char_athena.conf`. The delayed character deletion
flow is for 2010-08-03+ clients; the 2008 flow deletes directly after the email
check.

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
   * Indoors
 * Camera, zoom, rotation
 * Battle and Gameplay
   * Enemies
   * Path finding
   * Continuous held-click walking
   * Smooth camera following and zoom
   * Drops
   * Playable characters animation chain
   * Attack-ready stance
   * Jobs
     * Novice
     * 1-1
       * Swordman
       * Magician
       * Archer
       * Acolyte
       * Thief
   * Skill effects
   * Skill casting
   * Walk cancellation
   * Casting cancellation
   * Cursor snap
   * Noshift
   * Noctrl
   * Item drops
   * Item identification
   * Card composition
   * Trading
   * Vending
   * Show equipment
   * Pets
     * Capture slot machine
     * Egg hatching
     * Feeding
     * Status window and rename
     * Accessory equip/unequip
     * Performance actions
   * Friends
   * Parties
   * Whispers
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
   * Party & party settings
   * Stats
   * Skills (flat version)
   * Cart Storage
   * Kafra Storage (simple)
   * Teleport skill modal
   * Warp skill modal
   * Cart appearance modal
   * Trade window
   * Vending windows
   * Card composition window
   * Show-equipment window
   * Item and skill tooltips
   * Status icons with roBrowser-sourced metadata
 * Emotes
 * Overlay text
   * FPS meter
   * Character names and HP/SP bars
   * Speech bubbles
