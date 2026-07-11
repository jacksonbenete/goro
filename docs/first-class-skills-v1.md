# First-Class Skills V1 Checklist

Scope: pre-renewal 2008-ish first classes, Novice, and platinum/quest skills.

Sources of truth:

- roBrowser `src/DB/Skills/SkillEffect.js`
- roBrowser `src/DB/Skills/SkillAction.js`
- roBrowser `src/DB/Skills/SkillUnit.js`
- rAthena `db/pre-re/skill_tree.yml`
- rAthena `db/pre-re/skill_db.yml`

Legend:

- `[x]` means the client-side behavior is implemented and checked against
  roBrowser where roBrowser declares a client effect/action/unit.
- `[ ]` means the skill still needs client work or a live verification pass.
- `Server-only` means the skill is passive or server-authoritative and should
  not have a local visual hack in the client.

## Novice

- [x] `NV_BASIC` - Server-only basic interface/trade progression.
- [x] `NV_FIRSTAID` - Recovery skill, server-authoritative HP change, client
  effect path covered.
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
- [ ] `SM_AUTOBERSERK` - Platinum status toggle still needs a live pass against
  server status packets and any actor decoration expected by the 2008 client.

## Mage

- [x] `MG_SRECOVERY` - Server-only passive.
- [x] `MG_SIGHT` - Actor-attached Sight state/effect.
- [x] `MG_NAPALMBEAT` - roBrowser effect routing.
- [x] `MG_SAFETYWALL` - roBrowser ground skill unit effect.
- [x] `MG_SOULSTRIKE` - roBrowser effect routing.
- [x] `MG_COLDBOLT` - roBrowser bolt effect routing.
- [x] `MG_FROSTDIVER` - roBrowser effect routing.
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
- [x] `AC_CHARGEARROW` - Platinum attack action/effect.

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
- [ ] Live-test `SM_AUTOBERSERK` with a real server status transition.
- [ ] Run one manual first-class regression route per class on rAthena:
  targeting, cast bars, action stance, effect timing, status icon, sound,
  inventory/cart/shop side effects, and map transition cleanup.
- [ ] Add targeted tests for every remaining unchecked skill after implementing
  or explicitly dropping it from the v1 scope.
