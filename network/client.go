package network

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Client struct {
	clientDate int
	trace      bool
	framer     *Framer

	mu      sync.Mutex
	conn    net.Conn
	packets []Packet
	status  string
	errs    []error
}

func NewClient(clientDate int, trace bool) *Client {
	return &Client{
		clientDate: clientDate,
		trace:      trace,
		framer:     NewFramer(PacketLengths2008()),
		status:     "offline",
	}
}

func (c *Client) Connect(ctx context.Context, address string, port int) error {
	c.Close()

	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		c.setStatus("connect failed")
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.status = fmt.Sprintf("connected to %s:%d", address, port)
	c.mu.Unlock()

	go c.readLoop(conn)
	return nil
}

func (c *Client) Close() {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.status = "offline"
	c.mu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
}

func (c *Client) Send(data []byte) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("not connected")
	}

	if c.trace {
		headLen := min(len(data), 32)
		log.Printf("network write n=%d head=%s", len(data), hex.EncodeToString(data[:headLen]))
	}
	_, err := conn.Write(data)
	if err != nil {
		c.clearConn(conn)
		c.addError(err)
	}
	return err
}

func (c *Client) SendAccountLogin(username, password string, version uint32, clientType uint8) error {
	packet := BuildAccountLoginPacket(AccountLogin{
		Version:    version,
		Username:   username,
		Password:   password,
		ClientType: clientType,
	})
	return c.Send(packet)
}

func (c *Client) SendCharServerEnter(accountID, authCode, userLevel uint32, sex uint8) error {
	packet := BuildCharServerEnterPacket(CharServerEnter{
		AccountID: accountID,
		AuthCode:  authCode,
		UserLevel: userLevel,
		Sex:       sex,
	})
	return c.Send(packet)
}

func (c *Client) SendSelectCharacter(slot uint8) error {
	return c.Send(BuildSelectCharacterPacket(slot))
}

func (c *Client) SendMakeCharacter(character MakeCharacter) error {
	packet := BuildMakeCharacterPacket(character)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CH_MAKE_CHAR opcode=0x%04X name=%q slot=%d stats=%d/%d/%d/%d/%d/%d hair_color=%d hair_style=%d client_date=%d",
			ID(packet), character.Name, character.Slot, character.Str, character.Agi, character.Vit, character.Int, character.Dex, character.Luk, character.HairColor, character.HairStyle, c.clientDate)
	} else {
		log.Printf("send CH_MAKE_CHAR failed opcode=0x%04X len=%d name=%q slot=%d client_date=%d: %v",
			ID(packet), len(packet), character.Name, character.Slot, c.clientDate, err)
	}
	return err
}

func (c *Client) SendLoadEndAck() error {
	return c.Send(BuildLoadEndAckPacket())
}

func (c *Client) SendRestart(restartType uint8) error {
	packet := BuildRestartPacket(restartType)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_RESTART opcode=0x%04X type=%d client_date=%d", ID(packet), restartType, c.clientDate)
	} else {
		log.Printf("send CZ_RESTART failed opcode=0x%04X len=%d type=%d client_date=%d: %v", ID(packet), len(packet), restartType, c.clientDate, err)
	}
	return err
}

func (c *Client) SendTick(clientTick uint32) error {
	packet := BuildTickSendPacketForClientDate(clientTick, c.clientDate)
	err := c.Send(packet)
	if err == nil && c.trace {
		log.Printf("sent CZ_REQUEST_TIME opcode=0x%04X tick=%d client_date=%d", ID(packet), clientTick, c.clientDate)
	} else if err != nil {
		log.Printf("send CZ_REQUEST_TIME failed opcode=0x%04X len=%d client_date=%d: %v", ID(packet), len(packet), c.clientDate, err)
	}
	return err
}

func (c *Client) SendNameRequest(gid uint32) error {
	packet, ok := BuildNameRequestPacketForClientDate(gid, c.clientDate)
	if !ok {
		log.Printf("skip CZ_REQNAME target=%d client_date=%d: unsupported packet profile", gid, c.clientDate)
		return nil
	}
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_REQNAME opcode=0x%04X len=%d target=%d client_date=%d", ID(packet), len(packet), gid, c.clientDate)
	} else {
		log.Printf("send CZ_REQNAME failed opcode=0x%04X len=%d target=%d client_date=%d: %v", ID(packet), len(packet), gid, c.clientDate, err)
	}
	return err
}

