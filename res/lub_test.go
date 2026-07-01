package res

import "testing"

func TestExecuteLuaFunctionWithGopherLuaSupportsArithmetic(t *testing.T) {
	fn := luaFunction{
		constants: []luaValue{
			{kind: luaNumber, num: 40},
			{kind: luaNumber, num: 2},
			{kind: luaString, str: "Answer"},
		},
		code: []uint32{
			luaInstructionABx(luaOpLoadK, 0, 0),
			luaInstructionABx(luaOpLoadK, 1, 1),
			luaInstructionABC(luaOpAdd, 2, 0, 1),
			luaInstructionABx(luaOpSetGlobal, 2, 2),
			luaInstructionABC(luaOpReturn, 0, 1, 0),
		},
		maxStackSize: 3,
	}
	globals := make(map[string]luaValue)
	if err := executeLuaFunctionWithGopherLua(fn, globals, nil); err != nil {
		t.Fatal(err)
	}
	if got := globals["Answer"]; got.kind != luaNumber || got.num != 42 {
		t.Fatalf("Answer = %#v, want Lua number 42", got)
	}
}

func TestExecuteLuaFunctionWithGopherLuaBuildsTables(t *testing.T) {
	fn := luaFunction{
		constants: []luaValue{
			{kind: luaString, str: "Table"},
			{kind: luaNumber, num: 1},
			{kind: luaString, str: "hello"},
		},
		code: []uint32{
			luaInstructionABC(luaOpNewTable, 0, 0, 0),
			luaInstructionABx(luaOpLoadK, 1, 2),
			luaInstructionABC(luaOpSetTable, 0, luaBitRK|1, 1),
			luaInstructionABx(luaOpSetGlobal, 0, 0),
			luaInstructionABC(luaOpReturn, 0, 1, 0),
		},
		maxStackSize: 2,
	}
	globals := make(map[string]luaValue)
	if err := executeLuaFunctionWithGopherLua(fn, globals, nil); err != nil {
		t.Fatal(err)
	}
	table := globals["Table"]
	if table.kind != luaTable {
		t.Fatalf("Table = %#v, want Lua table", table)
	}
	if got := table.table[1]; got.kind != luaString || got.str != "hello" {
		t.Fatalf("Table[1] = %#v, want Lua string hello", got)
	}
}

func luaInstructionABC(op, a, b, c int) uint32 {
	return uint32(op) | uint32(a)<<6 | uint32(c)<<14 | uint32(b)<<23
}

func luaInstructionABx(op, a, bx int) uint32 {
	return uint32(op) | uint32(a)<<6 | uint32(bx)<<14
}
