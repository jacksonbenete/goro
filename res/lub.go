package res

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"unsafe"

	glua "github.com/yuin/gopher-lua"
)

type luaValue struct {
	kind  luaKind
	num   float64
	str   string
	table map[interface{}]luaValue
}

type luaKind int

const (
	luaNil luaKind = iota
	luaBool
	luaNumber
	luaString
	luaTable
)

type luaFunction struct {
	code         []uint32
	constants    []luaValue
	prototypes   []luaFunction
	sourceName   string
	lineDefined  int
	lastLine     int
	upvalues     []string
	numUpvalues  int
	numParams    int
	isVararg     int
	maxStackSize int
}

func parseLua51Bytecode(data []byte) (luaFunction, error) {
	reader := &luaReader{data: data}
	if err := reader.header(); err != nil {
		return luaFunction{}, err
	}
	fn, err := reader.function()
	if err != nil {
		return luaFunction{}, err
	}
	if reader.err != nil {
		return luaFunction{}, reader.err
	}
	return fn, nil
}

func executeLua51Bytecode(data []byte, globals map[string]luaValue) error {
	fn, err := parseLua51Bytecode(data)
	if err != nil {
		return err
	}
	if err := executeLuaFunctionWithGopherLua(fn, globals, nil); err == nil {
		return nil
	}
	return executeLuaFunction(fn, globals)
}

func executeLua51BytecodeWithSetup(data []byte, globals map[string]luaValue, setup func(*glua.LState)) error {
	fn, err := parseLua51Bytecode(data)
	if err != nil {
		return err
	}
	return executeLuaFunctionWithGopherLua(fn, globals, setup)
}

type luaReader struct {
	data      []byte
	off       int
	err       error
	sizeTSize int
}

func (r *luaReader) header() error {
	header := r.bytes(12)
	if r.err != nil {
		return r.err
	}
	if string(header[:4]) != "\x1bLua" || header[4] != 0x51 || header[5] != 0 || header[6] != 1 || header[7] != 4 || header[9] != 4 || header[10] != 8 {
		return fmt.Errorf("unsupported Lua bytecode header")
	}
	r.sizeTSize = int(header[8])
	if r.sizeTSize != 4 && r.sizeTSize != 8 {
		return fmt.Errorf("unsupported Lua size_t width %d", r.sizeTSize)
	}
	return nil
}

func (r *luaReader) function() (luaFunction, error) {
	sourceName := r.string()
	lineDefined := int(r.i32())
	lastLine := int(r.i32())
	numUpvalues := int(r.u8())
	numParams := int(r.u8())
	isVararg := int(r.u8())
	maxStack := int(r.u8())

	codeCount := int(r.i32())
	code := make([]uint32, codeCount)
	for i := range code {
		code[i] = r.u32()
	}

	constantCount := int(r.i32())
	constants := make([]luaValue, constantCount)
	for i := range constants {
		switch t := r.u8(); t {
		case 0:
			constants[i] = luaValue{kind: luaNil}
		case 1:
			if r.u8() != 0 {
				constants[i] = luaValue{kind: luaBool, num: 1}
			} else {
				constants[i] = luaValue{kind: luaBool}
			}
		case 3:
			constants[i] = luaValue{kind: luaNumber, num: math.Float64frombits(r.u64())}
		case 4:
			constants[i] = luaValue{kind: luaString, str: r.string()}
		default:
			return luaFunction{}, fmt.Errorf("unsupported Lua constant type %d", t)
		}
	}

	protoCount := int(r.i32())
	protos := make([]luaFunction, protoCount)
	for i := range protos {
		proto, err := r.function()
		if err != nil {
			return luaFunction{}, err
		}
		protos[i] = proto
	}

	lineInfoCount := int(r.i32())
	r.skip(lineInfoCount * 4)
	localCount := int(r.i32())
	for i := 0; i < localCount; i++ {
		r.string()
		r.i32()
		r.i32()
	}
	upvalueCount := int(r.i32())
	upvalues := make([]string, upvalueCount)
	for i := 0; i < upvalueCount; i++ {
		upvalues[i] = r.string()
	}
	if r.err != nil {
		return luaFunction{}, r.err
	}
	return luaFunction{
		code:         code,
		constants:    constants,
		prototypes:   protos,
		sourceName:   sourceName,
		lineDefined:  lineDefined,
		lastLine:     lastLine,
		upvalues:     upvalues,
		numUpvalues:  numUpvalues,
		numParams:    numParams,
		isVararg:     isVararg,
		maxStackSize: maxStack,
	}, nil
}