func (c *Client) SendMapServerEnter(accountID, charID, authCode, clientTick uint32, sex uint8) error {
	packet := BuildMapServerEnterPacketForClientDate(MapServerEnter{
		AccountID:  accountID,
		CharID:     charID,
		AuthCode:   authCode,
		ClientTick: clientTick,
		Sex:        sex,
	}, c.clientDate)
	if c.trace {
		log.Printf("sent CZ_ENTER2 opcode=0x%04X len=%d client_date=%d sex_offset=%d", ID(packet), len(packet), c.clientDate, len(packet)-1)
	}
	return c.Send(packet)
}

func (c *Client) SendWalkToXY(x, y int) error {
	packet, ok := BuildWalkToXYPacketForClientDate(x, y, c.clientDate)
	if !ok {
		return fmt.Errorf("invalid walk destination %d,%d", x, y)
	}
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_REQUEST_MOVE opcode=0x%04X dst=%d,%d client_date=%d", ID(packet), x, y, c.clientDate)
	} else {
		log.Printf("send CZ_REQUEST_MOVE failed opcode=0x%04X len=%d dst=%d,%d client_date=%d: %v", ID(packet), len(packet), x, y, c.clientDate, err)
	}
	return err
}

func (c *Client) SendChangeDirection(headDir, dir uint8) error {
	packet := BuildChangeDirectionPacketForClientDate(headDir, dir, c.clientDate)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_CHANGE_DIRECTION opcode=0x%04X head_dir=%d dir=%d client_date=%d", ID(packet), headDir, dir&7, c.clientDate)
	} else {
		log.Printf("send CZ_CHANGE_DIRECTION failed opcode=0x%04X len=%d head_dir=%d dir=%d client_date=%d: %v", ID(packet), len(packet), headDir, dir&7, c.clientDate, err)
	}
	return err
}

func (c *Client) SendActionRequest(targetGID uint32, action uint8) error {
	packet := BuildActionRequestPacketForClientDate(targetGID, action, c.clientDate)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_REQUEST_ACT opcode=0x%04X target=%d action=%d client_date=%d", ID(packet), targetGID, action, c.clientDate)
	} else {
		log.Printf("send CZ_REQUEST_ACT failed opcode=0x%04X len=%d target=%d action=%d client_date=%d: %v", ID(packet), len(packet), targetGID, action, c.clientDate, err)
	}
	return err
}

