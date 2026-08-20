package game

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/network"
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
	local inventory = goro.inventory()
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
		inventory = #inventory,
		inventory_index = inventory[1].index,
		inventory_item_id = inventory[1].item_id,
		inventory_amount = inventory[1].amount,
		inventory_identified = inventory[1].identified,
		inventory_usable = inventory[1].usable,
		second_inventory_usable = inventory[2].usable,
	}
end
`), 0o644); err != nil {
		t.Fatal(err)
	}

	sess := session.New()
	sess.CharID = 2000000
	sess.Vitals = session.Vitals{HP: 42, MaxHP: 100, SP: 7, MaxSP: 20}
	sess.Inventory.Items = []session.InventoryItem{
		{Index: 9, ItemID: 909, Type: db.ItemTypeEtc, Amount: 2, Identified: true},
		{Index: 4, ItemID: 501, Type: db.ItemTypeHealing, Amount: 5, Identified: true},
	}
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
	assertLuaNumber(t, seen, "inventory", 2)
	assertLuaNumber(t, seen, "inventory_index", 4)
	assertLuaNumber(t, seen, "inventory_item_id", 501)
	assertLuaNumber(t, seen, "inventory_amount", 5)
	assertLuaBool(t, seen, "inventory_identified", true)
	assertLuaBool(t, seen, "inventory_usable", true)
	assertLuaBool(t, seen, "second_inventory_usable", false)
}

func TestLuaBotCanUseInventoryItem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bot.lua")
	if err := os.WriteFile(path, []byte(`
function tick()
	local inventory = goro.inventory()
	used = goro.use_item(inventory[1].index)
end
`), 0o644); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan net.Conn, 1)
	acceptErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			acceptErr <- err
			return
		}
		accepted <- conn
	}()

	networkClient := network.NewClient(20080910, false)
	port := listener.Addr().(*net.TCPAddr).Port
	if err := networkClient.Connect(context.Background(), "127.0.0.1", port); err != nil {
		t.Fatal(err)
	}
	defer networkClient.Close()

	var serverConn net.Conn
	select {
	case serverConn = <-accepted:
		defer serverConn.Close()
	case err := <-acceptErr:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatal("timed out accepting bot test connection")
	}

	sess := session.New()
	sess.AccountID = 0x11223344
	sess.Inventory.Items = []session.InventoryItem{{
		Index:      7,
		ItemID:     501,
		Type:       db.ItemTypeHealing,
		Amount:     3,
		Identified: true,
	}}
	bot, err := newLuaBot(client.Context{Session: sess, Network: networkClient}, &WorldMode{}, path)
	if err != nil {
		t.Fatal(err)
	}
	defer bot.close()
	if err := bot.tick(); err != nil {
		t.Fatal(err)
	}
	if used, _ := bot.state.GetGlobal("used").(lua.LBool); !bool(used) {
		t.Fatal("goro.use_item returned false")
	}

	want := network.BuildUseInventoryItemPacketForClientDate(7, sess.AccountID, 20080910)
	got := make([]byte, len(want))
	if err := serverConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(serverConn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("use item packet = %x, want %x", got, want)
	}
}

func TestScriptUseItemRejectsInvalidInventoryEntries(t *testing.T) {
	sess := session.New()
	sess.Inventory.Items = []session.InventoryItem{{Index: 7, ItemID: 909, Type: db.ItemTypeEtc, Amount: 1}}
	ctx := client.Context{Session: sess}

	for _, index := range []int{-1, 0, 7, 8, 1 << 16} {
		if scriptUseItem(ctx, index) {
			t.Fatalf("scriptUseItem(%d) = true, want false", index)
		}
	}
}

func TestLuaBotCanRequestTargetSkillChase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bot.lua")
	if err := os.WriteFile(path, []byte(`
function tick()
	ok = goro.skill(300, 46)
end
`), 0o644); err != nil {
		t.Fatal(err)
	}

	sess := session.New()
	sess.CharID = 2000000
	sess.Skills.List = []session.Skill{{
		ID:    db.SkillACDouble,
		Type:  skillTargetEnemy,
		Level: 10,
		Range: 9,
		Name:  "Double Strafe",
	}}
	world := worldstate.New()
	world.GAT = flatWalkableGAT(64, 64)
	world.Player = worldstate.Actor{ID: sess.CharID, X: 10, Y: 20}
	world.Actors[300] = worldstate.Actor{ID: 300, Name: "Poring", X: 30, Y: 20, ObjectType: actorObjectTypeMob, HasObjectType: true}

	mode := &WorldMode{}
	bot, err := newLuaBot(client.Context{Session: sess, World: world}, mode, path)
	if err != nil {
		t.Fatal(err)
	}
	defer bot.close()
	if err := bot.tick(); err != nil {
		t.Fatal(err)
	}

	ok, _ := bot.state.GetGlobal("ok").(lua.LBool)
	if !bool(ok) {
		t.Fatal("goro.skill returned false")
	}
	if mode.pendingSkill.targetID != 300 || mode.pendingSkill.skill.ID != db.SkillACDouble {
		t.Fatalf("pending skill = %+v, want AC_DOUBLE target 300", mode.pendingSkill)
	}
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

func assertLuaBool(t *testing.T, table *lua.LTable, key string, want bool) {
	t.Helper()
	got, ok := table.RawGetString(key).(lua.LBool)
	if !ok {
		t.Fatalf("%s = %T, want bool", key, table.RawGetString(key))
	}
	if bool(got) != want {
		t.Fatalf("%s = %v, want %v", key, got, want)
	}
}
