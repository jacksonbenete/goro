package session

type Session struct {
	AccountID   uint32
	CharID      uint32
	AuthCode    uint32
	UserLevel   uint32
	Sex         byte
	Playing     bool
	CharServers []CharServer
	Characters  []Character
	Selected    Character
	Zone        ZoneServer
	PlayerX     int
	PlayerY     int
	PlayerDir   int
	Vitals      Vitals
}

func New() *Session {
	return &Session{}
}

type CharServer struct {
	Address   string
	Port      uint16
	Name      string
	UserCount uint16
	State     uint16
	Property  uint16
}

type Character struct {
	ID        uint32
	Name      string
	Slot      uint8
	Level     int16
	Job       int16
	HP        int16
	MaxHP     int16
	SP        int16
	MaxSP     int16
	Str       uint8
	Agi       uint8
	Vit       uint8
	Int       uint8
	Dex       uint8
	Luk       uint8
	Hair      int16
	HairColor uint8
	HeadPal   int16
	BodyPal   int16
	Weapon    int16
	Shield    int16
	HeadTop   int16
	HeadMid   int16
	HeadLow   int16
}

type Vitals struct {
	HP    int
	MaxHP int
	SP    int
	MaxSP int
}

type ZoneServer struct {
	Address string
	Port    uint16
	MapName string
}
