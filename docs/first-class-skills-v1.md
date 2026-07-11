# First-Class Skills V1 Checklist

Scope: pre-renewal 2008-ish first classes, Novice, and platinum/quest skills.

Sources of truth:

- roBrowser `src/DB/Skills/SkillEffect.js`
- roBrowser `src/DB/Skills/SkillAction.js`
- roBrowser `src/DB/Skills/SkillUnit.js`
- rAthena `db/pre-re/skill_tree.yml`
- rAthena `db/pre-re/skill_db.yml`

Checklist legend:

- `[x]` means the client-side behavior is implemented and checked against
  roBrowser where roBrowser declares a client effect/action/unit.
- `[ ]` means the skill still needs client work or a live verification pass.
- `Server-only` means the skill is passive or server-authoritative and should
  not have a local visual hack in the client.

Audit dimensions:

- `Behavior`: target mode, packets, UI, inventory/cart/shop side effects, and
  status routing.
- `Action`: roBrowser `SkillAction.js` stance/animation rule.
- `Effect routing`: roBrowser `SkillEffect.js` and `SkillUnit.js` fields are
  represented by `db.SkillEffects`, skill-unit mappings, or explicit
  state/system code.
- `Visual/SFX/timing`: the actual rendered primitive sequence, sounds, cast
  aura/bar, duplicate timing, offsets, alpha/blend, and cleanup have been
  compared live or by close source inspection.

Current audit state:

- Behavior coverage is broad for first-class skills, but a few systems remain.
- Static effect routing has been checked against roBrowser for the full
  first-class/platinum set.
- Visual/SFX/timing parity is being checked job by job. Treat rows as requiring
  live comparison unless explicitly called out below.

## Static roBrowser Effect Routing Audit

These are the first-class/platinum skills that roBrowser gives a non-empty
`SkillEffect.js` entry or a relevant `SkillUnit.js` entry.

