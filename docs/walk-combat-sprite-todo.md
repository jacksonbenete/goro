# Walk, Combat, and Sprite TODO

Short-term goals for making movement, combat actions, and sprite playback closer
to the 2008 RO client and robr.

- Add a central `setActorAction`, like robr's `setAction`, that owns action
  state changes: action, frame, speed, repeat, play, delay, next action, and
  walk-route reset.
- Expand `skillActionSpec` toward robr's `DB/Skills/SkillAction.js` shape so
  skill stance rules can live in data instead of scattered combat conditionals.
- Teach actor animations to honor robr-style fixed frame, length, play=false,
  delayed action, and next-action metadata.
- Resume walking after hurt when the actor still has a route and a focus target,
  matching robr's `resumeWalk` behavior.
- Revisit weapon attack timing using job, sex, weapon, and attack motion data
  instead of relying only on packet speed and broad action-family mapping.
- Add missing body/action states that affect sprites, such as freeze, stone,
  blade stop/root, and other status-driven stances.
- Keep sprite composition simple, but add action-specific layer behavior only
  where the original client needs it for visible correctness.
