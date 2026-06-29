package gamemode

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"strings"
	"time"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
)

type LoginMode struct {
	selected      int
	status        string
	packets       []string
	console       chatConsole
	autoAttempted bool
	enterWorld    bool
	username      string
	password      string
	focus         loginInputField
	background    *render.Image
	bgTiles       []*render.Image
	bgSource      string
	bgLoaded      bool
	bgmStarted    bool
}

type loginInputField int

const (
	loginFieldUser loginInputField = iota
	loginFieldPassword
)

func NewLoginMode() *LoginMode {
	return &LoginMode{status: "select a server", focus: loginFieldUser}
}

func (m *LoginMode) Name() string {
	return "login"
}

func (m *LoginMode) Enter(ctx Context) {
	if m.username == "" {
		m.username = ctx.Config.Login.Username
	}
	if m.password == "" {
		m.password = ctx.Config.Login.Password
	}
	m.loadBackground(ctx)
	m.playLoginBGM(ctx)
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

	m.updateFormInput(ctx)

	if ctx.Input.JustPressed(render.KeyArrowDown) {
		m.selected = (m.selected + 1) % len(conns)
	}
	if ctx.Input.JustPressed(render.KeyArrowUp) {
		m.selected = (m.selected + len(conns) - 1) % len(conns)
	}
	if ctx.Input.JustPressed(render.KeyEnter) {
		m.connectAndMaybeLogin(ctx, conns[m.selected])
	}
	if ctx.Input.JustPressed(render.KeyEscape) {
		ctx.Network.Close()
		m.status = "offline"
	}

	for _, pkt := range ctx.Network.DrainPackets() {
		log.Printf("recv packet 0x%04X len=%d", pkt.ID, len(pkt.Data))
		m.packets = append(m.packets, pkt.String())
		if chat, ok, err := network.ParseChatMessage(pkt); err != nil {
			m.packets = append(m.packets, "parse chat message: "+err.Error())
		} else if ok {
			addConsoleMessage(&m.console, ctx.Resources, chat)
			continue
		}
		if change, ok, err := network.ParseMapChange(pkt); err != nil {
			m.packets = append(m.packets, "parse ZC_NPCACK_MAPMOVE: "+err.Error())
		} else if ok {
			ctx.World.MapName = change.MapName
			ctx.Session.Zone.MapName = change.MapName
			applyWarpPosition(ctx, change.X, change.Y)
			ctx.Session.Playing = true
			m.status = fmt.Sprintf("map change: %s at %d,%d", change.MapName, change.X, change.Y)
			log.Printf("login map change map=%s x=%d y=%d server_move=%t addr=%s port=%d", change.MapName, change.X, change.Y, change.ServerMove, change.Address, change.Port)
			if change.ServerMove {
				ctx.Session.Zone.Address = change.Address
				ctx.Session.Zone.Port = change.Port
				m.connectMapServer(ctx, network.ZoneServerNotify{
					CharID:  ctx.Session.CharID,
					MapName: change.MapName,
					Address: change.Address,
					Port:    change.Port,
				})
			} else {
				m.enterWorld = true
			}
			continue
		}
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
						setSelectedCharacter(ctx.Session, convertCharacter(list.Characters[0]))
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
				applyMapAcceptEnter(ctx, enter)
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
		next := NewWorldMode()
		next.console = m.console
		return next, nil
	}
	return nil, nil
}

func (m *LoginMode) Draw(ctx Context, screen *render.Image) {
	m.drawBackground(ctx, screen)
	m.drawLoginWindow(ctx, screen)
}

