package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

const (
	shortcutSlots = 9
	shortcutSlot  = 34
	shortcutGap   = 2
	shortcutPad   = 3
)

type shortcutKind int

const (
	shortcutEmpty shortcutKind = iota
	shortcutItem
	shortcutSkill
)

type shortcutSlotState struct {
	kind       shortcutKind
	itemIndex  uint16
	itemID     uint16
	identified bool
	skillID    uint16
	skillLevel int
}

type ShortcutBar struct {
	slots      [shortcutSlots]shortcutSlotState
	status     string
	statusGood bool
	statusAt   time.Time
	loaded     bool
	path       string
}

type shortcutPersistFile struct {
	Version int                   `json:"version"`
	Slots   []shortcutPersistSlot `json:"slots"`
}

type shortcutPersistSlot struct {
	Kind       string `json:"kind,omitempty"`
	ItemIndex  uint16 `json:"item_index,omitempty"`
	ItemID     uint16 `json:"item_id,omitempty"`
	Identified bool   `json:"identified,omitempty"`
	SkillID    uint16 `json:"skill_id,omitempty"`
	SkillLevel int    `json:"skill_level,omitempty"`
}

func (b *ShortcutBar) Update(ctx Context, actions GameActions) bool {
	if ctx.Input == nil {
		return false
	}
	for i := 0; i < shortcutSlots; i++ {
		if ctx.Input.JustPressed(shortcutKey(i)) {
			b.activate(ctx, actions, i)
			return true
		}
	}
	slot, ok := b.slotAt(ctx, ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok {
		return false
	}
	if ctx.Input.MouseJustPressed(render.MouseButtonRight) {
		b.slots[slot] = shortcutSlotState{}
		b.setStatus(fmt.Sprintf("F%d cleared", slot+1), true)
		b.save(ctx)
		return true
	}
	if ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		b.activate(ctx, actions, slot)
		return true
	}
	return false
}

func (b *ShortcutBar) AcceptItemDrop(ctx Context, item session.InventoryItem, mx, my int) bool {
	slot, ok := b.slotAt(ctx, mx, my)
	if !ok {
		return false
	}
	b.slots[slot] = shortcutSlotState{
		kind:       shortcutItem,
		itemIndex:  item.Index,
		itemID:     item.ItemID,
		identified: item.Identified,
	}
	b.setStatus(fmt.Sprintf("%s assigned to F%d", trimRunes(inventoryItemDisplayName(ctx.Resources, item), 24), slot+1), true)
	b.save(ctx)
	return true
}

func (b *ShortcutBar) AcceptSkillDrop(ctx Context, skill session.Skill, mx, my int) bool {
	slot, ok := b.slotAt(ctx, mx, my)
	if !ok {
		return false
	}
	if skill.ID == 0 || skill.Level <= 0 {
		b.setStatus("Skill is not learned", false)
		return true
	}
	b.slots[slot] = shortcutSlotState{
		kind:       shortcutSkill,
		skillID:    skill.ID,
		skillLevel: skill.Level,
	}
	b.setStatus(fmt.Sprintf("%s assigned to F%d", trimRunes(skillDisplayName(ctx.Resources, skill), 24), slot+1), true)
	b.save(ctx)
	return true
}

func (b *ShortcutBar) ClearDepletedItem(ctx Context, index, itemID uint16) bool {
	changed := b.clearDepletedItemSlots(index, itemID)
	if changed {
		b.save(ctx)
	}
	return changed
}

func (b *ShortcutBar) clearDepletedItemSlots(index, itemID uint16) bool {
	if index == 0 {
		return false
	}
	changed := false
	for slot, entry := range b.slots {
		if entry.kind != shortcutItem || entry.itemIndex != index {
			continue
		}
		if itemID != 0 && entry.itemID != itemID {
			continue
		}
		b.slots[slot] = shortcutSlotState{}
		changed = true
	}
	return changed
}

