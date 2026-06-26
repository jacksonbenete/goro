package network

import (
	"bytes"
	"encoding/binary"
)

type Writer struct {
	bytes.Buffer
}

func (w *Writer) Uint8(value uint8) {
	_ = w.WriteByte(value)
}

func (w *Writer) Uint16(value uint16) {
	var tmp [2]byte
	binary.LittleEndian.PutUint16(tmp[:], value)
	_, _ = w.Write(tmp[:])
}

func (w *Writer) Uint32(value uint32) {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], value)
	_, _ = w.Write(tmp[:])
}

func (w *Writer) CString(value string, size int) {
	buf := make([]byte, size)
	copy(buf, []byte(value))
	_, _ = w.Write(buf)
}
