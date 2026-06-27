# Effects TODO

This tracks Goro's world/effect coverage against roBrowser's
`src/DB/Effects/EffectConst.js` and `src/DB/Effects/EffectTable.js`.

## Completion

Metric used: active numeric entries in roBrowser `EffectTable.js`.

- roBrowser active numeric effect table entries: 607
- roBrowser numeric effect constants: 1147
- Goro implemented active roBrowser effect IDs: 45
- Completion against active roBrowser table: 7.4%
- Completion against all roBrowser numeric constants: 3.8%

This only counts world/effect IDs handled by `worldEffectSpecForID` and direct
effect trigger mappings. It does not count unrelated UI rendering, damage number
rendering, cursor drawing, or actor sprite animation.

## Done

- [x] `EF_HIT2` `1`: Bash hit / simple hit burst.
- [x] `EF_COIN` `10`: Mammonite STR effect.
- [x] `EF_ENDURE` `11`: Endure billboard effect.
- [x] `EF_SOULSTRIKE` `15`: Soul Strike file-backed 3D particle layer,
  sprite-backed 3D projectile layer, and SFX; projectile trajectory is still
  approximate.
- [x] `EF_BASH` `16`: Bash start cylinder effect.
- [x] `EF_MAGNUMBREAK` `17`: Magnum Break cylinder effect and camera shake.
- [x] `EF_STEAL` `18`: Steal file-backed 3D particles and SFX.
- [x] `EF_PATTACK` `20`: Envenom file-backed 3D particles and SFX.
- [x] `EF_DETOXICATION` `21`: Detoxify file-backed 3D particles and SFX.
- [x] `EF_STONECURSE` `23`: Stone Curse STR effect.
- [x] `EF_FIREBALL` `24`: Fire Ball sprite-backed 3D caster projectile and SFX;
  projectile trajectory is still approximate.
- [x] `EF_FIREWALL` `25`: Fire Wall STR effect.
- [x] `EF_FROSTDIVER2` `28`: Frost Diver hit STR effect.
- [x] `EF_LIGHTBOLT` `29`: Lightning Bolt STR effect.
- [x] `EF_THUNDERSTORM` `30`: Thunder Storm STR effect.
- [x] `EF_INCAGILITY` `37`: Increase Agility declared SFX; 3D visual still pending.
- [x] `EF_DECAGILITY` `38`: Decrease Agility declared SFX; 3D visual still pending.
- [x] `EF_AQUA` `39`: Aqua Benedicta SPR attachment and SFX.
- [x] `EF_SIGNUM` `40`: Signum Crucis STR effect.
- [x] `EF_ANGELUS` `41`: Angelus STR effect.
- [x] `EF_BLESSING` `42`: Blessing SPR attachment, sprite-backed 3D particles,
  file-backed 3D glow, and SFX.
- [x] `EF_FIREHIT` `49`: Fire elemental hit STR effect.
- [x] `EF_FIRESPLASHHIT` `50`: Fire splash 2D textured billboard effect.
- [x] `EF_COLDHIT` `51`: Cold elemental hit declared SFX; 2D visual still pending.
- [x] `EF_WINDHIT` `52`: Wind elemental hit STR effect.
- [x] `EF_CURE` `66`: Cure STR effect.
- [x] `EF_CONCENTRATION` `153`: Improve Concentration STR effect.
- [x] `EF_REFINEOK` `154`: Refine Success STR effect.
- [x] `EF_REFINEFAIL` `155`: Refine Fail STR effect.
- [x] `EF_PROVOKE` `67`: Provoke STR effect.
- [x] `EF_JOBLVUP` `158`: Job level-up STR effect, no explicit SFX in roBrowser.
- [x] `EF_POTION1` `204`: Red potion STR effect.
- [x] `EF_POTION2` `205`: Orange potion STR effect.
- [x] `EF_POTION3` `206`: Yellow potion STR effect.
- [x] `EF_POTION4` `207`: White potion STR effect.
- [x] `EF_POTION5` `208`: Blue potion STR effect.
- [x] `EF_POTION6` `209`: Green potion STR effect.
- [x] `EF_POTION7` `210`: Food/consumable STR effect.
- [x] `EF_POTION8` `211`: Blue food/consumable STR effect.
- [x] `EF_TELEPORTATION2` `304`: Teleport / Fly Wing style cylinder stack.
- [x] `EF_PHARMACY_OK` `305`: Pharmacy Success STR effect.
- [x] `EF_PHARMACY_FAIL` `306`: Pharmacy Fail STR effect.
- [x] `EF_HEAL` `312`: Heal cylinders, file-backed 3D particles, and SFX.
- [x] `EF_PORTAL2` `317`: Warp Portal skill unit cylinder stack.
- [x] `EF_ANGEL` `371`: Base level-up STR effect with `levelup.wav`.

