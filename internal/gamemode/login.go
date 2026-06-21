package gamemode

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/res"
	"github.com/kivutar/goro/internal/session"
)

type LoginMode struct {
	selected      int
	status        string
	packets       []string
	autoAttempted bool
	enterWorld    bool
}

func NewLoginMode() *LoginMode {
	return &LoginMode{status: "select a server"}
}

func (m *LoginMode) Name() string {
	return "login"
}

func (m *LoginMode) Enter(ctx Context) {
	if len(ctx.Resources.ClientInfo.Connections) == 0 {
		m.status = "no login servers discovered"
	}
}

func (m *LoginMode) Update(ctx Context) (Mode, error) {
	conns := ctx.Resources.ClientInfo.Connections
	if len(conns) == 0 {
		return nil, nil
	}

	if ctx.Config.Login.AutoLogin && !m.autoAttempted {
		m.autoAttempted = true
		m.connectAndMaybeLogin(ctx, conns[m.selected])
	}

	if ctx.Input.JustPressed(ebiten.KeyArrowDown) || ctx.Input.JustPressed(ebiten.KeyTab) {
		m.selected = (m.selected + 1) % len(conns)
	}
	if ctx.Input.JustPressed(ebiten.KeyArrowUp) {
		m.selected = (m.selected + len(conns) - 1) % len(conns)
	}
	if ctx.Input.JustPressed(ebiten.KeyEnter) {
		m.connectAndMaybeLogin(ctx, conns[m.selected])
	}
	if ctx.Input.JustPressed(ebiten.KeyEscape) {
		ctx.Network.Close()
		m.status = "offline"
	}

	for _, pkt := range ctx.Network.DrainPackets() {
		log.Printf("recv packet 0x%04X len=%d", pkt.ID, len(pkt.Data))
		m.packets = append(m.packets, pkt.String())
		if pkt.ID == 0x0069 {
			login, err := network.ParseAccountAcceptLogin(pkt)
			if err != nil {
				m.packets = append(m.packets, "parse AC_ACCEPT_LOGIN: "+err.Error())
			} else {
				ctx.Session.AccountID = login.AccountID
				ctx.Session.AuthCode = login.AuthCode
				ctx.Session.UserLevel = login.UserLevel
				ctx.Session.Sex = login.Sex
				ctx.Session.CharServers = convertCharServers(login.CharServer)
				m.status = fmt.Sprintf("account accepted: aid=%d char_servers=%d", login.AccountID, len(login.CharServer))
				log.Printf("account accepted aid=%d sex=%d char_servers=%d", login.AccountID, login.Sex, len(login.CharServer))
				for _, server := range login.CharServer {
					m.packets = append(m.packets, fmt.Sprintf("char %s %s:%d users=%d", server.Name, server.Address, server.Port, server.UserCount))
				}
				if len(login.CharServer) > 0 {
					m.connectCharServer(ctx, login.CharServer[0])
				}
			}
		}
		if pkt.ID == 0x006B {
			list, err := network.ParseCharList(pkt)
			if err != nil {
				m.packets = append(m.packets, "parse HC_ACCEPT_ENTER: "+err.Error())
			} else {
				ctx.Session.Characters = convertCharacters(list.Characters)
				m.status = fmt.Sprintf("char list: %d characters (%s)", len(list.Characters), list.Layout)
				log.Printf("char list characters=%d layout=%s", len(list.Characters), list.Layout)
				for _, character := range list.Characters {
					m.packets = append(m.packets, fmt.Sprintf("char slot=%d gid=%d name=%s lv=%d job=%d", character.Slot, character.ID, character.Name, character.Level, character.Job))
				}
				if len(list.Characters) > 0 {
					if err := ctx.Network.SendSelectCharacter(list.Characters[0].Slot); err != nil {
						m.status = "select character failed: " + err.Error()
					} else {
						ctx.Session.CharID = list.Characters[0].ID
						ctx.Session.Selected = convertCharacter(list.Characters[0])
						m.status = fmt.Sprintf("selected character %s", list.Characters[0].Name)
					}
				}
			}
		}
		if pkt.ID == 0x0071 {
			zone, err := network.ParseZoneServerNotify(pkt)
			if err != nil {
				m.packets = append(m.packets, "parse HC_NOTIFY_ZONESVR: "+err.Error())
			} else {
				ctx.Session.CharID = zone.CharID
				ctx.Session.Zone = session.ZoneServer{
					Address: zone.Address,
					Port:    zone.Port,
					MapName: zone.MapName,
				}
				ctx.World.MapName = zone.MapName
				m.status = fmt.Sprintf("zone server: %s %s:%d", zone.MapName, zone.Address, zone.Port)
				log.Printf("zone server map=%s addr=%s port=%d char_id=%d", zone.MapName, zone.Address, zone.Port, zone.CharID)
				m.connectMapServer(ctx, zone)
			}
		}
		if pkt.ID == 0x0073 || pkt.ID == 0x02EB {
			enter, err := network.ParseMapAcceptEnter(pkt)
			if err != nil {
				m.packets = append(m.packets, "parse ZC_ACCEPT_ENTER: "+err.Error())
			} else {
				ctx.Session.PlayerX = enter.X
				ctx.Session.PlayerY = enter.Y
				ctx.Session.PlayerDir = enter.Dir
				ctx.Session.Playing = true
				ctx.World.SetPlayerPosition(enter.X, enter.Y, enter.Dir)
				m.status = fmt.Sprintf("entered map %s at %d,%d dir=%d tick=%d", ctx.World.MapName, enter.X, enter.Y, enter.Dir, enter.ServerTick)
				log.Printf("entered map=%s x=%d y=%d dir=%d tick=%d", ctx.World.MapName, enter.X, enter.Y, enter.Dir, enter.ServerTick)
				m.enterWorld = true
			}
		}
		if entry, ok, err := network.ParseActorEntry(pkt); err != nil {
			m.packets = append(m.packets, "parse actor entry: "+err.Error())
		} else if ok {
			upsertNetworkActor(ctx, entry)
		}
		if len(m.packets) > 8 {
			m.packets = m.packets[len(m.packets)-8:]
		}
	}
	for _, err := range ctx.Network.DrainErrors() {
		log.Printf("network frame error: %v", err)
		m.packets = append(m.packets, "frame error: "+err.Error())
		if len(m.packets) > 8 {
			m.packets = m.packets[len(m.packets)-8:]
		}
	}

	if m.enterWorld {
		return NewWorldMode(), nil
	}
	return nil, nil
}

