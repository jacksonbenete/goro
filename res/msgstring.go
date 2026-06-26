package res

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

var msgStringTableCandidates = []string{
	"data\\msgstringtable.txt",
	"data/msgstringtable.txt",
	"msgstringtable.txt",
}

var msgStringCSVCandidates = []string{
	"data\\msgstringtable.csv",
	"data/msgstringtable.csv",
	"msgstringtable.csv",
}

var msgStringLuaCandidates = []string{
	"data\\luafiles514\\lua files\\MsgString_KR.lub",
	"data\\luafiles514\\lua files\\msgstring_kr.lub",
	"data\\lua files\\MsgString_KR.lub",
	"data\\lua files\\msgstring_kr.lub",
	"lua files\\MsgString_KR.lub",
	"lua files\\msgstring_kr.lub",
}

func (m *Manager) MsgString(id int) (string, bool) {
	if id < 0 {
		return "", false
	}
	m.loadMsgStrings()
	value, ok := m.msgStrings[id]
	return value, ok && value != ""
}

func (m *Manager) loadMsgStrings() {
	if m.msgStringsLoaded {
		return
	}
	m.msgStringsLoaded = true
	m.msgStrings = make(map[int]string)
	if _, data, ok := m.ReadFirst(msgStringTableCandidates); ok {
		m.msgStrings = parseMsgStringTable(data)
		if len(m.msgStrings) > 0 {
			return
		}
	}
	if _, data, ok := m.ReadFirst(msgStringCSVCandidates); ok {
		m.msgStrings = parseMsgStringCSV(data)
		if len(m.msgStrings) > 0 {
			return
		}
	}
	if _, data, ok := m.ReadFirst(msgStringLuaCandidates); ok {
		m.msgStrings = parseMsgStringLua(data)
	}
}

func parseMsgStringTable(data []byte) map[int]string {
	out := make(map[int]string)
	for _, rawLine := range strings.Split(decodeROText(data), "\n") {
		line := strings.TrimRight(rawLine, "\r\n")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		if hash := strings.LastIndexByte(line, '#'); hash >= 0 {
			line = line[:hash]
		}
		out[len(out)] = strings.TrimSpace(line)
	}
	return out
}

func parseMsgStringCSV(data []byte) map[int]string {
	out := make(map[int]string)
	raw := bytes.TrimSpace(data)
	base64Encoded := bytes.HasSuffix(raw, []byte("="))
	reader := csv.NewReader(bytes.NewReader(raw))
	reader.FieldsPerRecord = -1
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if len(record) < 2 {
			continue
		}
		value := strings.TrimSpace(record[1])
		if base64Encoded {
			decoded, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				continue
			}
			value = decodeROText(decoded)
		}
		if strings.TrimSpace(value) == "" || strings.HasPrefix(strings.TrimSpace(value), "//") {
			continue
		}
		out[len(out)] = strings.TrimSpace(value)
	}
	return out
}

func parseMsgStringLua(data []byte) map[int]string {
	globals := make(map[string]luaValue)
	if err := executeLua51Bytecode(data, globals); err != nil {
		return nil
	}
	for _, name := range []string{"MsgStringTable", "MsgStrID", "MsgString"} {
		table := globals[name]
		if table.kind != luaTable || len(table.table) == 0 {
			continue
		}
		out := make(map[int]string)
		for key, value := range table.table {
			if value.kind != luaString || value.str == "" {
				continue
			}
			switch k := key.(type) {
			case int:
				out[k-1] = strings.TrimSpace(value.str)
			case string:
				if id, err := strconv.Atoi(k); err == nil {
					out[id-1] = strings.TrimSpace(value.str)
				}
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

func decodeROText(data []byte) string {
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	if utf8.Valid(data) {
		return string(data)
	}
	decoded, _, err := transform.Bytes(korean.EUCKR.NewDecoder(), data)
	if err != nil {
		return string(data)
	}
	return string(decoded)
}
