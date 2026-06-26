package network

import (
	"encoding/binary"
	"fmt"
	"log"
)

const PacketCZItemPickup uint16 = 0x009F

const (
	PacketCZACKSelectDealType  uint16 = 0x00C5
	PacketCZUseItem2           uint16 = 0x0439
	PacketCZUseItemLegacy      uint16 = 0x009F
	PacketCZReqWearEquip       uint16 = 0x00A9
	PacketCZReqTakeoffEquip    uint16 = 0x00AB
	PacketCZPCPurchaseItemList uint16 = 0x00C8
	PacketCZPCSellItemList     uint16 = 0x00C9
)

type itemPickupPacketLayout struct {
	date   int
	opcode uint16
	length int
	offset int
}

var itemPickupPacketLayouts = []itemPickupPacketLayout{
	// Keep this table aligned with rAthena's clif_packetdb.hpp and roBrowser's
	// PacketVersions.js. For 20080910 the last active main-client remap is the
	// 20070212 shuffled 0x00F5 packet.
	{date: 20101124, opcode: 0x0362, length: 6, offset: 2},
	{date: 20070212, opcode: 0x00F5, length: 8, offset: 4},
	{date: 20070108, opcode: 0x00F5, length: 11, offset: 7},
	{date: 20050719, opcode: 0x00F5, length: 13, offset: 9},
	{date: 20050718, opcode: 0x00F5, length: 7, offset: 3},
	{date: 20050628, opcode: 0x00F5, length: 13, offset: 9},
	{date: 20050509, opcode: 0x00F5, length: 8, offset: 4},
	{date: 20050110, opcode: 0x00F5, length: 9, offset: 5},
	{date: 20041129, opcode: 0x00A2, length: 7, offset: 3},
	{date: 20041025, opcode: 0x0113, length: 9, offset: 5},
	{date: 20041005, opcode: 0x0113, length: 10, offset: 6},
	{date: 20040920, opcode: 0x0113, length: 14, offset: 10},
	{date: 20040906, opcode: 0x0113, length: 11, offset: 7},
	{date: 20040809, opcode: 0x0094, length: 13, offset: 9},
	{date: 20040726, opcode: 0x0094, length: 10, offset: 6},
	{date: 20040713, opcode: 0x009F, length: 10, offset: 6},
}

type FloorItemEntry struct {
	ID         uint32
	ItemID     uint16
	Identified bool
	X          int
	Y          int
	SubX       uint8
	SubY       uint8
	Amount     uint16
	Falling    bool
}

type FloorItemDisappear struct {
	ID uint32
}

type ItemPickupAck struct {
	Index      uint16
	Amount     uint16
	ItemID     uint16
	Location   uint16
	Identified bool
	Type       uint8
	Damaged    bool
	Refine     uint8
	Result     uint8
}

type InventoryItem struct {
	Index      uint16
	ItemID     uint16
	Type       uint8
	Location   uint16
	Identified bool
	Amount     uint16
	Equip      bool
	Equipped   bool
	Damaged    bool
	Refine     uint8
}

type InventoryItemDelete struct {
	Index  uint16
	Amount uint16
	Reason uint16
}

type InventoryEquipAck struct {
	Index    uint16
	Location uint16
	Success  bool
	Unequip  bool
}

type ShopDealSelection struct {
	NPCID uint32
}

type ShopBuyItem struct {
	ItemID        uint16
	Type          uint8
	Price         uint32
	DiscountPrice uint32
}

type ShopSellItem struct {
	Index           uint16
	Price           uint32
	OverchargePrice uint32
}

type ShopResult struct {
	Sell   bool
	Result uint8
}

type SellRequestItem struct {
	Index  uint16
	Amount uint16
}

type BuyRequestItem struct {
	ItemID uint16
	Amount uint16
}