func (m *LoginMode) Draw(ctx Context, screen *ebiten.Image) {
	clear(screen)
	drawPanel(screen, 32, 32, 560, 420)
	debugText(screen, 52, 52, "Login")
	debugText(screen, 52, 76, "status: %s", m.status)
	debugText(screen, 52, 96, "network: %s", ctx.Network.Status())
	if ctx.Config.Login.Username != "" {
		debugText(screen, 52, 116, "env login: %s", ctx.Config.Login.Username)
	}

	y := 150
	for i, conn := range ctx.Resources.ClientInfo.Connections {
		prefix := " "
		if i == m.selected {
			prefix = ">"
		}
		debugText(screen, 52, y, "%s %s", prefix, describeConnection(conn))
		y += 18
	}

	y += 16
	debugText(screen, 52, y, "recent packets")
	y += 20
	for _, line := range m.packets {
		debugText(screen, 52, y, "%s", line)
		y += 18
	}
}

func (m *LoginMode) connectAndMaybeLogin(ctx Context, conn res.Connection) {
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err := ctx.Network.Connect(dialCtx, conn.Address, conn.Port)
	cancel()
	if err != nil {
		m.status = err.Error()
		return
	}

	m.status = ctx.Network.Status()
	log.Printf("connected login server %s:%d", conn.Address, conn.Port)
	if ctx.Config.Login.Username == "" && ctx.Config.Login.Password == "" {
		return
	}

	err = ctx.Network.SendAccountLogin(
		ctx.Config.Login.Username,
		ctx.Config.Login.Password,
		uint32(conn.Version),
		0,
	)
	if err != nil {
		m.status = "login packet failed: " + err.Error()
		return
	}
	m.status = "CA_LOGIN sent"
	log.Printf("sent CA_LOGIN user=%s version=%d", ctx.Config.Login.Username, conn.Version)
}

func (m *LoginMode) connectCharServer(ctx Context, server network.CharServer) {
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err := ctx.Network.Connect(dialCtx, server.Address, int(server.Port))
	cancel()
	if err != nil {
		m.status = "char connect failed: " + err.Error()
		return
	}

	err = ctx.Network.SendCharServerEnter(ctx.Session.AccountID, ctx.Session.AuthCode, ctx.Session.UserLevel, ctx.Session.Sex)
	if err != nil {
		m.status = "CA_ENTER failed: " + err.Error()
		return
	}
	m.status = "CA_ENTER sent to char server"
	log.Printf("sent CA_ENTER account_id=%d addr=%s port=%d", ctx.Session.AccountID, server.Address, server.Port)
}

func (m *LoginMode) connectMapServer(ctx Context, zone network.ZoneServerNotify) {
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err := ctx.Network.Connect(dialCtx, zone.Address, int(zone.Port))
	cancel()
	if err != nil {
		m.status = "map connect failed: " + err.Error()
		return
	}

	err = ctx.Network.SendMapServerEnter(ctx.Session.AccountID, zone.CharID, ctx.Session.AuthCode, uint32(time.Now().UnixMilli()), ctx.Session.Sex)
	if err != nil {
		m.status = "CZ_ENTER2 failed: " + err.Error()
		return
	}
	m.status = "CZ_ENTER2 sent to map server"
	log.Printf("sent CZ_ENTER2 account_id=%d char_id=%d addr=%s port=%d", ctx.Session.AccountID, zone.CharID, zone.Address, zone.Port)
}

func convertCharServers(servers []network.CharServer) []session.CharServer {
	out := make([]session.CharServer, 0, len(servers))
	for _, server := range servers {
		out = append(out, session.CharServer{
			Address:   server.Address,
			Port:      server.Port,
			Name:      server.Name,
			UserCount: server.UserCount,
			State:     server.State,
			Property:  server.Property,
		})
	}
	return out
}

func convertCharacters(characters []network.Character) []session.Character {
	out := make([]session.Character, 0, len(characters))
	for _, character := range characters {
		out = append(out, convertCharacter(character))
	}
	return out
}

func convertCharacter(character network.Character) session.Character {
	return session.Character{
		ID:        character.ID,
		Name:      character.Name,
		Slot:      character.Slot,
		Level:     character.Level,
		Job:       character.Job,
		HP:        character.HP,
		MaxHP:     character.MaxHP,
		SP:        character.SP,
		MaxSP:     character.MaxSP,
		Str:       character.Str,
		Agi:       character.Agi,
		Vit:       character.Vit,
		Int:       character.Int,
		Dex:       character.Dex,
		Luk:       character.Luk,
		Hair:      character.Hair,
		HairColor: character.HairColor,
	}
}

func describeConnection(conn res.Connection) string {
	return fmt.Sprintf("%s %s:%d v=%d lang=%d", conn.Display, conn.Address, conn.Port, conn.Version, conn.LangType)
}