func (b *ShortcutBar) activate(ctx Context, actions GameActions, slot int) {
	if slot < 0 || slot >= len(b.slots) {
		return
	}
	entry := b.slots[slot]
	switch entry.kind {
	case shortcutItem:
		item, ok := inventoryItemForShortcut(ctx.Session, entry.itemIndex, entry.itemID)
		if !ok {
			b.setStatus(fmt.Sprintf("F%d item unavailable", slot+1), false)
			return
		}
		if err := useInventoryItem(ctx, item); err != nil {
			b.setStatus(err.Error(), false)
			return
		}
		b.setStatus(fmt.Sprintf("F%d item used", slot+1), true)
		log.Printf("shortcut item use slot=%d index=%d item=%d", slot+1, item.Index, item.ItemID)
	case shortcutSkill:
		skill, ok := skillForShortcut(ctx.Session, entry)
		if !ok {
			b.setStatus(fmt.Sprintf("F%d skill unavailable", slot+1), false)
			return
		}
		if actions == nil {
			b.setStatus("No game actions", false)
			return
		}
		if err := actions.UseShortcutSkill(ctx, skill); err != nil {
			b.setStatus(err.Error(), false)
			return
		}
		b.setStatus(fmt.Sprintf("F%d skill used", slot+1), true)
	default:
		b.setStatus(fmt.Sprintf("F%d is empty", slot+1), false)
	}
}

func (b *ShortcutBar) Draw(screen *render.Image, ctx Context, assets AssetRenderer) {
	if screen == nil {
		return
	}
	x, y := b.bounds(ctx)
	width := shortcutSlots*shortcutSlot + (shortcutSlots-1)*shortcutGap + shortcutPad*2
	height := shortcutSlot + shortcutPad*2 + 12
	drawUISurface(screen, x, y, width, height, uiWindowBodyColor, uiWindowBorderColor)
	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i := 0; i < shortcutSlots; i++ {
		sx, sy := b.slotBounds(ctx, i)
		fill := uiButtonColor
		if pointInRect(mx, my, sx, sy, shortcutSlot, shortcutSlot) {
			fill = uiButtonHoverColor
		}
		drawUIButtonSurface(screen, sx, sy, shortcutSlot, shortcutSlot, fill)
		entry := b.slots[i]
		switch entry.kind {
		case shortcutItem:
			if assets != nil {
				item := session.InventoryItem{ItemID: entry.itemID, Index: entry.itemIndex, Identified: entry.identified, Amount: 1}
				if live, ok := inventoryItemForShortcut(ctx.Session, entry.itemIndex, entry.itemID); ok {
					item = live
				}
				assets.DrawInventoryItemIcon(screen, ctx.Resources, item, sx+5, sy+5)
				if item.Amount > 1 {
					render.DebugPrintAtColor(screen, fmt.Sprintf("%d", item.Amount), sx+shortcutSlot-17, sy+shortcutSlot-14, uiTextColor)
				}
			}
		case shortcutSkill:
			if assets != nil {
				skill, _ := skillForShortcut(ctx.Session, entry)
				if skill.ID == 0 {
					skill = session.Skill{ID: entry.skillID, Level: entry.skillLevel}
				}
				assets.DrawSkillIcon(screen, ctx.Resources, skill, sx+5, sy+5, 24)
				if skill.Level > 0 {
					drawShortcutSkillLevel(screen, sx, sy, skill.Level)
				}
			}
		}
		render.DebugPrintAtColor(screen, fmt.Sprintf("F%d", i+1), sx+7, sy+shortcutSlot+1, uiMutedTextColor)
	}
	if b.status != "" && time.Since(b.statusAt) < 1400*time.Millisecond {
		statusColor := skillWindowErrorColor
		if b.statusGood {
			statusColor = skillWindowGoodColor
		}
		render.DebugPrintAtColor(screen, trimRunes(b.status, 42), x+6, y+height+3, statusColor)
	}
}

func (b *ShortcutBar) CursorAction(ctx Context) (int, bool) {
	if ctx.Input == nil {
		return 0, false
	}
	if _, ok := b.slotAt(ctx, ctx.Input.MouseX, ctx.Input.MouseY); ok {
		return CursorActionClick, true
	}
	return 0, false
}