func ParseFloorItemEntry(packet Packet) (FloorItemEntry, bool, error) {
	switch packet.ID {
	case 0x009D:
		if len(packet.Data) < 17 {
			return FloorItemEntry{}, false, fmt.Errorf("ZC_ITEM_ENTRY too short: %d", len(packet.Data))
		}
		return FloorItemEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			ItemID:     binary.LittleEndian.Uint16(packet.Data[6:8]),
			Identified: packet.Data[8] != 0,
			X:          int(binary.LittleEndian.Uint16(packet.Data[9:11])),
			Y:          int(binary.LittleEndian.Uint16(packet.Data[11:13])),
			Amount:     binary.LittleEndian.Uint16(packet.Data[13:15]),
			SubX:       packet.Data[15],
			SubY:       packet.Data[16],
		}, true, nil
	case 0x009E:
		if len(packet.Data) < 17 {
			return FloorItemEntry{}, false, fmt.Errorf("ZC_ITEM_FALL_ENTRY too short: %d", len(packet.Data))
		}
		return FloorItemEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			ItemID:     binary.LittleEndian.Uint16(packet.Data[6:8]),
			Identified: packet.Data[8] != 0,
			X:          int(binary.LittleEndian.Uint16(packet.Data[9:11])),
			Y:          int(binary.LittleEndian.Uint16(packet.Data[11:13])),
			SubX:       packet.Data[13],
			SubY:       packet.Data[14],
			Amount:     binary.LittleEndian.Uint16(packet.Data[15:17]),
			Falling:    true,
		}, true, nil
	default:
		return FloorItemEntry{}, false, nil
	}
}

func ParseFloorItemDisappear(packet Packet) (FloorItemDisappear, bool, error) {
	if packet.ID != 0x00A1 {
		return FloorItemDisappear{}, false, nil
	}
	if len(packet.Data) < 6 {
		return FloorItemDisappear{}, false, fmt.Errorf("ZC_ITEM_DISAPPEAR too short: %d", len(packet.Data))
	}
	return FloorItemDisappear{ID: binary.LittleEndian.Uint32(packet.Data[2:6])}, true, nil
}

func ParseItemPickupAck(packet Packet) (ItemPickupAck, bool, error) {
	switch packet.ID {
	case 0x00A0, 0x029A, 0x02D4:
	default:
		return ItemPickupAck{}, false, nil
	}
	if len(packet.Data) < 23 {
		return ItemPickupAck{}, false, fmt.Errorf("ZC_ITEM_PICKUP_ACK 0x%04X too short: %d", packet.ID, len(packet.Data))
	}
	return ItemPickupAck{
		Index:      binary.LittleEndian.Uint16(packet.Data[2:4]),
		Amount:     binary.LittleEndian.Uint16(packet.Data[4:6]),
		ItemID:     binary.LittleEndian.Uint16(packet.Data[6:8]),
		Identified: packet.Data[8] != 0,
		Damaged:    packet.Data[9] != 0,
		Refine:     packet.Data[10],
		Location:   binary.LittleEndian.Uint16(packet.Data[19:21]),
		Type:       packet.Data[21],
		Result:     packet.Data[22],
	}, true, nil
}

func ParseInventoryItemList(packet Packet) ([]InventoryItem, bool, error) {
	switch packet.ID {
	case 0x00A3:
		return parseNormalInventoryItems(packet, 10)
	case 0x01EE:
		return parseNormalInventoryItems(packet, 18)
	case 0x02E8:
		return parseNormalInventoryItems(packet, 22)
	case 0x00A4:
		return parseEquipInventoryItems(packet, 20)
	case 0x0295:
		return parseEquipInventoryItems(packet, 24)
	case 0x02D0:
		return parseEquipInventoryItems(packet, 28)
	default:
		return nil, false, nil
	}
}

func parseNormalInventoryItems(packet Packet, entrySize int) ([]InventoryItem, bool, error) {
	if len(packet.Data) < 4 {
		return nil, false, fmt.Errorf("ZC_NORMAL_ITEMLIST 0x%04X too short: %d", packet.ID, len(packet.Data))
	}
	if (len(packet.Data)-4)%entrySize != 0 {
		return nil, false, fmt.Errorf("ZC_NORMAL_ITEMLIST 0x%04X invalid length: %d", packet.ID, len(packet.Data))
	}
	items := make([]InventoryItem, 0, (len(packet.Data)-4)/entrySize)
	for offset := 4; offset+entrySize <= len(packet.Data); offset += entrySize {
		items = append(items, InventoryItem{
			Index:      binary.LittleEndian.Uint16(packet.Data[offset : offset+2]),
			ItemID:     binary.LittleEndian.Uint16(packet.Data[offset+2 : offset+4]),
			Type:       packet.Data[offset+4],
			Identified: packet.Data[offset+5] != 0,
			Amount:     binary.LittleEndian.Uint16(packet.Data[offset+6 : offset+8]),
			Location:   binary.LittleEndian.Uint16(packet.Data[offset+8 : offset+10]),
		})
	}
	return items, true, nil
}