func (m *LoginMode) updateFormInput(ctx Context) {
	if ctx.Input == nil {
		return
	}
	if ctx.Input.JustPressed(render.KeyTab) {
		if m.focus == loginFieldUser {
			m.focus = loginFieldPassword
		} else {
			m.focus = loginFieldUser
		}
	}
	if ctx.Input.JustPressed(render.KeyBackspace) {
		if m.focus == loginFieldPassword {
			m.password = trimLastRune(m.password)
		} else {
			m.username = trimLastRune(m.username)
		}
	}
	if text := ctx.Input.TextInput(); text != "" {
		if m.focus == loginFieldPassword {
			m.password += text
		} else {
			m.username += text
		}
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return
	}
	winX, winY, winW, _ := loginWindowRect(ctx)
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	userX, userY, userW, userH := loginUserFieldRect(winX, winY, winW)
	passX, passY, passW, passH := loginPasswordFieldRect(winX, winY, winW)
	buttonX, buttonY, buttonW, buttonH := loginButtonRect(winX, winY, winW)
	if pointInRect(mx, my, userX, userY, userW, userH) {
		m.focus = loginFieldUser
		return
	}
	if pointInRect(mx, my, passX, passY, passW, passH) {
		m.focus = loginFieldPassword
		return
	}
	if pointInRect(mx, my, buttonX, buttonY, buttonW, buttonH) && len(ctx.Resources.ClientInfo.Connections) > 0 {
		m.connectAndMaybeLogin(ctx, ctx.Resources.ClientInfo.Connections[m.selected])
		return
	}
	serverY := winY + 103
	for i := range ctx.Resources.ClientInfo.Connections {
		if pointInRect(mx, my, winX+22, serverY+i*17, winW-44, 16) {
			m.selected = i
			return
		}
	}
}

func (m *LoginMode) drawBackground(ctx Context, screen *render.Image) {
	clear(screen)
	width, height := ctx.ScreenSize()
	if width <= 0 || height <= 0 {
		return
	}
	if len(m.bgTiles) == 12 {
		cellW := float64(width) / 4
		cellH := float64(height) / 3
		for i, tile := range m.bgTiles {
			if tile == nil {
				continue
			}
			b := tile.Bounds()
			if b.Dx() <= 0 || b.Dy() <= 0 {
				continue
			}
			var opts render.DrawImageOptions
			opts.GeoM.Scale(cellW/float64(b.Dx()), cellH/float64(b.Dy()))
			opts.GeoM.Translate(float64(i%4)*cellW, float64(i/4)*cellH)
			opts.Filter = render.FilterLinear
			screen.DrawImage(tile, &opts)
		}
		return
	}
	if m.background == nil {
		render.DrawRect(screen, 0, 0, float64(width), float64(height), color.RGBA{R: 10, G: 13, B: 22, A: 255})
		return
	}
	b := m.background.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Scale(float64(width)/float64(b.Dx()), float64(height)/float64(b.Dy()))
	opts.Filter = render.FilterLinear
	screen.DrawImage(m.background, &opts)
}