## High Priority Backlog

- [ ] Generic effect-table interpreter for roBrowser component types:
  `STR`, `CYLINDER`, `SPRITE`, `PARTICLE`, `FUNC`, `QUAKE`, repeated effects,
  randomized fields, delayed sound, delayed components, and attached/detached
  entity semantics.
- [ ] Generic `wav` handling from roBrowser effect components. Only play sounds
  when the component has `wav`, instead of inventing names from STR filenames.
- [ ] Generic `ZC_NOTIFY_EFFECT` mapping beyond base/job level-up.
  rAthena defines refine success/failure, pharmacy success/failure, Super Novice
  level-up variants, Taekwon level-up variants, and game-over.
- [ ] Generic skill effect routing from roBrowser `SkillEffect.js`:
  begin/caster effect, hit effect, ground effect, before-hit effect, and skill
  no-damage/success effects.
- [ ] Generic skill-unit routing from roBrowser `SkillUnit.js`, not only
  `UNT_WARPPORTAL`.
- [ ] Generic item effect routing from roBrowser `ItemEffect.js`, not only
  potion/food families and Butterfly Wing.

## Core Combat And Common Effects

- [ ] Remaining hit variants: `EF_HIT1`, `EF_HIT3`, `EF_HIT4`, `EF_HIT5`,
  `EF_HIT6`, elemental hits, poison hits, miss/critical/support message effects.
- [ ] Begin spell / casting aura: `EF_BEGINSPELL` and modern cast variants.
- [ ] Weapon/attack special effects beyond swordman basics.
- [ ] Healing and recovery effects: Heal, Increase Agi, Blessing, Angelus,
  Aqua Benedicta, Kyrie, Pneuma, Safety Wall, Sanctuary, Ruwach, Sight.
- [ ] Status effects: poison, curse, blind, silence, stun, frozen, stone curse,
  sleep, confusion, hiding/cloaking, and their state loops.
- [ ] Death, revive, resurrection, warp entry/exit, map entry, and old portal
  variants: `EF_ENTRY`, `EF_EXIT`, `EF_WARP`, `EF_TELEPORTATION`,
  `EF_READYPORTAL`, `EF_READYPORTAL2`.

## First/Second Job Skills

- [ ] Swordman remaining effects: Two-Hand Quicken, Counter Attack, Auto Berserk,
  Knight/Peco/Knight spear effects, Bowling Bash, Brandish Spear.
- [ ] Mage/Wizard effects: bolts, Fire Ball, Fire Wall, Frost Diver, Thunderstorm,
  Soul Strike, Napalm Beat, Safety Wall, Storm Gust, Meteor Storm, Lord of
  Vermilion, Jupitel Thunder, Water Ball, Quagmire.
- [ ] Acolyte/Priest effects: Heal, Blessing, Inc/Dec Agi, Angelus, Cure, Aqua,
  Signum, Pneuma, Warp Portal, Teleport UI/result handling, Kyrie, Magnificat,
  Gloria, Lex, Aspersio, Sanctuary, Resurrection, Magnus.
- [ ] Thief/Assassin/Rogue effects: Steal, Hiding, Envenom, Detoxify, Double
  Attack feedback, Sonic Blow, Grimtooth, Cloaking, Enchant Poison, Back Stab,
  Raid, Strip skills.
- [ ] Archer/Hunter/Bard/Dancer effects: arrows, Double Strafe, Arrow Shower,
  traps, Falcon effects, songs/dances, ensemble ground effects.
- [ ] Merchant/Blacksmith/Alchemist effects: Mammonite, Cart Revolution,
  Overthrust, Weapon Perfection, Adrenaline Rush, Hammer Fall, vending/shop
  feedback, potion creation success/failure.

## Expanded, Transcendent, Third Job, And Later

- [ ] Super Novice level-up variants and class-specific level-up effects.
- [ ] Taekwon/Star Gladiator/Soul Linker effects, including soul-link effects.
- [ ] Ninja/Gunslinger effects.
- [ ] Transcendent class effects not covered by first/second job sections.
- [ ] Third job and renewal effects: Rune Knight, Warlock, Ranger, Guillotine
  Cross, Mechanic, Genetic, Arch Bishop, Sura, Royal Guard, Sorcerer, Minstrel,
  Wanderer, Shadow Chaser.
- [ ] Modern fourth/expanded class effect IDs present in newer roBrowser tables.

## Ground, Map, Weather, And Environment

- [ ] Ground skill zones from roBrowser `SkillUnit.js`: Safety Wall, Fire Wall,
  Pneuma, Sanctuary, Magnus, Volcano, Deluge, Violent Gale, Land Protector,
  Spider Web, Basilica, Suiton, Epiclesis, and others.