func (b *ShortcutBar) slotAt(ctx Context, mx, my int) (int, bool) {
	for i := 0; i < shortcutSlots; i++ {
		x, y := b.slotBounds(ctx, i)
		if pointInRect(mx, my, x, y, shortcutSlot, shortcutSlot) {
			return i, true
		}
	}
	return 0, false
}

func (b *ShortcutBar) bounds(ctx Context) (int, int) {
	width, _ := ctx.ScreenSize()
	barW := shortcutSlots*shortcutSlot + (shortcutSlots-1)*shortcutGap + shortcutPad*2
	return maxInt(8, (width-barW)/2), 8
}

func (b *ShortcutBar) slotBounds(ctx Context, slot int) (int, int) {
	x, y := b.bounds(ctx)
	return x + shortcutPad + slot*(shortcutSlot+shortcutGap), y + shortcutPad
}

func (b *ShortcutBar) setStatus(text string, good bool) {
	b.status = text
	b.statusGood = good
	b.statusAt = time.Now()
}

func (b *ShortcutBar) Load(ctx Context) {
	if b.loaded {
		return
	}
	b.loaded = true
	path, legacyPath, err := shortcutStatePath(ctx.Session)
	if err != nil {
		log.Printf("shortcut load skipped: %v", err)
		return
	}
	b.path = path
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && legacyPath != "" && legacyPath != path {
			data, err = os.ReadFile(legacyPath)
			if err != nil {
				if !os.IsNotExist(err) {
					log.Printf("shortcut legacy load failed path=%s: %v", legacyPath, err)
				}
				return
			}
			log.Printf("shortcut bar migrating legacy path=%s target=%s", legacyPath, path)
		} else {
			if !os.IsNotExist(err) {
				log.Printf("shortcut load failed path=%s: %v", path, err)
			}
			return
		}
	}
	var saved shortcutPersistFile
	if err := json.Unmarshal(data, &saved); err != nil {
		log.Printf("shortcut load parse failed path=%s: %v", path, err)
		return
	}
	for i := 0; i < len(saved.Slots) && i < shortcutSlots; i++ {
		b.slots[i] = shortcutSlotFromPersist(saved.Slots[i])
	}
	log.Printf("shortcut bar loaded path=%s slots=%d", path, len(saved.Slots))
}

func (b *ShortcutBar) save(ctx Context) {
	if !b.loaded {
		b.Load(ctx)
	}
	path := b.path
	if path == "" {
		var err error
		path, _, err = shortcutStatePath(ctx.Session)
		if err != nil {
			log.Printf("shortcut save skipped: %v", err)
			return
		}
		b.path = path
		b.loaded = true
	}
	if path == "" {
		log.Printf("shortcut save skipped: no character selected")
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Printf("shortcut save mkdir failed path=%s: %v", path, err)
		return
	}
	saved := shortcutPersistFile{
		Version: 1,
		Slots:   make([]shortcutPersistSlot, shortcutSlots),
	}
	for i := 0; i < shortcutSlots; i++ {
		saved.Slots[i] = b.slots[i].persist()
	}
	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		log.Printf("shortcut save marshal failed: %v", err)
		return
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		log.Printf("shortcut save failed path=%s: %v", path, err)
	}
}

func shortcutStatePath(s *session.Session) (string, string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}
	legacy := filepath.Join(dir, "goro", "shortcuts.json")
	key := shortcutCharacterKey(s)
	if key == "" {
		return legacy, legacy, nil
	}
	return filepath.Join(dir, "goro", "shortcuts", key+".json"), legacy, nil
}

func shortcutCharacterKey(s *session.Session) string {
	if s == nil {
		return ""
	}
	if s.Selected.ID != 0 {
		return fmt.Sprintf("char-%d", s.Selected.ID)
	}
	if s.CharID != 0 {
		return fmt.Sprintf("char-%d", s.CharID)
	}
	name := strings.TrimSpace(s.Selected.Name)
	if name == "" {
		return ""
	}
	sanitized := sanitizeShortcutPathPart(name)
	if sanitized == "" {
		return ""
	}
	return "name-" + sanitized
}