func executeLuaFunctionWithGopherLua(fn luaFunction, globals map[string]luaValue, setup func(*glua.LState)) error {
	if globals == nil {
		return errors.New("nil Lua globals")
	}
	proto, err := gopherLuaProto(fn)
	if err != nil {
		return err
	}
	state := glua.NewState(glua.Options{SkipOpenLibs: true})
	defer state.Close()
	openLuaResourceLibs(state)
	if setup != nil {
		setup(state)
	}
	for key, value := range globals {
		if value.kind != luaNil {
			state.SetGlobal(key, luaValueToGopherLua(state, value))
		}
	}
	lfn := state.NewFunctionFromProto(proto)
	if err := state.CallByParam(glua.P{Fn: lfn, NRet: 0, Protect: true}); err != nil {
		return err
	}
	state.G.Global.ForEach(func(key glua.LValue, value glua.LValue) {
		name, ok := key.(glua.LString)
		if !ok || luaBuiltinGlobal(string(name)) {
			return
		}
		if converted, ok := gopherLuaToLuaValue(value); ok {
			globals[string(name)] = converted
		}
	})
	return nil
}

func openLuaResourceLibs(state *glua.LState) {
	for _, lib := range []struct {
		name string
		open glua.LGFunction
	}{
		{glua.BaseLibName, glua.OpenBase},
		{glua.TabLibName, glua.OpenTable},
		{glua.StringLibName, glua.OpenString},
		{glua.MathLibName, glua.OpenMath},
	} {
		state.Push(state.NewFunction(lib.open))
		state.Push(glua.LString(lib.name))
		state.Call(1, 0)
	}
	for _, unsafeName := range []string{"dofile", "loadfile"} {
		state.SetGlobal(unsafeName, glua.LNil)
	}
}

func gopherLuaProto(fn luaFunction) (*glua.FunctionProto, error) {
	constants := make([]glua.LValue, len(fn.constants))
	stringConstants := make([]string, len(fn.constants))
	for i, value := range fn.constants {
		constants[i] = luaValueToGopherLua(nil, value)
		if value.kind == luaString {
			stringConstants[i] = value.str
		}
	}
	code := make([]uint32, len(fn.code))
	for i, instruction := range fn.code {
		translated, err := translateLua51Instruction(instruction)
		if err != nil {
			return nil, err
		}
		code[i] = translated
	}
	protos := make([]*glua.FunctionProto, len(fn.prototypes))
	for i, child := range fn.prototypes {
		proto, err := gopherLuaProto(child)
		if err != nil {
			return nil, err
		}
		protos[i] = proto
	}
	dbgLines := make([]int, len(code))
	upvalues := make([]string, len(fn.upvalues))
	copy(upvalues, fn.upvalues)
	proto := &glua.FunctionProto{
		SourceName:         fn.sourceName,
		LineDefined:        fn.lineDefined,
		LastLineDefined:    fn.lastLine,
		NumUpvalues:        uint8(fn.numUpvalues),
		NumParameters:      uint8(fn.numParams),
		IsVarArg:           uint8(fn.isVararg),
		NumUsedRegisters:   uint8(fn.maxStackSize),
		Code:               code,
		Constants:          constants,
		FunctionPrototypes: protos,
		DbgSourcePositions: dbgLines,
		DbgLocals:          nil,
		DbgCalls:           nil,
		DbgUpvalues:        upvalues,
	}
	setGopherLuaStringConstants(proto, stringConstants)
	return proto, nil
}

func setGopherLuaStringConstants(proto *glua.FunctionProto, constants []string) {
	field := reflect.ValueOf(proto).Elem().FieldByName("stringConstants")
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(constants))
}

func luaBuiltinGlobal(name string) bool {
	switch name {
	case "_G", "_VERSION",
		"assert", "collectgarbage", "dofile", "error", "getfenv", "getmetatable",
		"ipairs", "load", "loadfile", "loadstring", "module", "next", "pairs",
		"pcall", "print", "rawequal", "rawget", "rawset", "require", "select",
		"setfenv", "setmetatable", "tonumber", "tostring", "type", "unpack", "xpcall",
		"math", "string", "table":
		return true
	default:
		return false
	}
}

