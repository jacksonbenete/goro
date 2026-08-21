-- Physical WASD positions: these keys are ZQSD on an AZERTY keyboard.
local controls = { "KeyW", "KeyA", "KeyS", "KeyD" }
local horizon = 8
local refill_distance = 3
local action_radius = 8
local active_dx = 0
local active_dy = 0
local target_x = nil
local target_y = nil
local loot_target_id = nil
local attack_target_id = nil
local fight_down = false
local loot_down = false
local attack_retry_seconds = 1.2
local loot_retry_seconds = 1.0
local last_attack_at = -math.huge
local last_loot_at = -math.huge

local function clear_target()
	active_dx = 0
	active_dy = 0
	target_x = nil
	target_y = nil
end

local function request_walk(player, dx, dy, max_distance)
	for distance = max_distance, 1, -1 do
		local x = player.x + dx * distance
		local y = player.y + dy * distance
		if goro.walk(x, y) then
			active_dx = dx
			active_dy = dy
			target_x = x
			target_y = y
			return true
		end
	end
	return false
end

local function select_nearest(entries, current_id)
	local nearest = nil
	for _, entry in ipairs(entries) do
		if entry.id == current_id then
			return entry
		end
		if entry.distance <= action_radius
			and (nearest == nil or entry.distance < nearest.distance) then
			nearest = entry
		end
	end
	return nearest
end

local function schedule_attack(now)
	local target = select_nearest(goro.enemies(), attack_target_id)
	if target == nil then
		attack_target_id = nil
		return false
	end
	if target.id ~= attack_target_id or now - last_attack_at >= attack_retry_seconds then
		if goro.attack(target.id) then
			attack_target_id = target.id
			last_attack_at = now
		end
	end
	return attack_target_id ~= nil
end

local function schedule_loot(now)
	local target = select_nearest(goro.items(), loot_target_id)
	if target == nil then
		loot_target_id = nil
		return false
	end
	if target.id ~= loot_target_id or now - last_loot_at >= loot_retry_seconds then
		if goro.loot(target.id) then
			loot_target_id = target.id
			last_loot_at = now
		end
	end
	return loot_target_id ~= nil
end

function tick()
	local now = os.clock()
	if fight_down and schedule_attack(now) then
		loot_target_id = nil
		clear_target()
		return
	end
	if loot_down and schedule_loot(now) then
		clear_target()
	end
end

function input()
	for _, code in ipairs(controls) do
		goro.keyboard.consume_press(code)
	end
	goro.keyboard.consume_press("Space")
	goro.keyboard.consume_press("KeyF")

	fight_down = goro.keyboard.is_down("KeyF")
	loot_down = goro.keyboard.is_down("Space")
	if not fight_down then
		attack_target_id = nil
	end
	if not loot_down then
		loot_target_id = nil
	end
	if attack_target_id ~= nil or loot_target_id ~= nil then
		clear_target()
		return
	end

	local dx = 0
	local dy = 0
	if goro.keyboard.is_down("KeyW") then dy = dy + 1 end
	if goro.keyboard.is_down("KeyA") then dx = dx - 1 end
	if goro.keyboard.is_down("KeyS") then dy = dy - 1 end
	if goro.keyboard.is_down("KeyD") then dx = dx + 1 end

	if dx == 0 and dy == 0 then
		if active_dx ~= 0 or active_dy ~= 0 then
			if goro.stop() then
				clear_target()
			end
		end
		return
	end

	local player = goro.player()
	if dx ~= active_dx or dy ~= active_dy then
		request_walk(player, dx, dy, horizon)
		return
	end

	if target_x ~= nil and target_y ~= nil then
		local remaining = math.max(math.abs(target_x - player.x), math.abs(target_y - player.y))
		if remaining <= refill_distance then
			request_walk(player, dx, dy, horizon)
		end
	end
end