- [ ] Map/weather effects: rain, snow, sakura, fog-like weather, thunder, torches,
  smoke, fireflies, portal-like map RSW effects.
- [ ] Persistent attached effects with correct lifecycle and removal when actors
  vanish, die, hide, change state, or leave the visible area.
- [ ] Effect depth/order policy for translucent world effects versus sprites,
  terrain, RSM models, water, fog, and UI overlays.

## Rendering Correctness Debt

- [ ] STR renderer parity: color blend modes, additive/subtractive behavior,
  alpha edge handling, texture frame selection, UV animation, rotation and scale
  interpolation, source offsets, and layer timing.
- [ ] Cylinder parity: roBrowser blend modes, repeat texture axes, partial circle
  sides, fixed perspective versus world-space orientation, late rotation deltas,
  duplicate/randomized component fields.
- [ ] Sprite-effect parity: SPR/ACT-based effect components, not only actor
  sprites and STR textures.
- [ ] Sound parity: roBrowser delayed `wav`, randomized `wav`, positional sound,
  and no sound for STR-only entries such as `EF_JOBLVUP`.
- [ ] Data-driven table generation from roBrowser-style definitions or a local
  normalized effect table, so adding effects does not require hardcoded Go
  switch cases for every ID.

## Ideal Implementation Order

The ideal order is not numeric effect ID order. It should first remove
hardcoded one-off work, then cover the effects players see constantly, then
expand through class and modern-content coverage.

### 1. Data-Driven Effect Core

Goal: stop writing one Go switch case per effect.

- [ ] Build a normalized local effect table generated from roBrowser
  `EffectTable.js` for the subset Goro can render today.
- [x] Add a parser for the roBrowser `EffectTable.js` subset Goro can render
  today: `STR`, `CYLINDER`, declared `wav`, timing, alpha, size, height, and
  rotation fields.
- [ ] Support remaining core roBrowser component fields: `type`, `file`, `texturePath`,
  `textureName`, `wav`, `delayWav`, `duration`, `delay`, `fade`, `fadeIn`,
  `fadeOut`, `alphaMax`, color channels, `blendMode`, `attachedEntity`,
  `renderBeforeEntities`, repeat flags, and randomized fields.
- [x] Move current manually supported specs into a table-backed path.
- [ ] Keep Go special handlers only for behavior that is not declarative in
  roBrowser, such as camera quake or game-state-triggered fade.
- [x] Add a test that tracks current implemented effect count and roBrowser
  coverage constants.
- [x] Add a parser-backed test that compares the current local roBrowser
  `EffectTable.js` active numeric entry count against Goro's coverage constant.
- [ ] Add a generated-table test that compares implemented effect IDs against
  normalized roBrowser table data instead of fixed constants.

### 2. Renderer Primitives Needed By Many Effects

Goal: make the generic table useful before adding many effects.

- [ ] Improve STR parity: blend modes, alpha, rotation/scale interpolation,
  UV animation, texture frame selection, layer timing, and source offsets.
- [ ] Improve CYLINDER parity: partial circles, repeat texture axes, fixed
  perspective versus world-space orientation, rotation deltas, duplicate
  instances, and randomization.
- [x] Add basic 2D textured billboard effect components with alpha, fade,
  size interpolation, rotation, and vertical offset.
- [x] Add basic file-backed 3D particle billboard components with alpha, fade,
  color tint, size interpolation/randomization, duplicate timing, delays, and
  randomized start/end offsets.
- [x] Add basic SPR/ACT effect components for roBrowser `SPR` attachments.
- [x] Add basic sprite-backed 3D effect components, because many RO effects use
  `spriteName`/`absoluteSpriteName` inside `3D` entries rather than texture
  `file` entries.
- [ ] Add sprite-backed 3D projectile trajectory fields: `toSrc`, `toTarget`,
  `rotateToTarget`, `rotateWithCamera`, `arc`, `retreat`, and target/source
  interpolation.
- [ ] Implement roBrowser `wav` behavior exactly: play only declared `wav`
  entries, support delay/randomization, and do not infer sound names from STRs.
- [ ] Add persistent/attached lifecycle handling: effect follows actor while
  alive, detaches when needed, and is removed on vanish/death/state clear.

### 3. Universal Feedback Effects

Goal: cover things every player sees regardless of class.

- [ ] Complete hit effects: all `EF_HIT*`, elemental hit effects, poison hit,
  miss, critical, damage message effects, and blocked/zero-damage feedback.
- [ ] Complete level-up and class-change family: base, job, Super Novice,
  Taekwon, class-change, homunculus/job variants.
