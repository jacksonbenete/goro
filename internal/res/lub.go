package res

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
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
	return executeLuaFunction(fn, globals)
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
	r.string()
	r.i32()
	r.i32()
	r.u8()
	r.u8()
	r.u8()
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
	for i := 0; i < upvalueCount; i++ {
		r.string()
	}
	if r.err != nil {
		return luaFunction{}, r.err
	}
	return luaFunction{code: code, constants: constants, prototypes: protos, maxStackSize: maxStack}, nil
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
)
