package network

const (
	PacketCAPlainLogin uint16 = 0x0064
	PacketCAEnter      uint16 = 0x0065
	PacketCZSelectChar uint16 = 0x0066
	PacketCZLoadEndAck uint16 = 0x007D
	PacketCZTickSend   uint16 = 0x0089
	PacketCZReqName    uint16 = 0x0094
	PacketCZWalkToXY   uint16 = 0x00A7
	PacketCZWalkToXYRE uint16 = 0x035F
	PacketCZEnter2     uint16 = 0x0436
)

type AccountLogin struct {
	Version    uint32
	Username   string
	Password   string
	ClientType uint8
}

func BuildAccountLoginPacket(login AccountLogin) []byte {
	var w Writer
	w.Uint16(PacketCAPlainLogin)
	w.Uint32(login.Version)
	w.CString(login.Username, 24)
	w.CString(login.Password, 24)
	w.Uint8(login.ClientType)
	return w.Bytes()
}

type CharServerEnter struct {
	AccountID uint32
	AuthCode  uint32
	UserLevel uint32
	Sex       uint8
}

func BuildCharServerEnterPacket(enter CharServerEnter) []byte {
	var w Writer
	w.Uint16(PacketCAEnter)
	w.Uint32(enter.AccountID)
	w.Uint32(enter.AuthCode)
	w.Uint32(enter.UserLevel)
	w.Uint16(0)
	w.Uint8(enter.Sex)
	return w.Bytes()
}

func BuildSelectCharacterPacket(slot uint8) []byte {
	var w Writer
	w.Uint16(PacketCZSelectChar)
	w.Uint8(slot)
	return w.Bytes()
}

func BuildLoadEndAckPacket() []byte {
	var w Writer
	w.Uint16(PacketCZLoadEndAck)
	return w.Bytes()
}

func BuildTickSendPacket(clientTick uint32) []byte {
	var w Writer
	w.Uint16(PacketCZTickSend)
	w.Uint8(0)
	w.Uint8(0)
	w.Uint32(clientTick)
	return w.Bytes()
}

func BuildNameRequestPacket(gid uint32) []byte {
	var w Writer
	w.Uint16(PacketCZReqName)
	w.Uint32(gid)
	return w.Bytes()
}

type MapServerEnter struct {
	AccountID  uint32
	CharID     uint32
	AuthCode   uint32
	ClientTick uint32
	Sex        uint8
}

func BuildMapServerEnterPacket(enter MapServerEnter) []byte {
	var w Writer
	w.Uint16(PacketCZEnter2)
	w.Uint32(enter.AccountID)
	w.Uint32(enter.CharID)
	w.Uint32(enter.AuthCode)
	w.Uint32(enter.ClientTick)
	w.Uint8(enter.Sex)
	return w.Bytes()
}

func BuildWalkToXYPacket(x, y int) ([]byte, bool) {
	return BuildWalkToXYPacketForClientDate(x, y, 20080910)
}

func BuildWalkToXYPacketForClientDate(x, y, clientDate int) ([]byte, bool) {
	dest, ok := EncodeMoveDestination(x, y)
	if !ok {
		return nil, false
	}

	var w Writer
	if clientDate >= 20211103 {
		w.Uint16(PacketCZWalkToXYRE)
	} else {
		w.Uint16(PacketCZWalkToXY)
		w.Uint8(0)
		w.Uint8(0)
		w.Uint8(0)
	}
	_, _ = w.Write(dest[:])
	return w.Bytes(), true
}

func EncodeMoveDestination(x, y int) ([3]byte, bool) {
	if x < 0 || x > 1023 || y < 0 || y > 1023 {
		return [3]byte{}, false
	}
	return [3]byte{
		byte((x >> 2) & 0xff),
		byte(((x & 0x03) << 6) | ((y >> 4) & 0x3f)),
		byte((y & 0x0f) << 4),
	}, true
}