| Skill | roBrowser fields | Goro routing | Notes |
| --- | --- | --- | --- |
| `SM_BASH` | `beginCastEffectId=16`, `hitEffectId=1` | Matched | SFX comes from effect IDs. |
| `SM_PROVOKE` | `successEffectId=67` | Matched | No-damage success routing. |
| `SM_MAGNUM` | `effectIdOnCaster=17`, `effectId=quake_magnum` | Matched | Includes camera shake synthetic ID. |
| `SM_ENDURE` | `effectId=11` | Matched | Ready-fight action is also imported. |
| `MG_NAPALMBEAT` | `hitEffectId=1` | Matched | Visual parity still needs live pass. |
| `MG_SAFETYWALL` | Unit `UNT_SAFETYWALL=EF_GLASSWALL2` | Matched by skill unit | Not direct `SkillEffect.js`. |
| `MG_SOULSTRIKE` | `beforeHitEffectId=15`, `hitEffectId=1` | Matched | Uses imported roBrowser-style primitives. |
| `MG_COLDBOLT` | `beforeHitEffectId=ef_coldbolt`, `hitEffectId=51` | Matched | Synthetic string-key ID. |
| `MG_FROSTDIVER` | `effectId=27`, `hitEffectId=28` | Partial visual parity | Hit effect is matched; travelling `effect/ice` FUNC primitive still needs a clean implementation. |
| `MG_STONECURSE` | `effectId=23` | Matched | |
| `MG_FIREBALL` | `beforeHitEffectId=24`, `hitEffectId=49` | Matched | |
| `MG_FIREWALL` | `groundEffectId=25`, `hitEffectId=49` | Matched | Also unit-routed for wall cells. |
| `MG_FIREBOLT` | `beforeHitEffectId=ef_firebolt`, `hitEffectId=49` | Matched | Synthetic string-key ID. |
| `MG_LIGHTNINGBOLT` | `effectId=29`, `hitEffectId=52` | Matched | |
| `MG_THUNDERSTORM` | `effectId=30`, `hitEffectId=52` | Matched | Ground target/cast timing still needs live pass. |
| `MG_ENERGYCOAT` | `effectId=169` | Matched | Platinum skill. |
| `AL_RUWACH` | `hitEffectId=1`; state effect separate | Matched | Persistent state effect is explicit code. |
| `AL_PNEUMA` | `groundEffectId=141` | Matched | Unit-routed. |
| `AL_HEAL` | `effectId=312`, `hitEffectId=320` | Matched | Includes enemy-target offensive heal effect. |
| `AL_INCAGI` | `effectId=37` | Matched | |
| `AL_DECAGI` | `effectId=38` | Matched | |
| `AL_HOLYWATER` | `effectId=39` | Matched | |
| `AL_CRUCIS` | `effectId=40` | Matched | |
| `AL_ANGELUS` | `effectId=41` | Matched | Cast/cylinder height has prior focused fix. |
| `AL_BLESSING` | `effectId=42` | Matched | |
| `AL_CURE` | `effectId=66` | Matched | |
| `AL_HOLYLIGHT` | `effectId=152` | Matched | Platinum skill. |
| `AC_CONCENTRATION` | `effectId=153` | Matched | |
| `AC_DOUBLE` | `beginCastEffectId=16`, `beforeHitEffectId=ef_arrow_projectile`, `hitEffectId=1` | Matched | |
| `AC_SHOWER` | `effectId=ef_arrow_shower_projectile`, `hitEffectId=1` | Matched | |
| `AC_CHARGEARROW` | `hideCastAura`, `beforeHitEffectId=ef_arrow_projectile` | Matched | Fixed in this audit pass. |
| `TF_STEAL` | `successEffectId=18` | Matched | |
| `TF_HIDING` | handled by state | Matched by status code | Direct roBrowser row is intentionally empty. |
| `TF_POISON` | `hitEffectId=20` | Matched | |
| `TF_DETOXIFY` | `effectId=21` | Matched | |
| `TF_SPRINKLESAND` | `effectId=310` | Matched | Platinum skill. |
| `TF_PICKSTONE` | `hideCastAura` | Matched | Platinum skill; no visual effect. |
| `TF_THROWSTONE` | `beforeHitEffectId=308` | Matched | Platinum skill. |
| `MC_MAMMONITE` | `effectId=10` | Matched | |
| `MC_CARTREVOLUTION` | `beginCastEffectId=170`, `hitEffectId=170` | Matched | Platinum skill. |
| `MC_LOUD` | `effectId=311` | Matched | Platinum skill. |
| `NV_FIRSTAID` | `effectId=309` | Matched | Platinum/basic quest skill. |

roBrowser declares `{}` for `NV_TRICKDEAD`, `SM_AUTOBERSERK`,
`AC_MAKINGARROW`, `TF_BACKSLIDING`, `MC_CHANGECART`, `MC_IDENTIFY`,
`MC_VENDING`, and `MC_CARTDECORATE`, so missing work there is behavior/status/UI
rather than direct `SkillEffect.js` routing.

## Deep Visual/SFX/Timing Audit

### Novice And Swordman

Source pass completed against roBrowser `SkillEffect.js`, `SkillAction.js`, and
the relevant `EffectTable.js` entries.

