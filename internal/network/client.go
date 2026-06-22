package network

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Client struct {
	clientDate int
	framer     *Framer

	mu      sync.Mutex
	conn    net.Conn
	packets []Packet
	status  string
	errs    []error
}

func NewClient(clientDate int) *Client {
	return &Client{
		clientDate: clientDate,
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

	if os.Getenv("GORO_NET_TRACE") == "1" {
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

func (c *Client) SendLoadEndAck() error {
	return c.Send(BuildLoadEndAckPacket())
}

func (c *Client) SendTick(clientTick uint32) error {
	packet := BuildTickSendPacketForClientDate(clientTick, c.clientDate)
	err := c.Send(packet)
	if err == nil && os.Getenv("GORO_NET_TRACE") == "1" {
		log.Printf("sent CZ_REQUEST_TIME opcode=0x%04X tick=%d client_date=%d", ID(packet), clientTick, c.clientDate)
	} else if err != nil {
		log.Printf("send CZ_REQUEST_TIME failed opcode=0x%04X len=%d client_date=%d: %v", ID(packet), len(packet), c.clientDate, err)
	}
	return err
}

func (c *Client) SendNameRequest(gid uint32) error {
	return c.Send(BuildNameRequestPacket(gid))
}

func (c *Client) SendMapServerEnter(accountID, charID, authCode, clientTick uint32, sex uint8) error {
	packet := BuildMapServerEnterPacketForClientDate(MapServerEnter{
		AccountID:  accountID,
		CharID:     charID,
		AuthCode:   authCode,
		ClientTick: clientTick,
		Sex:        sex,
	}, c.clientDate)
	if os.Getenv("GORO_NET_TRACE") == "1" {
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
			if os.Getenv("GORO_NET_TRACE") == "1" {
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
			if err != io.EOF {
				c.addError(err)
			}
			c.clearConn(conn)
			return
		}
	}
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