func (m *LoginMode) drawLoginWindow(ctx Context, screen *render.Image) {
	x, y, w, h := loginWindowRect(ctx)
	drawUIWindowFrame(screen, x, y, w, h)
	drawUIRowSurface(screen, x+1, y+1, w-2, 20, color.RGBA{R: 44, G: 49, B: 60, A: 240})
	render.DebugPrintAtColor(screen, "Ragnarok Online", x+10, y+4, color.RGBA{R: 235, G: 226, B: 194, A: 255})

	labelColor := color.RGBA{R: 225, G: 219, B: 204, A: 255}
	mutedColor := color.RGBA{R: 182, G: 187, B: 197, A: 255}
	userX, userY, userW, userH := loginUserFieldRect(x, y, w)
	passX, passY, passW, passH := loginPasswordFieldRect(x, y, w)
	render.DebugPrintAtColor(screen, "Account", x+24, userY-17, labelColor)
	render.DebugPrintAtColor(screen, "Password", x+24, passY-17, labelColor)
	drawLoginInput(screen, userX, userY, userW, userH, m.username, m.focus == loginFieldUser)
	drawLoginInput(screen, passX, passY, passW, passH, strings.Repeat("*", len([]rune(m.password))), m.focus == loginFieldPassword)

	serverY := y + 103
	render.DebugPrintAtColor(screen, "Server", x+24, serverY-17, labelColor)
	for i, conn := range ctx.Resources.ClientInfo.Connections {
		rowY := serverY + i*17
		bg := color.RGBA{R: 20, G: 23, B: 31, A: 130}
		if i == m.selected {
			bg = color.RGBA{R: 74, G: 88, B: 118, A: 205}
		}
		drawUIRowSurface(screen, x+22, rowY, w-44, 16, bg)
		render.DebugPrintAtColor(screen, trimRunes(conn.Display, 22), x+28, rowY+1, labelColor)
		render.DebugPrintAtColor(screen, fmt.Sprintf("%s:%d", conn.Address, conn.Port), x+180, rowY+1, mutedColor)
	}

	buttonX, buttonY, buttonW, buttonH := loginButtonRect(x, y, w)
	buttonBG := color.RGBA{R: 72, G: 78, B: 92, A: 235}
	if ctx.Input != nil && pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, buttonX, buttonY, buttonW, buttonH) {
		buttonBG = color.RGBA{R: 92, G: 101, B: 121, A: 245}
	}
	drawUIButtonSurface(screen, buttonX, buttonY, buttonW, buttonH, buttonBG)
	render.DebugPrintAtColor(screen, "Login", buttonX+34, buttonY+4, labelColor)
	render.DebugPrintAtColor(screen, trimRunes(m.status, 48), x+14, y+h-20, mutedColor)
}

func drawLoginInput(screen *render.Image, x, y, w, h int, text string, focused bool) {
	bg := color.RGBA{R: 236, G: 232, B: 220, A: 238}
	border := color.RGBA{R: 90, G: 94, B: 108, A: 220}
	if focused {
		border = color.RGBA{R: 240, G: 216, B: 126, A: 255}
	}
	drawUISurface(screen, x, y, w, h, bg, border)
	render.DebugPrintAtColor(screen, trimRunes(text, maxInt(1, (w-14)/7)), x+6, y+4, color.RGBA{R: 36, G: 36, B: 39, A: 255})
}

func (m *LoginMode) loadBackground(ctx Context) {
	if m.bgLoaded {
		return
	}
	m.bgLoaded = true
	for _, set := range loginBackgroundSets(ctx.Config.Packet.ClientDate) {
		if len(set) == 1 {
			img, source, ok := loadLoginBackgroundImage(ctx.Resources, set[0])
			if ok {
				m.background = img
				m.bgSource = source
				return
			}
			continue
		}
		tiles := make([]*render.Image, 0, len(set))
		sources := make([]string, 0, len(set))
		ok := true
		for _, name := range set {
			img, source, loaded := loadLoginBackgroundImage(ctx.Resources, name)
			if !loaded {
				ok = false
				break
			}
			tiles = append(tiles, img)
			sources = append(sources, source)
		}
		if ok {
			m.bgTiles = tiles
			m.bgSource = fmt.Sprintf("%d login tiles", len(sources))
			return
		}
	}
	m.bgSource = "fallback"
}

func (m *LoginMode) playLoginBGM(ctx Context) {
	if m.bgmStarted || ctx.Audio == nil {
		return
	}
	m.bgmStarted = true
	for _, path := range []string{"01.mp3", "BGM\\01.mp3", "bgm\\01.mp3"} {
		if err := ctx.Audio.Play(path); err == nil {
			return
		}
	}
}

func loadLoginBackgroundImage(manager *res.Manager, name string) (*render.Image, string, bool) {
	for _, candidate := range loginInterfaceCandidates(name) {
		img, source, err := res.LoadImageExact(manager, []string{candidate})
		if err == nil {
			return render.NewImageFromImage(img), source, true
		}
	}
	return nil, "", false
}

func loginWindowRect(ctx Context) (int, int, int, int) {
	width, height := ctx.ScreenSize()
	w, h := 380, 202
	x := (width - w) / 2
	y := height - h - 82
	if y < 48 {
		y = (height - h) / 2
	}
	if x < 8 {
		x = 8
	}
	if y < 8 {
		y = 8
	}
	return x, y, w, h
}