| Skill | Visual/SFX/timing status | Notes |
| --- | --- | --- |
| `NV_BASIC` | Matched action behavior | roBrowser suppresses default skill action; goro now does too. |
| `NV_FIRSTAID` | Matched by source | Effect `309`: 2D `effect/pikapika2.bmp`, additive blend, `posz=2`, 1000ms, `_heal_effect`. |
| `NV_TRICKDEAD` | Matched behavior | roBrowser declares no skill effect/action; goro waits for server status and holds the death pose. |
| `SM_BASH` | Matched by source | Begin effect `16` and hit effect `1`; SFX comes from those effect IDs. |
| `SM_PROVOKE` | Matched by source | Success effect `67`: STR `provoke` plus `effect/swordman_provoke`. |
| `SM_MAGNUM` | Matched by source | Caster effect `17` matches roBrowser cylinders/SFX; target `quake_magnum` is modeled as camera shake. |
| `SM_ENDURE` | Matched by source | Effect `11` and ready-fight action match roBrowser. |
| `SM_MOVINGRECOVERY` | Server-only | No roBrowser skill effect/action. |
| `SM_FATALBLOW` | Server-only | Bash extension; no roBrowser skill effect/action. |
| `SM_AUTOBERSERK` | Status-driven | roBrowser declares no skill effect. rAthena toggle is `SC_AUTOBERSERK`; active combat state uses `SC_BERSERK`/`Opt3Berserk`, now propagated through imported Opt3 status state. |

### Mage

Source pass completed against roBrowser `SkillEffect.js`, `SkillAction.js`,
`SkillUnit.js`, and the relevant numeric/string-key `EffectTable.js` entries.
roBrowser has no Mage-specific `SkillAction.js` overrides in this range, so
Mage active skills use the default skill action unless the cast pipeline
overrides stance/timing from server cast ACKs.

| Skill | Visual/SFX/timing status | Notes |
| --- | --- | --- |
| `MG_SRECOVERY` | Server-only | Passive SP recovery; no roBrowser skill effect/action. |
| `MG_SIGHT` | Matched by source | Effect `22`: roBrowser shadow-texture orbit, `sight` sprite orbit, and `effect/ef_sight` SFX. Removed the old local cylinder placeholder. |
| `MG_NAPALMBEAT` | Matched by source | roBrowser routes hit effect `1`; effect `32` is SFX-only/commented FUNC in the table. |
| `MG_SAFETYWALL` | Matched by source | Unit `UNT_SAFETYWALL` maps to effect `315` (`EF_GLASSWALL2`): STR `safetywall` plus three staggered `alpha_down` cylinders. |
| `MG_SOULSTRIKE` | Matched by source | Before-hit effect `15` plus hit effect `1`; projectile particles use roBrowser timing and source-to-target movement. |
| `MG_COLDBOLT` | Matched by source | String-key `ef_coldbolt`: falling `icearrow.tga`, random `ef_icearrow1..3`, delayed blue ring, then hit effect `51`. |
| `MG_FROSTDIVER` | Partial | Hit effect `28` matches STR `freeze` plus `ef_frostdiver2`. Travelling effect `27` points at old roBrowser `effect/ice` FUNC behavior; goro keeps a visible approximation but removed the guessed travel SFX. |
| `MG_STONECURSE` | Matched by source | Effect `23`; server owns success/failure/status. |
| `MG_FIREBALL` | Matched by source | Before-hit effect `24` and hit effect `49`; projectile angle/timing follows roBrowser table. |
| `MG_FIREWALL` | Matched by source | Ground effect `25` and hit effect `49`; unit cells are server-driven. |
| `MG_FIREBOLT` | Matched by source | String-key `ef_firebolt`: falling fire arrow frame list, random `ef_firearrow1..3`, then hit effect `49`. |
| `MG_LIGHTNINGBOLT` | Matched by source | Effect `29`: `lightning` STR, random `windhit1..3`, random `ef_lightningbolt1..3`, then hit effect `52`. |
| `MG_THUNDERSTORM` | Matched by source | Effect `30` and hit effect `52`; ground target/cast delivery is server-driven. |
| `MG_ENERGYCOAT` | Matched by source | Platinum effect `169`; status routing uses imported status/Opt3 metadata. |

### Archer

Source pass completed against roBrowser `SkillEffect.js`, `SkillAction.js`,
and string-key `EffectTable.js` projectile entries.

