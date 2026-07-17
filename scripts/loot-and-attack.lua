-- Minimal goro bot script.
-- Usage:
--   ./goro --data-dir ~/Telechargements/OldRO --username Kivutar --password ... --script scripts/loot-and-attack.lua

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
			goro.attack(enemy.id)
		end
	else
		current_target = nil
	end
end