func parseEquipInventoryItems(packet Packet, entrySize int) ([]InventoryItem, bool, error) {
	if len(packet.Data) < 4 {
		return nil, false, fmt.Errorf("ZC_EQUIPMENT_ITEMLIST 0x%04X too short: %d", packet.ID, len(packet.Data))
	}
	if (len(packet.Data)-4)%entrySize != 0 {
		return nil, false, fmt.Errorf("ZC_EQUIPMENT_ITEMLIST 0x%04X invalid length: %d", packet.ID, len(packet.Data))
	}
	items := make([]InventoryItem, 0, (len(packet.Data)-4)/entrySize)
	for offset := 4; offset+entrySize <= len(packet.Data); offset += entrySize {
		wearState := binary.LittleEndian.Uint16(packet.Data[offset+8 : offset+10])
		items = append(items, InventoryItem{
			Index:      binary.LittleEndian.Uint16(packet.Data[offset : offset+2]),
			ItemID:     binary.LittleEndian.Uint16(packet.Data[offset+2 : offset+4]),
			Type:       packet.Data[offset+4],
			Identified: packet.Data[offset+5] != 0,
			Location:   binary.LittleEndian.Uint16(packet.Data[offset+6 : offset+8]),
			Amount:     1,
			Equip:      true,
			Equipped:   wearState != 0,
			Damaged:    packet.Data[offset+10] != 0,
			Refine:     packet.Data[offset+11],
		})
	}
	return items, true, nil
}

func ParseInventoryItemDelete(packet Packet) (InventoryItemDelete, bool, error) {
	switch packet.ID {
	case 0x00AF:
		if len(packet.Data) < 6 {
			return InventoryItemDelete{}, false, fmt.Errorf("ZC_ITEM_THROW_ACK too short: %d", len(packet.Data))
		}
		return InventoryItemDelete{
			Index:  binary.LittleEndian.Uint16(packet.Data[2:4]),
			Amount: binary.LittleEndian.Uint16(packet.Data[4:6]),
		}, true, nil
	case 0x07FA:
		if len(packet.Data) < 8 {
			return InventoryItemDelete{}, false, fmt.Errorf("ZC_DELETE_ITEM_FROM_BODY too short: %d", len(packet.Data))
		}
		return InventoryItemDelete{
			Reason: binary.LittleEndian.Uint16(packet.Data[2:4]),
			Index:  binary.LittleEndian.Uint16(packet.Data[4:6]),
			Amount: binary.LittleEndian.Uint16(packet.Data[6:8]),
		}, true, nil
	default:
		return InventoryItemDelete{}, false, nil
	}
}

func ParseInventoryEquipAck(packet Packet) (InventoryEquipAck, bool, error) {
	switch packet.ID {
	case 0x00AA:
		if len(packet.Data) < 7 {
			return InventoryEquipAck{}, false, fmt.Errorf("ZC_REQ_WEAR_EQUIP_ACK too short: %d", len(packet.Data))
		}
		return InventoryEquipAck{
			Index:    binary.LittleEndian.Uint16(packet.Data[2:4]),
			Location: binary.LittleEndian.Uint16(packet.Data[4:6]),
			Success:  packet.Data[6] != 0,
		}, true, nil
	case 0x00AC:
		if len(packet.Data) < 7 {
			return InventoryEquipAck{}, false, fmt.Errorf("ZC_REQ_TAKEOFF_EQUIP_ACK too short: %d", len(packet.Data))
		}
		return InventoryEquipAck{
			Index:    binary.LittleEndian.Uint16(packet.Data[2:4]),
			Location: binary.LittleEndian.Uint16(packet.Data[4:6]),
			Success:  packet.Data[6] != 0,
			Unequip:  true,
		}, true, nil
	default:
		return InventoryEquipAck{}, false, nil
	}
}

func ParseShopDealSelection(packet Packet) (ShopDealSelection, bool, error) {
	if packet.ID != 0x00C4 {
		return ShopDealSelection{}, false, nil
	}
	if len(packet.Data) < 6 {
		return ShopDealSelection{}, false, fmt.Errorf("ZC_SELECT_DEALTYPE too short: %d", len(packet.Data))
	}
	return ShopDealSelection{NPCID: binary.LittleEndian.Uint32(packet.Data[2:6])}, true, nil
}

