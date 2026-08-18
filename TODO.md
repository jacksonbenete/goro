# TODO

This file tracks intentionally ugly or temporary implementation choices, plus
large gameplay/rendering work that should not become invisible project
assumptions.

## Technical Debt

### Runtime And Portability

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

- **`nofakecgo` build tag requirement**
  - Current state: builds are expected to use `CGO_ENABLED=0 -tags nofakecgo`.
  - Ugly part: the tag exists to avoid duplicate fake-cgo symbol providers in
    the pure-Go GoGPU dependency stack.
  - Better fix: remove the tag requirement once the dependency stack no longer
    needs competing fake-cgo shims.

### Rendering

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
  - Better fix: document and implement the exact reference-client/OpenMidgard approach
    for sprite depth bias, top/head clipping, and object interaction.

- **Lighting model**
  - Current state: fog uses reference client's camera defaults, fog-table scaling,
    shader-side depth formula, and non-additive color mix. Additive GND lightmap
    layers are explicitly exempt from fog so fog is applied once to the composed
    world color.
  - Ugly part: parts of the RSW/GND lighting pipeline are still approximations,
    especially the split base/lightmap rendering and per-map visual parity.
  - Better fix: collapse GND base, lightmap alpha, posterized light color, and
    fog into one shader path like the reference client, then add screenshot-style regression
    fixtures for representative outdoor, indoor, and dungeon maps.

### Gameplay State

- **Local death as held animation**
  - Current state: local player death uses the transient actor animation path,
    with the final death frame held until positive HP or map change.
  - Ugly part: reference client models death as persistent entity action/state, not as
    an expiring combat-style animation.
  - Better fix: split persistent actor state such as idle, walk, sit, and dead
    from transient overlays such as attack, hurt, and pickup, then clear dead
    state only on resurrection, respawn, or map transition.

### Data And Protocol

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

## Effects

Source of truth: roBrowser's `src/DB/Effects/EffectConst.js`,
`src/DB/Effects/EffectTable.js`, `src/DB/Skills/SkillEffect.js`,
`src/DB/Skills/SkillUnit.js`, and `src/DB/Items/ItemEffect.js`.

### Status

- Reference active numeric `EffectTable.js` entries: `607`.
- Reference numeric effect constants: `1147`.
- Goro implemented active reference effect IDs: `102`.
- Current active-table coverage: about `16.8%`.
- Current all-constant coverage: about `8.9%`.
- Coverage only counts world/effect IDs handled through `worldEffectSpecForID`
  and direct effect trigger mappings. It does not count UI rendering, damage
  number rendering, cursor drawing, actor sprite animation, or map logic that is
  not modeled as an effect ID.

### Current Shape

- Effects are primarily data-table driven through `skillEffectSpecs`,
  `worldEffectSpecs`, and parser-backed tests against roBrowser's effect table.
- Renderer support exists for STR, CYLINDER, 2D billboards, 3D billboards,
  SPR/ACT attachments, sprite-backed 3D effects, duplicate timing, delays,
  basic alpha/fade, color tint, size interpolation, and source-to-target motion.
- Special Go behavior is still allowed for effects that are not purely
  declarative in roBrowser data, such as camera shake, map transitions, or UI
  choices produced by a skill.
- RSW-placed map effects and map-wide weather are tracked separately in
  [docs/map-effects-weather-todo.md](docs/map-effects-weather-todo.md).

### Recently Covered Families

- Swordman first-job effects: Bash, Provoke, Magnum Break, Endure.
- Mage first-job effects: Napalm Beat, Soul Strike, Cold Bolt, Frost Diver,
  Stone Curse, Fire Ball, Fire Wall, Fire Bolt, Lightning Bolt, Thunder Storm,
  Sight, Safety Wall, and Energy Coat routing/effects where currently supported.
- Acolyte first-job effects: Heal, Increase/Decrease Agility, Aqua Benedicta,
  Signum Crucis, Angelus, Blessing, Cure, Ruwach, Pneuma, Teleport, Warp Portal,
  and Holy Light.
- Thief first-job effects: Steal, Hiding, Envenom, Detoxify, Sprinkle Sand,
  Throw Stone.
- Archer first-job effects: Improve Concentration, Double Strafe, Arrow Shower,
  arrow projectile and ammo handling.
- Merchant first-job effects: Mammonite, Cart Revolution, Crazy Uproar,
  cart state/appearance, vending bubbles and shop interaction.
- Ninja expanded-class support: normal/baby jobs, full skill tree and level
  metadata, roBrowser skill actions/effects, and Ninja status routing.
- Common effects: hit feedback, potion/food families, Heal/recovery feedback,
  base/job level-up, teleport/portal, refine/pharmacy success/failure.

### Priority Backlog

- [ ] Improve STR parity: blend modes, alpha edge behavior, texture frame
  selection, UV animation, layer timing, rotation/scale interpolation, source
  offsets, and subtractive/additive modes.
- [ ] Improve CYLINDER parity: partial circles, repeat texture axes, fixed
  perspective versus world-space orientation, rotation deltas, duplicate
  instances, randomized fields, and exact alpha behavior.
- [ ] Complete reference `wav` behavior: delayed sounds, randomized sounds,
  positional sounds, and no inferred sounds when roBrowser declares none.
- [ ] Finish generic routing for `SkillEffect.js`, `SkillUnit.js`, and
  `ItemEffect.js` so adding new effects is mostly data import instead of custom
  Go code.
- [ ] Add persistent attached lifecycle handling for buffs/status loops:
  follow actors while active, detach only when the reference client does, and
  clear on vanish/death/status removal.
- [ ] Add a debug command or CLI flag that spawns an effect ID at the player for
  fast visual comparison.
- [ ] Generate an `effects_coverage.json` style report with support levels:
  `none`, `sound-only`, `partial`, `close`, and `parity`.

### Gameplay-Visible Gaps

- [ ] Remaining hit variants: `EF_HIT1`, `EF_HIT3`, `EF_HIT4`, `EF_HIT5`,
  `EF_HIT6`, elemental/poison hit variants, miss/critical/support message
  effects, and blocked/zero-damage feedback.
- [ ] Ground skill units from `SkillUnit.js`: Sanctuary, Magnus, Quagmire,
  traps, Bard/Dancer songs, Volcano, Deluge, Violent Gale, Land Protector,
  Spider Web, Basilica, Suiton, Epiclesis, and their server update/removal
  lifecycle.
- [ ] Status loops: poison, curse, blind, silence, stun, sleep, confusion,
  frozen, stone curse, hiding/cloaking/invisible states, spirit spheres, aura
  loops, and actor-attached buff/debuff decorations.
- [ ] Second-job pre-renewal sweep: Knight/Crusader, Wizard/Sage,
  Priest/Monk, Assassin/Rogue, Hunter/Bard/Dancer, Blacksmith/Alchemist.
- [ ] Remaining expanded and later content: Super Novice, Gunslinger,
  transcendent classes, renewal third jobs, and modern fourth-job effect IDs.
