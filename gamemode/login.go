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
	selected       int
	phase          loginPhase
	status         string
	packets        []string
	console        chatConsole
	autoAttempted  bool
	fade           loginFadeState
	username       string
	password       string
	focus          loginInputField
	background     *render.Image
	bgTiles        []*render.Image
	bgSource       string
	bgLoaded       bool
	bgmStarted     bool
	selectedSlot   int
	maxSlots       int
	charViews      map[uint32]*humanoidSpriteView
	charViewFailed map[uint32]struct{}
	charWindow     *render.Image
	charBox        *render.Image
	cursor         roCursorState
}

type loginPhase int

const (
	loginPhaseAccount loginPhase = iota
	loginPhaseCharacter
)

type loginFadePhase int

const (
	loginFadeNone loginFadePhase = iota
	loginFadeOut
	loginFadeIn
)

type loginFadeState struct {
	phase      loginFadePhase
	started    time.Time
	target     loginPhase
	hasTarget  bool
	enterWorld bool
}

type loginInputField int

const (
	loginFieldUser loginInputField = iota
	loginFieldPassword
)

const (
	loginTransitionDuration    = 500 * time.Millisecond
	charSelectPreviewDirection = 4
	charSelectPreviewScale     = 0.92
	charSelectPreviewFeetLift  = 10
)

func NewLoginMode() *LoginMode {
	return &LoginMode{status: "select a server", focus: loginFieldUser, maxSlots: 9}
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
	m.loadCharacterSelectSkin(ctx)
	m.cursor.ensureLoaded(ctx)
	render.SetCursorMode(render.CursorModeHidden)
	m.playLoginBGM(ctx)
	if len(ctx.Resources.ClientInfo.Connections) == 0 {
		m.status = "no login servers discovered"
	}
}