func translateLua51Instruction(instruction uint32) (uint32, error) {
	op := luaInstructionOp(instruction)
	if op >= len(lua51ToGopherLuaOpcode) {
		return 0, fmt.Errorf("unsupported Lua opcode %d", op)
	}
	mapped := lua51ToGopherLuaOpcode[op]
	a := luaInstructionA(instruction)
	switch lua51OpcodeMode[op] {
	case luaOpcodeABx, luaOpcodeAsBx:
		return gopherLuaInstructionABx(mapped, a, luaInstructionBx(instruction)), nil
	default:
		return gopherLuaInstructionABC(mapped, a, luaInstructionB(instruction), luaInstructionC(instruction)), nil
	}
}

func luaInstructionOp(instruction uint32) int {
	return int(instruction & 0x3f)
}

func luaInstructionA(instruction uint32) int {
	return int((instruction >> 6) & 0xff)
}

func luaInstructionC(instruction uint32) int {
	return int((instruction >> 14) & 0x1ff)
}

func luaInstructionB(instruction uint32) int {
	return int((instruction >> 23) & 0x1ff)
}

func luaInstructionBx(instruction uint32) int {
	return int((instruction >> 14) & 0x3ffff)
}

func gopherLuaInstructionABC(op, a, b, c int) uint32 {
	return uint32(op)<<26 | uint32(a&0xff)<<18 | uint32(c&0x1ff)<<9 | uint32(b&0x1ff)
}

func gopherLuaInstructionABx(op, a, bx int) uint32 {
	return uint32(op)<<26 | uint32(a&0xff)<<18 | uint32(bx&0x3ffff)
}

func luaValueToGopherLua(state *glua.LState, value luaValue) glua.LValue {
	switch value.kind {
	case luaBool:
		return glua.LBool(value.num != 0)
	case luaNumber:
		return glua.LNumber(value.num)
	case luaString:
		return glua.LString(value.str)
	case luaTable:
		if state == nil {
			return glua.LNil
		}
		var table *glua.LTable
		table = state.NewTable()
		for key, child := range value.table {
			table.RawSet(luaKeyToGopherLua(key), luaValueToGopherLua(state, child))
		}
		return table
	default:
		return glua.LNil
	}
}

func luaKeyToGopherLua(key interface{}) glua.LValue {
	switch key := key.(type) {
	case int:
		return glua.LNumber(key)
	case float64:
		return glua.LNumber(key)
	case string:
		return glua.LString(key)
	default:
		return glua.LNil
	}
}

func gopherLuaToLuaValue(value glua.LValue) (luaValue, bool) {
	switch value := value.(type) {
	case glua.LBool:
		if bool(value) {
			return luaValue{kind: luaBool, num: 1}, true
		}
		return luaValue{kind: luaBool}, true
	case glua.LNumber:
		return luaValue{kind: luaNumber, num: float64(value)}, true
	case glua.LString:
		return luaValue{kind: luaString, str: string(value)}, true
	case *glua.LTable:
		out := luaValue{kind: luaTable, table: make(map[interface{}]luaValue)}
		value.ForEach(func(key glua.LValue, child glua.LValue) {
			if converted, ok := gopherLuaToLuaValue(child); ok {
				out.table[gopherLuaKey(key)] = converted
			}
		})
		return out, true
	default:
		return luaValue{}, false
	}
}

func gopherLuaKey(key glua.LValue) interface{} {
	switch key := key.(type) {
	case glua.LNumber:
		num := float64(key)
		if math.Trunc(num) == num {
			return int(num)
		}
		return num
	case glua.LString:
		return string(key)
	case glua.LBool:
		return bool(key)
	default:
		return nil
	}
}

func (r *luaReader) string() string {
	size := r.sizeT()
	if size == 0 {
		return ""
	}
	data := r.bytes(int(size))
	if r.err != nil {
		return ""
	}
	if len(data) > 0 && data[len(data)-1] == 0 {
		data = data[:len(data)-1]
	}
	return string(data)
}

func (r *luaReader) sizeT() uint64 {
	switch r.sizeTSize {
	case 4:
		return uint64(r.u32())
	case 8:
		return r.u64()
	default:
		r.err = fmt.Errorf("unsupported Lua size_t width %d", r.sizeTSize)
		return 0
	}
}

func (r *luaReader) u8() byte {
	data := r.bytes(1)
	if len(data) == 0 {
		return 0
	}
	return data[0]
}

func (r *luaReader) i32() int32 {
	return int32(r.u32())
}

func (r *luaReader) u32() uint32 {
	data := r.bytes(4)
	if len(data) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(data)
}

func (r *luaReader) u64() uint64 {
	data := r.bytes(8)
	if len(data) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(data)
}

func (r *luaReader) skip(n int) {
	r.bytes(n)
}

