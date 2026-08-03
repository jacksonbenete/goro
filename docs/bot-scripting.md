# Bot Scripting

Goro can run a Lua script while the player is in-game. This is intended for local experimentation and simple automation.

Run a script with:

```sh
./goro --data-dir ~/Telechargements/OldRO --username Kivutar --password ... --script scripts/loot-and-attack.lua
```

The script must define a global `tick()` function. Goro calls it roughly every 150 ms while the world mode is active.

```lua
function tick()
	-- bot logic here
end
```

## API

All functions are exposed through the global `goro` table.

### `goro.player()`

Returns the local player state.

Fields:

- `id`
- `x`
- `y`
- `hp`
- `max_hp`
- `sp`
- `max_sp`

### `goro.hp()`

Returns two values:

```lua
local hp, max_hp = goro.hp()
```

### `goro.sp()`

Returns two values:

```lua
local sp, max_sp = goro.sp()
```

### `goro.enemies()`

Returns an array of currently attackable enemies. Actors already playing their death animation are filtered out.

Each enemy has:

- `id`
- `name`
- `x`
- `y`
- `job`
- `object_type`
- `distance`

### `goro.attack(id)`

Requests a normal attack on the enemy actor with this id.

Returns `true` if the target exists and is attackable, otherwise `false`.

This uses the same path as a normal player click, including chase and range handling. Scripts should avoid calling it every tick for the same target; keep a small retry delay.

### `goro.target(id)`

Alias for `goro.attack(id)`.

### `goro.skill(id, skill)`

Requests a target skill on the enemy actor with this id. `skill` can be either a numeric skill id or a learned skill name such as `"AC_DOUBLE"`.

Returns `true` if the target and skill are usable, otherwise `false`.

This uses the same path as a skill-window or shortcut target click, including chase and range handling. Scripts should avoid calling it every tick for the same target; keep a small retry delay.

### `goro.items()`

Returns an array of visible floor items.

Each item has:

- `id`
- `item_id`
- `amount`
- `x`
- `y`
- `identified`
- `distance`

### `goro.loot(id)`

Requests pickup for the floor item with this id.

Returns `true` if the item exists, otherwise `false`.

This uses the same path as a normal player click, including walking into pickup range. Scripts should avoid calling it every tick for the same item; keep a small retry delay.

## Example

This loots the nearest item first, then attacks the nearest enemy. It stops when HP is under 25%.

```lua
local function nearest(entries)
	local best = nil
	for _, entry in ipairs(entries) do
		if best == nil or entry.distance < best.distance then
			best = entry
		end
	end
	return best
end

local current_target = nil
local last_attack_at = 0
local attack_retry_seconds = 1.2
local double_strafe_id = 46
local current_item = nil
local last_loot_at = 0
local loot_retry_seconds = 1.0

function tick()
	local hp, max_hp = goro.hp()
	if max_hp > 0 and hp / max_hp < 0.25 then
		return
	end

	local item = nearest(goro.items())
	if item ~= nil then
		local now = os.clock()
		current_target = nil
		if current_item ~= item.id or now - last_loot_at >= loot_retry_seconds then
			current_item = item.id
			last_loot_at = now
			goro.loot(item.id)
		end
		return
	end
	current_item = nil

	local enemy = nearest(goro.enemies())
	if enemy ~= nil then
		local now = os.clock()
		if current_target ~= enemy.id or now - last_attack_at >= attack_retry_seconds then
			current_target = enemy.id
			last_attack_at = now
			if not goro.skill(enemy.id, double_strafe_id) then
				goro.attack(enemy.id)
			end
		end
	else
		current_target = nil
	end
end
```

The same script is available at `scripts/loot-and-attack.lua`.
