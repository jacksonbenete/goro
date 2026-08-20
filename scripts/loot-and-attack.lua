-- Example goro bot script with combat, looting, healing, and rest behavior.
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
local attack_retry_seconds = 1.2
local last_attack_at = -math.huge
local double_strafe_id = 46
local current_item = nil
local loot_retry_seconds = 1.0
local last_loot_at = -math.huge
local potion_retry_seconds = 1.0
local last_potion_at = -math.huge
local resting = false

local healing_item_ids = {
	[501] = true, -- Red Potion
	[502] = true, -- Orange Potion
	[503] = true, -- Yellow Potion
	[504] = true, -- White Potion
	[545] = true, -- Condensed Red Potion
	[546] = true, -- Condensed Yellow Potion
	[547] = true, -- Condensed White Potion
}

local function ratio(value, maximum)
	if maximum <= 0 then
		return 1
	end
	return value / maximum
end

local function healing_item()
	for _, item in ipairs(goro.inventory()) do
		if item.usable and item.amount > 0 and healing_item_ids[item.item_id] then
			return item
		end
	end
	return nil
end

local function update_recovery(now, hp_ratio, sp_ratio)
	local potion = healing_item()
	if hp_ratio < 0.55 and potion ~= nil and now - last_potion_at >= potion_retry_seconds then
		if goro.use_item(potion.index) then
			last_potion_at = now
		end
	end

	local needs_rest = hp_ratio < 0.30 or sp_ratio < 0.15 or (hp_ratio < 0.55 and potion == nil)
	if not resting and needs_rest then
		if goro.message("/sit") then
			resting = true
			current_target = nil
			goro.message("I'm tired. Resting for a moment.")
		end
	elseif resting and hp_ratio >= 0.90 and sp_ratio >= 0.80 then
		if goro.message("/stand") then
			resting = false
			goro.message("Ready to go again.")
		end
	end
	return resting
end

function tick()
	local hp, max_hp = goro.hp()
	local sp, max_sp = goro.sp()
	local now = os.clock()
	if update_recovery(now, ratio(hp, max_hp), ratio(sp, max_sp)) then
		return
	end

	local item = nearest(goro.items())
	if item ~= nil then
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
