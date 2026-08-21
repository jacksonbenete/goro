package game

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	lua "github.com/yuin/gopher-lua"
)

func registerLuaKeyboardAPI(state *lua.LState, api *lua.LTable, ctx client.Context, bot *luaBot) {
	available := func() bool {
		return bot != nil && bot.keyboardAvailable && ctx.Input != nil
	}
	keyboard := state.NewTable()
	state.SetFuncs(keyboard, map[string]lua.LGFunction{
		"available": func(L *lua.LState) int {
			L.Push(lua.LBool(available()))
			return 1
		},
		"is_down": func(L *lua.LState) int {
			code, ok := input.KeyCodeFromName(L.CheckString(1))
			L.Push(lua.LBool(ok && available() && ctx.Input.KeyCodeDown(code)))
			return 1
		},
		"was_pressed": func(L *lua.LState) int {
			code, ok := input.KeyCodeFromName(L.CheckString(1))
			L.Push(lua.LBool(ok && available() && ctx.Input.KeyCodeJustPressed(code)))
			return 1
		},
		"was_released": func(L *lua.LState) int {
			code, ok := input.KeyCodeFromName(L.CheckString(1))
			L.Push(lua.LBool(ok && available() && ctx.Input.KeyCodeJustReleased(code)))
			return 1
		},
		"consume_press": func(L *lua.LState) int {
			code, ok := input.KeyCodeFromName(L.CheckString(1))
			L.Push(lua.LBool(ok && available() && ctx.Input.ConsumeKeyCodePress(code)))
			return 1
		},
		"text": func(L *lua.LState) int {
			if !available() {
				L.Push(lua.LString(""))
			} else {
				L.Push(lua.LString(ctx.Input.TextInput()))
			}
			return 1
		},
	})
	api.RawSetString("keyboard", keyboard)
}