func ParseShopBuyList(packet Packet) ([]ShopBuyItem, bool, error) {
	if packet.ID != 0x00C6 {
		return nil, false, nil
	}
	if len(packet.Data) < 4 {
		return nil, false, fmt.Errorf("ZC_PC_PURCHASE_ITEMLIST too short: %d", len(packet.Data))
	}
	if (len(packet.Data)-4)%11 != 0 {
		return nil, false, fmt.Errorf("ZC_PC_PURCHASE_ITEMLIST invalid length: %d", len(packet.Data))
	}
	items := make([]ShopBuyItem, 0, (len(packet.Data)-4)/11)
	for offset := 4; offset+11 <= len(packet.Data); offset += 11 {
		items = append(items, ShopBuyItem{
			Price:         binary.LittleEndian.Uint32(packet.Data[offset : offset+4]),
			DiscountPrice: binary.LittleEndian.Uint32(packet.Data[offset+4 : offset+8]),
			Type:          packet.Data[offset+8],
			ItemID:        binary.LittleEndian.Uint16(packet.Data[offset+9 : offset+11]),
		})
	}
	return items, true, nil
}

func ParseShopSellList(packet Packet) ([]ShopSellItem, bool, error) {
	if packet.ID != 0x00C7 {
		return nil, false, nil
	}
	if len(packet.Data) < 4 {
		return nil, false, fmt.Errorf("ZC_PC_SELL_ITEMLIST too short: %d", len(packet.Data))
	}
	if (len(packet.Data)-4)%10 != 0 {
		return nil, false, fmt.Errorf("ZC_PC_SELL_ITEMLIST invalid length: %d", len(packet.Data))
	}
	items := make([]ShopSellItem, 0, (len(packet.Data)-4)/10)
	for offset := 4; offset+10 <= len(packet.Data); offset += 10 {
		items = append(items, ShopSellItem{
			Index:           binary.LittleEndian.Uint16(packet.Data[offset : offset+2]),
			Price:           binary.LittleEndian.Uint32(packet.Data[offset+2 : offset+6]),
			OverchargePrice: binary.LittleEndian.Uint32(packet.Data[offset+6 : offset+10]),
		})
	}
	return items, true, nil
}

func ParseShopResult(packet Packet) (ShopResult, bool, error) {
	switch packet.ID {
	case 0x00CA:
		if len(packet.Data) < 3 {
			return ShopResult{}, false, fmt.Errorf("ZC_PC_PURCHASE_RESULT too short: %d", len(packet.Data))
		}
		return ShopResult{Result: packet.Data[2]}, true, nil
	case 0x00CB:
		if len(packet.Data) < 3 {
			return ShopResult{}, false, fmt.Errorf("ZC_PC_SELL_RESULT too short: %d", len(packet.Data))
		}
		return ShopResult{Sell: true, Result: packet.Data[2]}, true, nil
	default:
		return ShopResult{}, false, nil
	}
}

func BuildItemPickupPacket(gid uint32) []byte {
	var w Writer
	w.Uint16(PacketCZItemPickup)
	w.Uint32(gid)
	return w.Bytes()
}

func BuildItemPickupPacketForClientDate(gid uint32, clientDate int) []byte {
	for _, layout := range itemPickupPacketLayouts {
		if clientDate >= layout.date {
			packet := make([]byte, layout.length)
			binary.LittleEndian.PutUint16(packet[0:2], layout.opcode)
			binary.LittleEndian.PutUint32(packet[layout.offset:layout.offset+4], gid)
			return packet
		}
	}
	return BuildItemPickupPacket(gid)
}

func BuildUseInventoryItemPacketForClientDate(index uint16, targetAID uint32, clientDate int) []byte {
	if clientDate >= 20180307 {
		packet := make([]byte, 8)
		binary.LittleEndian.PutUint16(packet[0:2], PacketCZUseItem2)
		binary.LittleEndian.PutUint16(packet[2:4], index)
		binary.LittleEndian.PutUint32(packet[4:8], targetAID)
		return packet
	}
	packet := make([]byte, 14)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZUseItemLegacy)
	binary.LittleEndian.PutUint16(packet[4:6], index)
	binary.LittleEndian.PutUint32(packet[10:14], targetAID)
	return packet
}

func BuildWearEquipPacket(index, location uint16) []byte {
	packet := make([]byte, 6)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZReqWearEquip)
	binary.LittleEndian.PutUint16(packet[2:4], index)
	binary.LittleEndian.PutUint16(packet[4:6], location)
	return packet
}

func BuildTakeoffEquipPacket(index uint16) []byte {
	packet := make([]byte, 4)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZReqTakeoffEquip)
	binary.LittleEndian.PutUint16(packet[2:4], index)
	return packet
}

func BuildShopDealSelectionPacket(npcID uint32, dealType uint8) []byte {
	var w Writer
	w.Uint16(PacketCZACKSelectDealType)
	w.Uint32(npcID)
	w.Uint8(dealType)
	return w.Bytes()
}

