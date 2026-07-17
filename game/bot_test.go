package game

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/session"
	worldstate "github.com/kivutar/goro/world"
	lua "github.com/yuin/gopher-lua"
)

func TestLuaBotExposesWorldState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bot.lua")
	if err := os.WriteFile(path, []byte(`
function tick()
	local hp, max_hp = goro.hp()
	local sp, max_sp = goro.sp()
	local player = goro.player()
	local enemies = goro.enemies()
	local items = goro.items()
	seen = {
		hp = hp,
		max_hp = max_hp,
		sp = sp,
		max_sp = max_sp,
		player_x = player.x,
		player_y = player.y,
		enemies = #enemies,
		enemy_id = enemies[1].id,
		items = #items,
		item_id = items[1].item_id,
	}
end
`), 0o644); err != nil {
		t.Fatal(err)
	}

	sess := session.New()
	sess.CharID = 2000000
	sess.Vitals = session.Vitals{HP: 42, MaxHP: 100, SP: 7, MaxSP: 20}
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: sess.CharID, X: 10, Y: 20}
	world.Actors[300] = worldstate.Actor{ID: 300, Name: "Poring", X: 12, Y: 21, ObjectType: actorObjectTypeMob, HasObjectType: true}
	world.Items[400] = worldstate.FloorItem{ID: 400, ItemID: 501, X: 11, Y: 20, Amount: 2, Identified: true}

	mode := &WorldMode{}
	bot, err := newLuaBot(client.Context{Session: sess, World: world}, mode, path)
	if err != nil {
		t.Fatal(err)
	}
	defer bot.close()
	if err := bot.tick(); err != nil {
		t.Fatal(err)
	}

	seen, ok := bot.state.GetGlobal("seen").(*lua.LTable)
	if !ok {
		t.Fatalf("seen = %T, want table", bot.state.GetGlobal("seen"))
	}
	assertLuaNumber(t, seen, "hp", 42)
	assertLuaNumber(t, seen, "max_hp", 100)
	assertLuaNumber(t, seen, "sp", 7)
	assertLuaNumber(t, seen, "max_sp", 20)
	assertLuaNumber(t, seen, "player_x", 10)
	assertLuaNumber(t, seen, "player_y", 20)
	assertLuaNumber(t, seen, "enemies", 1)
	assertLuaNumber(t, seen, "enemy_id", 300)
	assertLuaNumber(t, seen, "items", 1)
	assertLuaNumber(t, seen, "item_id", 501)
}

func TestLuaBotDoesNotExposeDyingEnemies(t *testing.T) {
	sess := session.New()
	sess.CharID = 2000000
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: sess.CharID, X: 10, Y: 20}
	world.Actors[300] = worldstate.Actor{ID: 300, Name: "Poring", X: 12, Y: 21, ObjectType: actorObjectTypeMob, HasObjectType: true}

	enemies := luaEnemyList(lua.NewState(), client.Context{Session: sess, World: world}, map[uint32]time.Time{
		300: time.Now().Add(time.Second),
	})
	if enemies.Len() != 0 {
		t.Fatalf("enemies len = %d, want 0", enemies.Len())
	}
}

func assertLuaNumber(t *testing.T, table *lua.LTable, key string, want float64) {
	t.Helper()
	got, ok := table.RawGetString(key).(lua.LNumber)
	if !ok {
		t.Fatalf("%s = %T, want number", key, table.RawGetString(key))
	}
	if float64(got) != want {
		t.Fatalf("%s = %v, want %v", key, got, want)
	}
}