func (c *Client) SendNPCContact(npcID uint32) error {
	packet := BuildNPCContactPacket(npcID, 0)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_CONTACTNPC opcode=0x%04X npc=%d client_date=%d", ID(packet), npcID, c.clientDate)
	} else {
		log.Printf("send CZ_CONTACTNPC failed opcode=0x%04X npc=%d client_date=%d: %v", ID(packet), npcID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendNPCNext(npcID uint32) error {
	packet := BuildNPCNextPacket(npcID)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_REQ_NEXT_SCRIPT opcode=0x%04X npc=%d client_date=%d", ID(packet), npcID, c.clientDate)
	} else {
		log.Printf("send CZ_REQ_NEXT_SCRIPT failed opcode=0x%04X npc=%d client_date=%d: %v", ID(packet), npcID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendNPCClose(npcID uint32) error {
	packet := BuildNPCClosePacket(npcID)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_CLOSE_DIALOG opcode=0x%04X npc=%d client_date=%d", ID(packet), npcID, c.clientDate)
	} else {
		log.Printf("send CZ_CLOSE_DIALOG failed opcode=0x%04X npc=%d client_date=%d: %v", ID(packet), npcID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendNPCMenuChoice(npcID uint32, choice uint8) error {
	packet := BuildNPCMenuChoicePacket(npcID, choice)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_CHOOSE_MENU opcode=0x%04X npc=%d choice=%d client_date=%d", ID(packet), npcID, choice, c.clientDate)
	} else {
		log.Printf("send CZ_CHOOSE_MENU failed opcode=0x%04X npc=%d choice=%d client_date=%d: %v", ID(packet), npcID, choice, c.clientDate, err)
	}
	return err
}

func (c *Client) SendStatusIncrease(statusID uint16) error {
	packet := BuildStatusIncreasePacket(statusID)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_STATUS_CHANGE opcode=0x%04X status=%d amount=1 client_date=%d", ID(packet), statusID, c.clientDate)
	} else {
		log.Printf("send CZ_STATUS_CHANGE failed opcode=0x%04X len=%d status=%d client_date=%d: %v", ID(packet), len(packet), statusID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendSkillLevelUp(skillID uint16) error {
	packet := BuildSkillLevelUpPacket(skillID)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_UPGRADE_SKILLLEVEL opcode=0x%04X skill=%d client_date=%d", ID(packet), skillID, c.clientDate)
	} else {
		log.Printf("send CZ_UPGRADE_SKILLLEVEL failed opcode=0x%04X len=%d skill=%d client_date=%d: %v", ID(packet), len(packet), skillID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendRememberWarpPoint() error {
	packet := BuildRememberWarpPointPacket()
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_REMEMBER_WARPPOINT opcode=0x%04X client_date=%d", ID(packet), c.clientDate)
	} else {
		log.Printf("send CZ_REMEMBER_WARPPOINT failed opcode=0x%04X len=%d client_date=%d: %v", ID(packet), len(packet), c.clientDate, err)
	}
	return err
}

func (c *Client) SendUseSkillToID(skillID, level uint16, targetID uint32) error {
	packet := BuildUseSkillToIDPacketForClientDate(skillID, level, targetID, c.clientDate)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_USE_SKILL opcode=0x%04X skill=%d level=%d target=%d client_date=%d", ID(packet), skillID, level, targetID, c.clientDate)
	} else {
		log.Printf("send CZ_USE_SKILL failed opcode=0x%04X len=%d skill=%d level=%d target=%d client_date=%d: %v", ID(packet), len(packet), skillID, level, targetID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendUseSkillToGround(skillID, level uint16, x, y int) error {
	packet := BuildUseSkillToGroundPacketForClientDate(skillID, level, x, y, c.clientDate)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_USE_SKILL_TOGROUND opcode=0x%04X skill=%d level=%d dst=%d,%d client_date=%d", ID(packet), skillID, level, x, y, c.clientDate)
	} else {
		log.Printf("send CZ_USE_SKILL_TOGROUND failed opcode=0x%04X len=%d skill=%d level=%d dst=%d,%d client_date=%d: %v", ID(packet), len(packet), skillID, level, x, y, c.clientDate, err)
	}
	return err
}

func (c *Client) SendSelectWarpPoint(skillID uint16, mapName string) error {
	packet := BuildSelectWarpPointPacket(skillID, mapName)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_SELECT_WARPPOINT opcode=0x%04X skill=%d map=%q client_date=%d", ID(packet), skillID, mapName, c.clientDate)
	} else {
		log.Printf("send CZ_SELECT_WARPPOINT failed opcode=0x%04X len=%d skill=%d map=%q client_date=%d: %v", ID(packet), len(packet), skillID, mapName, c.clientDate, err)
	}
	return err
}

func (c *Client) Pump() {
}

func (c *Client) Status() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

func (c *Client) DrainPackets() []Packet {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]Packet, len(c.packets))
	copy(out, c.packets)
	c.packets = c.packets[:0]
	return out
}

func (c *Client) DrainErrors() []error {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]error, len(c.errs))
	copy(out, c.errs)
	c.errs = c.errs[:0]
	return out
}

func (c *Client) readLoop(conn net.Conn) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			if c.trace {
				headLen := min(n, 32)
				log.Printf("network read n=%d head=%s", n, hex.EncodeToString(buf[:headLen]))
			}
			packets, frameErr := c.framer.Push(buf[:n])
			c.mu.Lock()
			c.packets = append(c.packets, packets...)
			if frameErr != nil {
				c.errs = append(c.errs, frameErr)
			}
			c.mu.Unlock()
		}
		if err != nil {
			if err != io.EOF && c.isCurrentConn(conn) {
				c.addError(err)
			}
			c.clearConn(conn)
			return
		}
	}
}

func (c *Client) isCurrentConn(conn net.Conn) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn == conn
}

func (c *Client) clearConn(conn net.Conn) {
	c.mu.Lock()
	if c.conn == conn {
		c.conn = nil
		c.status = "offline"
	}
	c.mu.Unlock()
	_ = conn.Close()
}

func (c *Client) setStatus(status string) {
	c.mu.Lock()
	c.status = status
	c.mu.Unlock()
}

func (c *Client) addError(err error) {
	c.mu.Lock()
	c.errs = append(c.errs, err)
	c.mu.Unlock()
}