| Skill | Visual/SFX/timing status | Notes |
| --- | --- | --- |
| `AC_OWL` | Server-only | Passive DEX bonus; no roBrowser skill effect/action. |
| `AC_VULTURE` | Server-only | Passive range bonus; no roBrowser skill effect/action. |
| `AC_CONCENTRATION` | Matched by source | Effect `153`: STR `concentration` plus `effect/ac_concentration`; server owns status/stat updates. |
| `AC_DOUBLE` | Matched by source | roBrowser action `ATTACK3`, begin effect `16`, projectile `ef_arrow_projectile`, and hit effect `1`. |
| `AC_SHOWER` | Matched by source | Projectile `ef_arrow_shower_projectile`, hit effect `1`, and roBrowser action timing `ATTACK` speed 50ms followed by `READYFIGHT`. |
| `AC_MAKINGARROW` | Missing behavior | roBrowser declares no visual effect; goro still needs the Arrow Crafting item-selection flow. |
| `AC_CHARGEARROW` | Matched by source | roBrowser action `ATTACK`, hidden cast aura, and before-hit `ef_arrow_projectile`. |

## Novice

- [x] `NV_BASIC` - Server-only basic interface/trade progression.
- [x] `NV_FIRSTAID` - Recovery skill, server-authoritative HP change, client
  effect/SFX/timing source parity covered.
- [x] `NV_TRICKDEAD` - Status-driven dead pose; no fake skill action before the
  server status update.

## Swordman

- [x] `SM_SWORD` - Server-only passive.
- [x] `SM_TWOHAND` - Server-only passive.
- [x] `SM_RECOVERY` - Server-only passive/recovery tick.
- [x] `SM_BASH` - roBrowser attack action/effect.
- [x] `SM_PROVOKE` - roBrowser no-damage effect/action routing.
- [x] `SM_MAGNUM` - roBrowser attack action/effect.
- [x] `SM_ENDURE` - roBrowser ready-fight action/effect/status routing.
- [x] `SM_MOVINGRECOVERY` - Server-only platinum passive.
- [x] `SM_FATALBLOW` - Server-only platinum Bash extension.
- [x] `SM_AUTOBERSERK` - Platinum status toggle; direct skill effect is empty in
  roBrowser and active Berserk Opt3 state is now imported/status-driven.

## Mage

- [x] `MG_SRECOVERY` - Server-only passive.
- [x] `MG_SIGHT` - Actor-attached Sight state/effect.
- [x] `MG_NAPALMBEAT` - roBrowser effect routing.
- [x] `MG_SAFETYWALL` - roBrowser ground skill unit effect.
- [x] `MG_SOULSTRIKE` - roBrowser effect routing.
- [x] `MG_COLDBOLT` - roBrowser bolt effect routing.
- [ ] `MG_FROSTDIVER` - Hit effect matches roBrowser; travelling `effect/ice`
  FUNC primitive needs a clean implementation.
- [x] `MG_STONECURSE` - roBrowser effect/cast routing.
- [x] `MG_FIREBALL` - roBrowser source-to-target effect routing.
- [x] `MG_FIREWALL` - roBrowser ground skill unit effect.
- [x] `MG_FIREBOLT` - roBrowser bolt effect routing.
- [x] `MG_LIGHTNINGBOLT` - roBrowser bolt effect routing.
- [x] `MG_THUNDERSTORM` - roBrowser ground-target cast and hit effects.
- [x] `MG_ENERGYCOAT` - Platinum status/effect routing.

## Archer

- [x] `AC_OWL` - Server-only passive.
- [x] `AC_VULTURE` - Server-only passive/range.
- [x] `AC_CONCENTRATION` - roBrowser effect/status routing.
- [x] `AC_DOUBLE` - roBrowser attack action/effect and arrow projectile path.
- [x] `AC_SHOWER` - roBrowser attack action/effect.
- [ ] `AC_MAKINGARROW` - Platinum Arrow Crafting item-selection flow is not
  implemented yet.
- [x] `AC_CHARGEARROW` - Platinum attack action, hidden cast aura, and arrow
  projectile effect routing.