func (m *LoginMode) Update(ctx Context) (Mode, error) {
	now := time.Now()
	if m.updateFade(now) {
		return m.nextWorldMode(now), nil
	}

	conns := ctx.Resources.ClientInfo.Connections
	if len(conns) == 0 {
		return nil, nil
	}

	if ctx.Config.Login.AutoLogin && !m.autoAttempted {
		m.autoAttempted = true
		m.connectAndMaybeLogin(ctx, conns[m.selected])
	}

	fading := m.fade.phase != loginFadeNone
	if !fading {
		if m.phase == loginPhaseCharacter {
			m.updateCharacterSelectInput(ctx)
		} else {
			m.updateFormInput(ctx)
		}

		if m.phase == loginPhaseAccount && ctx.Input.JustPressed(render.KeyArrowDown) {
			m.selected = (m.selected + 1) % len(conns)
		}
		if m.phase == loginPhaseAccount && ctx.Input.JustPressed(render.KeyArrowUp) {
			m.selected = (m.selected + len(conns) - 1) % len(conns)
		}
		if m.phase == loginPhaseAccount && ctx.Input.JustPressed(render.KeyEnter) {
			m.connectAndMaybeLogin(ctx, conns[m.selected])
		}
		if ctx.Input.JustPressed(render.KeyEscape) {
			if m.phase == loginPhaseCharacter {
				m.startPhaseFade(loginPhaseAccount, now)
				ctx.Network.Close()
				m.status = "char select cancelled"
			} else {
				ctx.Network.Close()
				m.status = "offline"
			}
		}
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
				m.startWorldFade(time.Now())
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
				m.maxSlots = 9
				m.selectedSlot = 0
				if len(list.Characters) > 0 {
					m.maxSlots = charSelectMaxSlots(ctx.Session.Characters)
					m.selectedSlot = firstOccupiedCharacterSlot(ctx.Session.Characters)
					m.status = "select a character"
				} else {
					m.status = "no characters"
				}
				m.startPhaseFade(loginPhaseCharacter, time.Now())
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
				m.startWorldFade(time.Now())
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

	if m.updateFade(time.Now()) {
		return m.nextWorldMode(time.Now()), nil
	}
	return nil, nil
}

func (m *LoginMode) Draw(ctx Context, screen *render.Image) {
	m.drawBackground(ctx, screen)
	if m.phase == loginPhaseCharacter {
		m.drawCharacterSelect(ctx, screen)
	} else {
		m.drawLoginWindow(ctx, screen)
	}
	now := time.Now()
	m.drawFade(ctx, screen, now)
	m.drawROCursor(screen, ctx, now)
}

func (m *LoginMode) drawROCursor(screen *render.Image, ctx Context, now time.Time) {
	if ctx.Input == nil {
		return
	}
	render.SetCursorMode(render.CursorModeHidden)
	m.cursor.draw(screen, ctx, m.cursorAction(ctx), now)
}

func (m *LoginMode) cursorAction(ctx Context) int {
	if ctx.Input == nil {
		return cursorActionDefault
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	if m.phase == loginPhaseCharacter {
		x, y, _, _ := charSelectWindowRect(ctx)
		for localSlot := 0; localSlot < 3; localSlot++ {
			slotX, slotY, slotW, slotH := charSelectSlotRect(x, y, localSlot)
			if pointInRect(mx, my, slotX, slotY, slotW, slotH) {
				return cursorActionClick
			}
		}
		for _, rect := range [][4]int{
			rectArray(charSelectLeftArrowRect(x, y)),
			rectArray(charSelectRightArrowRect(x, y)),
			rectArray(charSelectDeleteButtonRect(x, y)),
			rectArray(charSelectMakeButtonRect(x, y)),
			rectArray(charSelectOKButtonRect(x, y)),
			rectArray(charSelectCancelButtonRect(x, y)),
		} {
			if pointInRect(mx, my, rect[0], rect[1], rect[2], rect[3]) {
				return cursorActionClick
			}
		}
		return cursorActionDefault
	}
	winX, winY, winW, _ := loginWindowRect(ctx)
	userX, userY, userW, userH := loginUserFieldRect(winX, winY, winW)
	passX, passY, passW, passH := loginPasswordFieldRect(winX, winY, winW)
	buttonX, buttonY, buttonW, buttonH := loginButtonRect(winX, winY, winW)
	if pointInRect(mx, my, userX, userY, userW, userH) ||
		pointInRect(mx, my, passX, passY, passW, passH) ||
		pointInRect(mx, my, buttonX, buttonY, buttonW, buttonH) {
		return cursorActionClick
	}
	serverY := winY + 103
	for i := range ctx.Resources.ClientInfo.Connections {
		if pointInRect(mx, my, winX+22, serverY+i*17, winW-44, 16) {
			return cursorActionClick
		}
	}
	return cursorActionDefault
}

func rectArray(x, y, w, h int) [4]int {
	return [4]int{x, y, w, h}
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

func (m *LoginMode) updateCharacterSelectInput(ctx Context) {
	if ctx.Input == nil {
		return
	}
	if ctx.Input.JustPressed(render.KeyArrowLeft) {
		m.moveSelectedSlot(-1)
	}
	if ctx.Input.JustPressed(render.KeyArrowRight) {
		m.moveSelectedSlot(1)
	}
	if ctx.Input.JustPressed(render.KeyEnter) {
		m.submitSelectedCharacter(ctx)
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	x, y, _, _ := charSelectWindowRect(ctx)
	for localSlot := 0; localSlot < 3; localSlot++ {
		slotX, slotY, slotW, slotH := charSelectSlotRect(x, y, localSlot)
		if pointInRect(mx, my, slotX, slotY, slotW, slotH) {
			clickedSlot := charSelectPage(m.selectedSlot)*3 + localSlot
			if clickedSlot == m.selectedSlot {
				if _, ok := characterBySlot(ctx.Session.Characters, clickedSlot); ok {
					m.submitSelectedCharacter(ctx)
				} else {
					m.status = "character creation is not implemented yet"
				}
				return
			}
			m.selectedSlot = clampCharacterSlot(clickedSlot, m.maxSlots)
			return
		}
	}
	leftX, leftY, leftW, leftH := charSelectLeftArrowRect(x, y)
	rightX, rightY, rightW, rightH := charSelectRightArrowRect(x, y)
	if pointInRect(mx, my, leftX, leftY, leftW, leftH) {
		m.moveSelectedSlot(-1)
		return
	}
	if pointInRect(mx, my, rightX, rightY, rightW, rightH) {
		m.moveSelectedSlot(1)
		return
	}
	okX, okY, okW, okH := charSelectOKButtonRect(x, y)
	cancelX, cancelY, cancelW, cancelH := charSelectCancelButtonRect(x, y)
	makeX, makeY, makeW, makeH := charSelectMakeButtonRect(x, y)
	deleteX, deleteY, deleteW, deleteH := charSelectDeleteButtonRect(x, y)
	switch {
	case pointInRect(mx, my, okX, okY, okW, okH):
		m.submitSelectedCharacter(ctx)
	case pointInRect(mx, my, cancelX, cancelY, cancelW, cancelH):
		m.startPhaseFade(loginPhaseAccount, time.Now())
		ctx.Network.Close()
		m.status = "char select cancelled"
	case pointInRect(mx, my, makeX, makeY, makeW, makeH):
		m.status = "character creation is not implemented yet"
	case pointInRect(mx, my, deleteX, deleteY, deleteW, deleteH):
		m.status = "character deletion is not implemented yet"
	}
}

func (m *LoginMode) startPhaseFade(target loginPhase, now time.Time) {
	if m.fade.phase != loginFadeNone && m.fade.hasTarget && m.fade.target == target && !m.fade.enterWorld {
		return
	}
	if m.phase == target && m.fade.phase == loginFadeNone {
		return
	}
	m.fade = loginFadeState{
		phase:     loginFadeOut,
		started:   now,
		target:    target,
		hasTarget: true,
	}
}

func (m *LoginMode) startWorldFade(now time.Time) {
	if m.fade.phase != loginFadeNone && m.fade.enterWorld {
		return
	}
	m.fade = loginFadeState{
		phase:      loginFadeOut,
		started:    now,
		enterWorld: true,
	}
}

func (m *LoginMode) updateFade(now time.Time) bool {
	switch m.fade.phase {
	case loginFadeOut:
		if now.Sub(m.fade.started) < loginTransitionDuration {
			return false
		}
		if m.fade.enterWorld {
			return true
		}
		if m.fade.hasTarget {
			m.phase = m.fade.target
		}
		m.fade = loginFadeState{phase: loginFadeIn, started: now}
	case loginFadeIn:
		if now.Sub(m.fade.started) >= loginTransitionDuration {
			m.fade = loginFadeState{}
		}
	}
	return false
}

func (m *LoginMode) fadeAlpha(now time.Time) uint8 {
	if m.fade.started.IsZero() {
		return 0
	}
	switch m.fade.phase {
	case loginFadeOut:
		return clampColor(255 * clampUnit(float64(now.Sub(m.fade.started))/float64(loginTransitionDuration)))
	case loginFadeIn:
		return clampColor(255 * (1 - clampUnit(float64(now.Sub(m.fade.started))/float64(loginTransitionDuration))))
	default:
		return 0
	}
}

func (m *LoginMode) drawFade(ctx Context, screen *render.Image, now time.Time) {
	alpha := m.fadeAlpha(now)
	if alpha == 0 {
		return
	}
	width, height := ctx.ScreenSize()
	render.DrawRect(screen, 0, 0, float64(width), float64(height), color.RGBA{A: alpha})
}

func (m *LoginMode) nextWorldMode(now time.Time) *WorldMode {
	next := NewWorldMode()
	next.console = m.console
	next.startMapFadeIn(now)
	return next
}

func (m *LoginMode) moveSelectedSlot(delta int) {
	m.selectedSlot = clampCharacterSlot(m.selectedSlot+delta, m.maxSlots)
}

func (m *LoginMode) submitSelectedCharacter(ctx Context) {
	character, ok := characterBySlot(ctx.Session.Characters, m.selectedSlot)
	if !ok {
		m.status = "empty character slot"
		return
	}
	if err := ctx.Network.SendSelectCharacter(character.Slot); err != nil {
		m.status = "select character failed: " + err.Error()
		return
	}
	ctx.Session.CharID = character.ID
	setSelectedCharacter(ctx.Session, character)
	m.status = fmt.Sprintf("selected character %s", character.Name)
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
	drawUITitledWindowFrame(screen, x, y, w, h, 21)
	render.DebugPrintAtColor(screen, "Ragnarok Online", x+10, y+4, uiTitleTextColor)

	labelColor := uiTextColor
	mutedColor := uiMutedTextColor
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
		bg := uiPanelAltColor
		if i == m.selected {
			bg = uiSelectionColor
		}
		drawUIRowSurface(screen, x+22, rowY, w-44, 16, bg)
		render.DebugPrintAtColor(screen, trimRunes(conn.Display, 22), x+28, rowY+1, labelColor)
		render.DebugPrintAtColor(screen, fmt.Sprintf("%s:%d", conn.Address, conn.Port), x+180, rowY+1, mutedColor)
	}

	buttonX, buttonY, buttonW, buttonH := loginButtonRect(x, y, w)
	buttonBG := uiButtonColor
	if ctx.Input != nil && pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, buttonX, buttonY, buttonW, buttonH) {
		buttonBG = uiButtonHoverColor
	}
	drawUIButtonSurface(screen, buttonX, buttonY, buttonW, buttonH, buttonBG)
	render.DebugPrintAtColor(screen, "Login", buttonX+34, buttonY+4, labelColor)
	render.DebugPrintAtColor(screen, trimRunes(m.status, 48), x+14, y+h-20, mutedColor)
}

func (m *LoginMode) drawCharacterSelect(ctx Context, screen *render.Image) {
	x, y, w, h := charSelectWindowRect(ctx)
	drawUITitledWindowFrame(screen, x, y, w, h, 23)
	render.DebugPrintAtColor(screen, "Select Character", x+12, y+5, uiTitleTextColor)

	page := charSelectPage(m.selectedSlot)
	pageStart := page * 3
	for localSlot := 0; localSlot < 3; localSlot++ {
		slot := pageStart + localSlot
		slotX, slotY, slotW, slotH := charSelectSlotRect(x, y, localSlot)
		selected := slot == m.selectedSlot
		bg := uiPanelBodyColor
		border := uiWindowBorderColor
		if selected {
			bg = uiSelectionColor
			border = uiSelectionBorder
		}
		drawUISurface(screen, slotX, slotY, slotW, slotH, bg, border)
		if character, ok := characterBySlot(ctx.Session.Characters, slot); ok {
			m.drawCharacterPreview(screen, ctx, character, slotX+slotW/2, slotY+slotH-15-charSelectPreviewFeetLift)
			render.DrawOutlinedTextAt(screen, trimRunes(character.Name, 16), slotX+8, slotY+slotH-18, uiTextColor, color.RGBA{A: 160})
		} else {
			render.DebugPrintAtColor(screen, "Create", slotX+45, slotY+58, uiMutedTextColor)
		}
	}

	leftX, leftY, leftW, leftH := charSelectLeftArrowRect(x, y)
	rightX, rightY, rightW, rightH := charSelectRightArrowRect(x, y)
	drawCharSelectArrow(screen, leftX, leftY, leftW, leftH, "<")
	drawCharSelectArrow(screen, rightX, rightY, rightW, rightH, ">")

	m.drawSelectedCharacterInfo(screen, ctx, x, y)
	m.drawCharacterSelectFooter(screen, ctx, x, y, w, h)
}

func (m *LoginMode) drawCharacterPreview(screen *render.Image, ctx Context, character session.Character, centerX, feetY int) {
	view := m.characterPreviewView(ctx, character)
	if view == nil {
		render.DebugPrintAtColor(screen, "?", centerX-3, feetY-72, color.RGBA{R: 220, G: 220, B: 220, A: 255})
		return
	}
	billboard, ok := humanoidBillboardForState(view, spriteState{
		actionFamily: spriteActionIdle,
		direction:    charSelectPreviewDirection,
		started:      time.Now(),
		loopIdle:     true,
	}, time.Now())
	if !ok || billboard == nil || billboard.image == nil {
		return
	}
	var opts render.DrawImageOptions
	scale := charSelectPreviewScale
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(float64(centerX)-billboard.anchorX*scale, float64(feetY)-billboard.anchorY*scale)
	opts.Filter = spriteDrawFilter()
	screen.DrawImage(billboard.image, &opts)
}

func (m *LoginMode) characterPreviewView(ctx Context, character session.Character) *humanoidSpriteView {
	if character.ID == 0 {
		return nil
	}
	if _, failed := m.charViewFailed[character.ID]; failed {
		return nil
	}
	if m.charViews == nil {
		m.charViews = make(map[uint32]*humanoidSpriteView)
	}
	if view := m.charViews[character.ID]; view != nil {
		return view
	}
	view, status := loadPlayerHumanoidSpriteView(ctx.Resources, character, ctx.Session.Sex)
	if view == nil {
		if m.charViewFailed == nil {
			m.charViewFailed = make(map[uint32]struct{})
		}
		m.charViewFailed[character.ID] = struct{}{}
		log.Printf("char select sprite resources char_id=%d name=%s job=%d %s", character.ID, character.Name, character.Job, status)
		return nil
	}
	m.charViews[character.ID] = view
	return view
}

func (m *LoginMode) drawSelectedCharacterInfo(screen *render.Image, ctx Context, x, y int) {
	character, ok := characterBySlot(ctx.Session.Characters, m.selectedSlot)
	panelX, panelY, panelW, panelH := x+16, y+204, 318, 108
	drawUIPanelSurface(screen, panelX, panelY, panelW, panelH, uiPanelBodyColor)
	text := uiTextColor
	if !ok {
		render.DebugPrintAtColor(screen, "Empty Slot", panelX+18, panelY+14, text)
		render.DebugPrintAtColor(screen, "Use Make to create a character later.", panelX+18, panelY+34, text)
		return
	}
	render.DebugPrintAtColor(screen, trimRunes(character.Name, 24), panelX+14, panelY+10, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("Job: %s", trimRunes(characterJobName(character), 18)), panelX+14, panelY+28, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("Lv: %d / Job %d", character.Level, character.JobLevel), panelX+14, panelY+46, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("HP: %d / %d", character.HP, character.MaxHP), panelX+14, panelY+64, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("SP: %d / %d", character.SP, character.MaxSP), panelX+14, panelY+82, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("STR %d", character.Str), panelX+180, panelY+10, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("AGI %d", character.Agi), panelX+180, panelY+28, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("VIT %d", character.Vit), panelX+180, panelY+46, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("INT %d", character.Int), panelX+246, panelY+10, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("DEX %d", character.Dex), panelX+246, panelY+28, text)
	render.DebugPrintAtColor(screen, fmt.Sprintf("LUK %d", character.Luk), panelX+246, panelY+46, text)
}

func (m *LoginMode) drawCharacterSelectFooter(screen *render.Image, ctx Context, x, y, w, h int) {
	page := charSelectPage(m.selectedSlot)
	pageCount := maxInt(1, (m.maxSlots+2)/3)
	statusColor := uiMutedTextColor
	labelColor := uiTextColor
	render.DebugPrintAtColor(screen, fmt.Sprintf("%d / %d", len(ctx.Session.Characters), m.maxSlots), x+w-112, y+198, statusColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("%d / %d", page+1, pageCount), x+w/2-18, y+190, statusColor)
	render.DebugPrintAtColor(screen, trimRunes(m.status, 42), x+12, y+h-22, statusColor)

	deleteX, deleteY, deleteW, deleteH := charSelectDeleteButtonRect(x, y)
	makeX, makeY, makeW, makeH := charSelectMakeButtonRect(x, y)
	okX, okY, okW, okH := charSelectOKButtonRect(x, y)
	cancelX, cancelY, cancelW, cancelH := charSelectCancelButtonRect(x, y)
	drawCharSelectButton(screen, ctx, deleteX, deleteY, deleteW, deleteH, "Delete", labelColor)
	drawCharSelectButton(screen, ctx, makeX, makeY, makeW, makeH, "Make", labelColor)
	drawCharSelectButton(screen, ctx, okX, okY, okW, okH, "OK", labelColor)
	drawCharSelectButton(screen, ctx, cancelX, cancelY, cancelW, cancelH, "Cancel", labelColor)
}

func drawCharSelectButton(screen *render.Image, ctx Context, x, y, w, h int, label string, textColor color.RGBA) {
	bg := uiButtonColor
	if ctx.Input != nil && pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, x, y, w, h) {
		bg = uiButtonHoverColor
	}
	drawUIButtonSurface(screen, x, y, w, h, bg)
	render.DebugPrintAtColor(screen, label, x+(w-len(label)*7)/2, y+4, textColor)
}

func drawCharSelectArrow(screen *render.Image, x, y, w, h int, label string) {
	drawUIButtonSurface(screen, x, y, w, h, uiButtonColor)
	render.DebugPrintAtColor(screen, label, x+5, y+1, uiTextColor)
}

func drawLoginInput(screen *render.Image, x, y, w, h int, text string, focused bool) {
	bg := uiPanelBodyColor
	border := uiButtonBorderColor
	if focused {
		border = uiSelectionBorder
	}
	drawUISurface(screen, x, y, w, h, bg, border)
	render.DebugPrintAtColor(screen, trimRunes(text, maxInt(1, (w-14)/7)), x+6, y+4, uiTextColor)
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

func (m *LoginMode) loadCharacterSelectSkin(ctx Context) {
	if m.charWindow == nil {
		if img, _, ok := loadLoginBackgroundImage(ctx.Resources, "login_interface/win_select.bmp"); ok {
			m.charWindow = img
		}
	}
	if m.charBox == nil {
		if img, _, ok := loadLoginBackgroundImage(ctx.Resources, "login_interface/box_select.bmp"); ok {
			m.charBox = img
		}
	}
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
	y := (height*2)/3 - h/2
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

func charSelectWindowRect(ctx Context) (int, int, int, int) {
	width, height := ctx.ScreenSize()
	w, h := 576, 342
	x := (width - w) / 2
	y := (height - h) / 2
	if x < 8 {
		x = 8
	}
	if y < 8 {
		y = 8
	}
	return x, y, w, h
}

func charSelectSlotRect(x, y, localSlot int) (int, int, int, int) {
	lefts := [3]int{60, 224, 386}
	if localSlot < 0 || localSlot >= len(lefts) {
		localSlot = 0
	}
	return x + lefts[localSlot] - 5, y + 40, 139, 144
}

func charSelectLeftArrowRect(x, y int) (int, int, int, int) {
	return x + 40, y + 105, 18, 18
}

func charSelectRightArrowRect(x, y int) (int, int, int, int) {
	return x + 518, y + 105, 18, 18
}

func charSelectDeleteButtonRect(x, y int) (int, int, int, int) {
	return x + 4, y + 318, 58, 20
}

func charSelectMakeButtonRect(x, y int) (int, int, int, int) {
	return x + 434, y + 318, 42, 20
}

func charSelectOKButtonRect(x, y int) (int, int, int, int) {
	return x + 484, y + 318, 42, 20
}

func charSelectCancelButtonRect(x, y int) (int, int, int, int) {
	return x + 530, y + 318, 42, 20
}

func charSelectPage(slot int) int {
	if slot < 0 {
		return 0
	}
	return slot / 3
}

func charSelectMaxSlots(characters []session.Character) int {
	maxSlots := 9
	for _, character := range characters {
		if int(character.Slot)+1 > maxSlots {
			maxSlots = int(character.Slot) + 1
		}
	}
	if maxSlots%3 != 0 {
		maxSlots += 3 - maxSlots%3
	}
	return maxSlots
}

func firstOccupiedCharacterSlot(characters []session.Character) int {
	if len(characters) == 0 {
		return 0
	}
	slot := int(characters[0].Slot)
	for _, character := range characters[1:] {
		if int(character.Slot) < slot {
			slot = int(character.Slot)
		}
	}
	return slot
}

func clampCharacterSlot(slot, maxSlots int) int {
	if maxSlots <= 0 {
		maxSlots = 1
	}
	if slot < 0 {
		return 0
	}
	if slot >= maxSlots {
		return maxSlots - 1
	}
	return slot
}

func characterBySlot(characters []session.Character, slot int) (session.Character, bool) {
	for _, character := range characters {
		if int(character.Slot) == slot {
			return character, true
		}
	}
	return session.Character{}, false
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