func (r *luaReader) bytes(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || r.off+n > len(r.data) {
		r.err = errors.New("unexpected end of Lua bytecode")
		return nil
	}
	out := r.data[r.off : r.off+n]
	r.off += n
	return out
}

func executeLuaFunction(fn luaFunction, globals map[string]luaValue) error {
	if globals == nil {
		return errors.New("nil Lua globals")
	}
	regs := make([]luaValue, fn.maxStackSize+64)
	rk := func(index int) luaValue {
		if index&luaBitRK != 0 {
			constIndex := index &^ luaBitRK
			if constIndex >= 0 && constIndex < len(fn.constants) {
				return fn.constants[constIndex]
			}
			return luaValue{}
		}
		if index >= 0 && index < len(regs) {
			return regs[index]
		}
		return luaValue{}
	}

	for pc := 0; pc < len(fn.code); pc++ {
		instruction := fn.code[pc]
		op := int(instruction & 0x3f)
		a := int((instruction >> 6) & 0xff)
		c := int((instruction >> 14) & 0x1ff)
		b := int((instruction >> 23) & 0x1ff)
		bx := int((instruction >> 14) & 0x3ffff)
		switch op {
		case luaOpMove:
			regs[a] = regs[b]
		case luaOpLoadK:
			regs[a] = luaConstant(fn, bx)
		case luaOpLoadBool:
			regs[a] = luaValue{kind: luaBool}
			if b != 0 {
				regs[a].num = 1
			}
			if c != 0 {
				pc++
			}
		case luaOpLoadNil:
			for index := a; index <= b && index < len(regs); index++ {
				regs[index] = luaValue{}
			}
		case luaOpGetGlobal:
			name := luaConstant(fn, bx).str
			regs[a] = globals[name]
		case luaOpGetTable:
			regs[a] = tableGet(regs[b], rk(c))
		case luaOpSetGlobal:
			name := luaConstant(fn, bx).str
			globals[name] = regs[a]
		case luaOpSetTable:
			tableSet(regs[a], rk(b), rk(c))
		case luaOpNewTable:
			regs[a] = luaValue{kind: luaTable, table: make(map[interface{}]luaValue)}
		case luaOpSetList:
			count := b
			if count == 0 {
				count = len(regs) - a - 1
			}
			block := c
			if block == 0 {
				pc++
				if pc >= len(fn.code) {
					return errors.New("lua SETLIST missing extension word")
				}
				block = int(fn.code[pc])
			}
			base := (block - 1) * luaFieldsPerFlush
			for i := 1; i <= count; i++ {
				tableSet(regs[a], luaValue{kind: luaNumber, num: float64(base + i)}, regs[a+i])
			}
		case luaOpReturn:
			return nil
		default:
			return fmt.Errorf("unsupported Lua opcode %d", op)
		}
	}
	return nil
}

func luaConstant(fn luaFunction, index int) luaValue {
	if index < 0 || index >= len(fn.constants) {
		return luaValue{}
	}
	return fn.constants[index]
}

func tableGet(table luaValue, key luaValue) luaValue {
	if table.kind != luaTable || table.table == nil {
		return luaValue{}
	}
	return table.table[luaKey(key)]
}

func tableSet(table luaValue, key luaValue, value luaValue) {
	if table.kind != luaTable || table.table == nil {
		return
	}
	table.table[luaKey(key)] = value
}

func luaKey(value luaValue) interface{} {
	switch value.kind {
	case luaString:
		return value.str
	case luaNumber:
		if math.Trunc(value.num) == value.num {
			return int(value.num)
		}
		return value.num
	default:
		return nil
	}
}

const (
	luaBitRK          = 1 << 8
	luaFieldsPerFlush = 50
)

const (
	luaOpMove = iota
	luaOpLoadK
	luaOpLoadBool
	luaOpLoadNil
	luaOpGetUpval
	luaOpGetGlobal
	luaOpGetTable
	luaOpSetGlobal
	luaOpSetupval
	luaOpSetTable
	luaOpNewTable
	luaOpSelf
	luaOpAdd
	luaOpSub
	luaOpMul
	luaOpDiv
	luaOpMod
	luaOpPow
	luaOpUnm
	luaOpNot
	luaOpLen
	luaOpConcat
	luaOpJmp
	luaOpEq
	luaOpLt
	luaOpLe
	luaOpTest
	luaOpTestSet
	luaOpCall
	luaOpTailCall
	luaOpReturn
	luaOpForLoop
	luaOpForPrep
	luaOpTForLoop
	luaOpSetList
	luaOpClose
	luaOpClosure
	luaOpVararg
)