## Acolyte

- [x] `AL_DP` - Server-only passive.
- [x] `AL_DEMONBANE` - Server-only passive.
- [x] `AL_RUWACH` - roBrowser effect table routing.
- [x] `AL_PNEUMA` - roBrowser ground skill unit effect.
- [x] `AL_TELEPORT` - Level-specific modal/instant behavior and warp effect.
- [x] `AL_WARP` - Destination modal, portal unit effect, and portal lifecycle.
- [x] `AL_HEAL` - Targeting rules, cast action, heal effect, and enemy-target
  support.
- [x] `AL_INCAGI` - Cast/action/effect/status routing.
- [x] `AL_DECAGI` - roBrowser effect/status routing.
- [x] `AL_HOLYWATER` - roBrowser effect routing; server owns item requirement.
- [x] `AL_CRUCIS` - roBrowser effect routing.
- [x] `AL_ANGELUS` - Cast/action/effect/status routing.
- [x] `AL_BLESSING` - Targeting rules and roBrowser effect/status routing.
- [x] `AL_CURE` - roBrowser effect routing.
- [x] `AL_HOLYLIGHT` - Platinum cast bar/action/effect routing.

## Merchant

- [x] `MC_INCCARRY` - Server-only passive.
- [x] `MC_DISCOUNT` - Server-only shop price modifier.
- [x] `MC_OVERCHARGE` - Server-only sell price modifier.
- [x] `MC_PUSHCART` - Cart status, cart sprite, cart storage, and cart item
  packets.
- [x] `MC_IDENTIFY` - Magnifier/item identify UI and packet flow.
- [x] `MC_VENDING` - Vending setup, vending board, buyer shop flow, and shop
  bubble.
- [x] `MC_MAMMONITE` - roBrowser attack action/effect.
- [x] `MC_CARTREVOLUTION` - Platinum attack action/effect.
- [x] `MC_CHANGECART` - Platinum cart-change packet and cart sprite update.
- [x] `MC_LOUD` - Platinum Crazy Uproar effect/status routing.
- [ ] `MC_CARTDECORATE` - Not implemented; confirm whether this belongs in the
  targeted 2008/pre-re v1 set before adding UI/packet support.

## Thief

- [x] `TF_DOUBLE` - Server-only passive.
- [x] `TF_MISS` - Server-only passive.
- [x] `TF_STEAL` - roBrowser effect and server failure/success handling.
- [x] `TF_HIDING` - Status-driven hidden state and transition effects.
- [x] `TF_POISON` - roBrowser attack action/effect.
- [x] `TF_DETOXIFY` - roBrowser effect routing.
- [x] `TF_SPRINKLESAND` - Platinum attack action/effect.
- [x] `TF_BACKSLIDING` - Platinum no-action skill plus authoritative jump packet.
- [x] `TF_PICKSTONE` - Platinum pickup action routing; server owns item creation.
- [x] `TF_THROWSTONE` - Platinum attack action/effect.

## Remaining V1 Work

- [ ] Implement `AC_MAKINGARROW`: parse the server item list, show the selection
  UI, send the selected item, and verify inventory/equipped arrow updates.
- [ ] Decide whether `MC_CARTDECORATE` belongs in the 2008 v1 target. If yes,
  implement the packet/UI and cart appearance handling.
- [ ] Implement the old roBrowser `effect/ice` FUNC primitive used by Frost
  Diver travelling effect `27`, then remove the remaining local approximation.
- [ ] Perform a focused visual/SFX/timing comparison for every row in the
  static routing audit table above that is not already covered in the deep audit
  section.
- [ ] Run one manual first-class regression route per class on rAthena:
  targeting, cast bars, action stance, effect timing, status icon, sound,
  inventory/cart/shop side effects, and map transition cleanup.
- [ ] Add targeted tests for every remaining unchecked skill after implementing
  or explicitly dropping it from the v1 scope.