func sanitizeShortcutPathPart(value string) string {
	var out strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out.WriteRune(r)
		case r == '-' || r == '_':
			out.WriteRune(r)
		default:
			out.WriteByte('_')
		}
	}
	return strings.Trim(out.String(), "_")
}

func shortcutSlotFromPersist(saved shortcutPersistSlot) shortcutSlotState {
	switch saved.Kind {
	case "item":
		return shortcutSlotState{
			kind:       shortcutItem,
			itemIndex:  saved.ItemIndex,
			itemID:     saved.ItemID,
			identified: saved.Identified,
		}
	case "skill":
		return shortcutSlotState{
			kind:       shortcutSkill,
			skillID:    saved.SkillID,
			skillLevel: saved.SkillLevel,
		}
	default:
		return shortcutSlotState{}
	}
}

func (s shortcutSlotState) persist() shortcutPersistSlot {
	switch s.kind {
	case shortcutItem:
		return shortcutPersistSlot{
			Kind:       "item",
			ItemIndex:  s.itemIndex,
			ItemID:     s.itemID,
			Identified: s.identified,
		}
	case shortcutSkill:
		return shortcutPersistSlot{
			Kind:       "skill",
			SkillID:    s.skillID,
			SkillLevel: s.skillLevel,
		}
	default:
		return shortcutPersistSlot{}
	}
}

func shortcutKey(slot int) render.Key {
	switch slot {
	case 0:
		return render.KeyF1
	case 1:
		return render.KeyF2
	case 2:
		return render.KeyF3
	case 3:
		return render.KeyF4
	case 4:
		return render.KeyF5
	case 5:
		return render.KeyF6
	case 6:
		return render.KeyF7
	case 7:
		return render.KeyF8
	default:
		return render.KeyF9
	}
}

func inventoryItemByIndex(s *session.Session, index uint16) (session.InventoryItem, bool) {
	if s == nil {
		return session.InventoryItem{}, false
	}
	for _, item := range s.Inventory.Items {
		if item.Index != index || item.Amount == 0 {
			continue
		}
		return item, true
	}
	return session.InventoryItem{}, false
}

func inventoryItemForShortcut(s *session.Session, index, itemID uint16) (session.InventoryItem, bool) {
	if item, ok := inventoryItemByIndex(s, index); ok {
		if itemID == 0 || item.ItemID == itemID {
			return item, true
		}
	}
	if s == nil || itemID == 0 {
		return session.InventoryItem{}, false
	}
	for _, item := range s.Inventory.Items {
		if item.ItemID != itemID || item.Amount == 0 {
			continue
		}
		return item, true
	}
	return session.InventoryItem{}, false
}

func skillForShortcut(s *session.Session, entry shortcutSlotState) (session.Skill, bool) {
	if entry.kind != shortcutSkill {
		return session.Skill{}, false
	}
	skill, ok := skillByID(s, entry.skillID)
	if !ok || skill.Level <= 0 {
		return session.Skill{}, false
	}
	level := entry.skillLevel
	if level <= 0 || level > skill.Level {
		level = skill.Level
	}
	skill.Level = level
	return skill, true
}

func drawShortcutSkillLevel(screen *render.Image, x, y int, level int) {
	label := fmt.Sprintf("Lv%d", maxInt(1, level))
	render.DebugPrintAtColor(screen, label, x+2, y+1, color.RGBA{A: 150})
	render.DebugPrintAtColor(screen, label, x+3, y+1, uiTitleTextColor)
}

func useInventoryItem(ctx Context, item session.InventoryItem) error {
	if ctx.Network == nil {
		return fmt.Errorf("not connected")
	}
	if !inventoryItemIsUsable(item) {
		return fmt.Errorf("item cannot be used")
	}
	target := uint32(0)
	if ctx.Session != nil {
		target = ctx.Session.AccountID
		if target == 0 {
			target = ctx.Session.CharID
		}
	}
	if target == 0 {
		return fmt.Errorf("missing player id")
	}
	return ctx.Network.SendUseInventoryItem(item.Index, target)
}