type luaOpcodeMode int

const (
	luaOpcodeABC luaOpcodeMode = iota
	luaOpcodeABx
	luaOpcodeAsBx
)

var lua51ToGopherLuaOpcode = [...]int{
	luaOpMove:      glua.OP_MOVE,
	luaOpLoadK:     glua.OP_LOADK,
	luaOpLoadBool:  glua.OP_LOADBOOL,
	luaOpLoadNil:   glua.OP_LOADNIL,
	luaOpGetUpval:  glua.OP_GETUPVAL,
	luaOpGetGlobal: glua.OP_GETGLOBAL,
	luaOpGetTable:  glua.OP_GETTABLE,
	luaOpSetGlobal: glua.OP_SETGLOBAL,
	luaOpSetupval:  glua.OP_SETUPVAL,
	luaOpSetTable:  glua.OP_SETTABLE,
	luaOpNewTable:  glua.OP_NEWTABLE,
	luaOpSelf:      glua.OP_SELF,
	luaOpAdd:       glua.OP_ADD,
	luaOpSub:       glua.OP_SUB,
	luaOpMul:       glua.OP_MUL,
	luaOpDiv:       glua.OP_DIV,
	luaOpMod:       glua.OP_MOD,
	luaOpPow:       glua.OP_POW,
	luaOpUnm:       glua.OP_UNM,
	luaOpNot:       glua.OP_NOT,
	luaOpLen:       glua.OP_LEN,
	luaOpConcat:    glua.OP_CONCAT,
	luaOpJmp:       glua.OP_JMP,
	luaOpEq:        glua.OP_EQ,
	luaOpLt:        glua.OP_LT,
	luaOpLe:        glua.OP_LE,
	luaOpTest:      glua.OP_TEST,
	luaOpTestSet:   glua.OP_TESTSET,
	luaOpCall:      glua.OP_CALL,
	luaOpTailCall:  glua.OP_TAILCALL,
	luaOpReturn:    glua.OP_RETURN,
	luaOpForLoop:   glua.OP_FORLOOP,
	luaOpForPrep:   glua.OP_FORPREP,
	luaOpTForLoop:  glua.OP_TFORLOOP,
	luaOpSetList:   glua.OP_SETLIST,
	luaOpClose:     glua.OP_CLOSE,
	luaOpClosure:   glua.OP_CLOSURE,
	luaOpVararg:    glua.OP_VARARG,
}

var lua51OpcodeMode = [...]luaOpcodeMode{
	luaOpMove:      luaOpcodeABC,
	luaOpLoadK:     luaOpcodeABx,
	luaOpLoadBool:  luaOpcodeABC,
	luaOpLoadNil:   luaOpcodeABC,
	luaOpGetUpval:  luaOpcodeABC,
	luaOpGetGlobal: luaOpcodeABx,
	luaOpGetTable:  luaOpcodeABC,
	luaOpSetGlobal: luaOpcodeABx,
	luaOpSetupval:  luaOpcodeABC,
	luaOpSetTable:  luaOpcodeABC,
	luaOpNewTable:  luaOpcodeABC,
	luaOpSelf:      luaOpcodeABC,
	luaOpAdd:       luaOpcodeABC,
	luaOpSub:       luaOpcodeABC,
	luaOpMul:       luaOpcodeABC,
	luaOpDiv:       luaOpcodeABC,
	luaOpMod:       luaOpcodeABC,
	luaOpPow:       luaOpcodeABC,
	luaOpUnm:       luaOpcodeABC,
	luaOpNot:       luaOpcodeABC,
	luaOpLen:       luaOpcodeABC,
	luaOpConcat:    luaOpcodeABC,
	luaOpJmp:       luaOpcodeAsBx,
	luaOpEq:        luaOpcodeABC,
	luaOpLt:        luaOpcodeABC,
	luaOpLe:        luaOpcodeABC,
	luaOpTest:      luaOpcodeABC,
	luaOpTestSet:   luaOpcodeABC,
	luaOpCall:      luaOpcodeABC,
	luaOpTailCall:  luaOpcodeABC,
	luaOpReturn:    luaOpcodeABC,
	luaOpForLoop:   luaOpcodeAsBx,
	luaOpForPrep:   luaOpcodeAsBx,
	luaOpTForLoop:  luaOpcodeABC,
	luaOpSetList:   luaOpcodeABC,
	luaOpClose:     luaOpcodeABC,
	luaOpClosure:   luaOpcodeABx,
	luaOpVararg:    luaOpcodeABC,
}
