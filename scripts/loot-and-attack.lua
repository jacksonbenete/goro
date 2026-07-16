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

function tick()
	local hp, max_hp = goro.hp()
	if max_hp > 0 and hp / max_hp < 0.25 then
		return
	end

	local item = nearest(goro.items())
	if item ~= nil then
		goro.loot(item.id)
		return
	end

	local enemy = nearest(goro.enemies())
	if enemy ~= nil then
		goro.attack(enemy.id)
	end
end