func BuildSellItemListPacket(items []SellRequestItem) []byte {
	size := 4 + len(items)*4
	packet := make([]byte, size)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZPCSellItemList)
	binary.LittleEndian.PutUint16(packet[2:4], uint16(size))
	offset := 4
	for _, item := range items {
		binary.LittleEndian.PutUint16(packet[offset:offset+2], item.Index)
		binary.LittleEndian.PutUint16(packet[offset+2:offset+4], item.Amount)
		offset += 4
	}
	return packet
}

func BuildBuyItemListPacket(items []BuyRequestItem) []byte {
	size := 4 + len(items)*4
	packet := make([]byte, size)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZPCPurchaseItemList)
	binary.LittleEndian.PutUint16(packet[2:4], uint16(size))
	offset := 4
	for _, item := range items {
		binary.LittleEndian.PutUint16(packet[offset:offset+2], item.Amount)
		binary.LittleEndian.PutUint16(packet[offset+2:offset+4], item.ItemID)
		offset += 4
	}
	return packet
}

func (c *Client) SendItemPickup(gid uint32) error {
	packet := BuildItemPickupPacketForClientDate(gid, c.clientDate)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_ITEM_PICKUP opcode=0x%04X target=%d client_date=%d", ID(packet), gid, c.clientDate)
	} else {
		log.Printf("send CZ_ITEM_PICKUP failed opcode=0x%04X len=%d target=%d client_date=%d: %v", ID(packet), len(packet), gid, c.clientDate, err)
	}
	return err
}

func (c *Client) SendUseInventoryItem(index uint16, targetAID uint32) error {
	packet := BuildUseInventoryItemPacketForClientDate(index, targetAID, c.clientDate)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_USE_ITEM opcode=0x%04X index=%d target=%d client_date=%d", ID(packet), index, targetAID, c.clientDate)
	} else {
		log.Printf("send CZ_USE_ITEM failed opcode=0x%04X len=%d index=%d target=%d client_date=%d: %v", ID(packet), len(packet), index, targetAID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendWearEquip(index, location uint16) error {
	packet := BuildWearEquipPacket(index, location)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_REQ_WEAR_EQUIP opcode=0x%04X index=%d location=0x%04X client_date=%d", ID(packet), index, location, c.clientDate)
	} else {
		log.Printf("send CZ_REQ_WEAR_EQUIP failed opcode=0x%04X index=%d location=0x%04X client_date=%d: %v", ID(packet), index, location, c.clientDate, err)
	}
	return err
}

func (c *Client) SendTakeoffEquip(index uint16) error {
	packet := BuildTakeoffEquipPacket(index)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_REQ_TAKEOFF_EQUIP opcode=0x%04X index=%d client_date=%d", ID(packet), index, c.clientDate)
	} else {
		log.Printf("send CZ_REQ_TAKEOFF_EQUIP failed opcode=0x%04X index=%d client_date=%d: %v", ID(packet), index, c.clientDate, err)
	}
	return err
}

func (c *Client) SendShopDealSelection(npcID uint32, dealType uint8) error {
	packet := BuildShopDealSelectionPacket(npcID, dealType)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_ACK_SELECT_DEALTYPE opcode=0x%04X npc=%d type=%d client_date=%d", ID(packet), npcID, dealType, c.clientDate)
	} else {
		log.Printf("send CZ_ACK_SELECT_DEALTYPE failed opcode=0x%04X npc=%d type=%d client_date=%d: %v", ID(packet), npcID, dealType, c.clientDate, err)
	}
	return err
}

func (c *Client) SendShopSellItems(items []SellRequestItem) error {
	packet := BuildSellItemListPacket(items)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_PC_SELL_ITEMLIST opcode=0x%04X count=%d client_date=%d", ID(packet), len(items), c.clientDate)
	} else {
		log.Printf("send CZ_PC_SELL_ITEMLIST failed opcode=0x%04X count=%d client_date=%d: %v", ID(packet), len(items), c.clientDate, err)
	}
	return err
}

func (c *Client) SendShopBuyItems(items []BuyRequestItem) error {
	packet := BuildBuyItemListPacket(items)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_PC_PURCHASE_ITEMLIST opcode=0x%04X count=%d client_date=%d", ID(packet), len(items), c.clientDate)
	} else {
		log.Printf("send CZ_PC_PURCHASE_ITEMLIST failed opcode=0x%04X count=%d client_date=%d: %v", ID(packet), len(items), c.clientDate, err)
	}
	return err
}