func loginUserFieldRect(x, y, w int) (int, int, int, int) {
	return x + 110, y + 39, w - 135, 22
}

func loginPasswordFieldRect(x, y, w int) (int, int, int, int) {
	return x + 110, y + 72, w - 135, 22
}

func loginButtonRect(x, y, w int) (int, int, int, int) {
	return x + w - 126, y + 162, 96, 24
}

func loginBackgroundSets(clientDate int) [][]string {
	tiles2018 := []string{
		"t_\xB9\xE8\xB0\xE61-1.bmp", "t_\xB9\xE8\xB0\xE61-2.bmp", "t_\xB9\xE8\xB0\xE61-3.bmp", "t_\xB9\xE8\xB0\xE61-4.bmp",
		"t_\xB9\xE8\xB0\xE62-1.bmp", "t_\xB9\xE8\xB0\xE62-2.bmp", "t_\xB9\xE8\xB0\xE62-3.bmp", "t_\xB9\xE8\xB0\xE62-4.bmp",
		"t_\xB9\xE8\xB0\xE63-1.bmp", "t_\xB9\xE8\xB0\xE63-2.bmp", "t_\xB9\xE8\xB0\xE63-3.bmp", "t_\xB9\xE8\xB0\xE63-4.bmp",
	}
	sets := make([][]string, 0, 3)
	if clientDate >= 20221207 {
		sets = append(sets, []string{"t_login.jpg"})
	}
	if clientDate >= 20181114 {
		sets = append(sets, tiles2018)
	}
	sets = append(sets, []string{"bgi_temp.bmp"}, []string{"t_login.jpg"}, tiles2018)
	return sets
}

func loginInterfaceCandidates(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	const ui = "data\\texture\\\xC0\xAF\xC0\xFA\xC0\xCE\xC5\xCD\xC6\xE4\xC0\xCC\xBD\xBA\\"
	candidates := []string{
		ui + name,
		strings.ReplaceAll(ui, "\\", "/") + name,
		"texture\\\xC0\xAF\xC0\xFA\xC0\xCE\xC5\xCD\xC6\xE4\xC0\xCC\xBD\xBA\\" + name,
		"data\\texture\\interface\\" + name,
		"data/texture/interface/" + name,
		name,
	}
	return uniqueLoginStrings(candidates)
}

func uniqueLoginStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func trimLastRune(text string) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	return string(runes[:len(runes)-1])
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
	if strings.TrimSpace(m.username) == "" && m.password == "" {
		return
	}

	err = ctx.Network.SendAccountLogin(
		m.username,
		m.password,
		uint32(conn.Version),
		0,
	)
	if err != nil {
		m.status = "login packet failed: " + err.Error()
		return
	}
	m.status = "CA_LOGIN sent"
	log.Printf("sent CA_LOGIN user=%s version=%d", m.username, conn.Version)
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
		Money:     character.Money,
		Name:      character.Name,
		Slot:      character.Slot,
		Level:     character.Level,
		JobLevel:  character.JobLevel,
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
		HeadPal:   character.HeadPal,
		BodyPal:   character.BodyPal,
		Weapon:    character.Weapon,
		Shield:    character.Shield,
		HeadTop:   character.HeadTop,
		HeadMid:   character.HeadMid,
		HeadLow:   character.HeadLow,
	}
}

func setSelectedCharacter(sessionState *session.Session, character session.Character) {
	sessionState.Selected = character
	sessionState.Vitals = sessionVitalsFromCharacter(character)
	sessionState.Progress = sessionProgressFromCharacter(character)
	sessionState.Inventory.Zeny = character.Money
}

func describeConnection(conn res.Connection) string {
	return fmt.Sprintf("%s %s:%d v=%d lang=%d", conn.Display, conn.Address, conn.Port, conn.Version, conn.LangType)
}