- [ ] Complete item/consumable effects through roBrowser `ItemEffect.js`:
  potions, food, speed potions, fly/butterfly wings, status consumables.
- [ ] Complete warp/map-transition effects: entry, exit, old/new teleport,
  ready portal, portal, map-entry effects.
- [ ] Complete refine/pharmacy/common crafting effects.

### 4. First-Job And Basic Skill Coverage

Goal: make early gameplay feel correct before broad class sweeps.

- [x] Swordman first-job routing: Bash, Provoke, Magnum Break, Endure.
- [x] Mage first-job routing for supported client effects: Napalm Beat, Soul
  Strike SFX, Cold Bolt hit SFX, Frost Diver hit, Stone Curse, Fire Ball SFX,
  Fire Wall hit, Fire Bolt hit, Lightning Bolt, Thunder Storm.
- [x] Acolyte first-job routing for supported client effects: Ruwach hit, Heal
  rings/SFX, Increase/Decrease Agi SFX, Aqua SFX, Signum, Angelus, Blessing
  SFX, Cure.
- [x] Thief first-job routing for supported client effects: Steal SFX, Envenom
  SFX/hit routing, Detoxify SFX.
- [x] Archer first-job routing for supported client effects: Improve
  Concentration, Double Strafe start/hit, Arrow Shower hit.
- [x] Merchant first-job routing for supported client effects: Mammonite.
- [ ] Missing first-job visuals blocked by renderer primitives: Sight state loop,
  Safety Wall and Pneuma ground units, Fire Wall ground unit, Hiding state loop,
  Increase/Decrease Agi remaining sprite-backed or mixed particles, accurate
  Mage projectile/before-hit trajectories, Archer projectile trajectories, and
  cold-hit 2D shard visuals.

### 5. Ground Skill Units

Goal: implement persistent world effects that affect navigation/combat.

- [ ] Drive all ground units from roBrowser `SkillUnit.js`.
- [ ] Prioritize Safety Wall, Fire Wall, Pneuma, Sanctuary, Magnus, Quagmire,
  traps, Bard/Dancer songs, Volcano/Deluge/Violent Gale, Land Protector,
  Spider Web, Basilica, Suiton, Epiclesis.
- [ ] Add duration, repeat, ownership, and remove/update behavior from server
  unit packets.
- [ ] Verify depth and fog behavior against terrain, RSM models, sprites,
  water, and names/lifebars.

### 6. Status And Attached Loop Effects

Goal: make actor state visually legible.

- [ ] Poison, curse, blind, silence, stun, sleep, confusion.
- [ ] Frozen and stone curse, including explosion/break sounds.
- [ ] Hiding, cloaking, invisible state interactions.
- [ ] Buff/debuff loops that attach to actors and are removed by status updates.
- [ ] Spirit spheres, elemental spheres, aura-like loops, and similar persistent
  actor decorations.

### 7. Second Job And Transcendent Skill Sweep

Goal: cover classic pre-renewal content broadly.

- [ ] Knight/Crusader, Wizard/Sage, Priest/Monk, Assassin/Rogue,
  Hunter/Bard/Dancer, Blacksmith/Alchemist.
- [ ] Add test maps/NPC shortcuts or atcommand helpers to exercise each family.
- [ ] Compare each family against roBrowser first, open-midgard second.

### 8. Environment And Map Effects

Goal: make maps feel like RO, not just skill combat.

- [ ] Weather: rain, snow, sakura, thunder and ambient weather sounds.
- [ ] RSW/map effects: smoke, torch/fire, firefly, portal-like map effects.
- [ ] Ensure effects are lit/fogged consistently with the world when they should
  be, and exempt from fog/UI depth when they should not.

### 9. Renewal, Third Job, Expanded, And Modern Effects

Goal: use the now-generic system to increase coverage quickly.

- [ ] Ninja, Gunslinger, Taekwon, Star Gladiator, Soul Linker.
- [ ] Third jobs and renewal skill families.
- [ ] Fourth/modern classes present in newer roBrowser tables.
- [ ] Modern event/UI effect entries such as enchant/refine interfaces where
  they matter to gameplay.

### 10. Coverage And Regression Tooling

Goal: keep effect work measurable.

- [ ] Generate `effects_coverage.json` or similar during tests, listing
  roBrowser ID, name, table presence, Goro support level, and notes.
- [ ] Add support levels: `none`, `sound-only`, `placeholder`, `partial`,
  `close`, `parity`.
- [ ] Add a debug command or CLI flag to spawn an effect ID at the player for
  manual visual comparison.
- [ ] Add screenshot/video capture scripts for effect families.
- [ ] Update this document's percentage whenever generated coverage changes.
