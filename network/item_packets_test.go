package network

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestParseFloorItemEntryExistingItem(t *testing.T) {
	data := make([]byte, 17)
	binary.LittleEndian.PutUint16(data[0:2], 0x009D)
	binary.LittleEndian.PutUint32(data[2:6], 1001)
	binary.LittleEndian.PutUint16(data[6:8], 909)
	data[8] = 1
	binary.LittleEndian.PutUint16(data[9:11], 150)
	binary.LittleEndian.PutUint16(data[11:13], 160)
	binary.LittleEndian.PutUint16(data[13:15], 3)
	data[15] = 4
	data[16] = 8

	item, ok, err := ParseFloorItemEntry(Packet{ID: 0x009D, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if item.ID != 1001 || item.ItemID != 909 || !item.Identified || item.X != 150 || item.Y != 160 || item.Amount != 3 || item.SubX != 4 || item.SubY != 8 || item.Falling {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestParseFloorItemEntryFallingItem(t *testing.T) {
	data := make([]byte, 17)
	binary.LittleEndian.PutUint16(data[0:2], 0x009E)
	binary.LittleEndian.PutUint32(data[2:6], 2002)
	binary.LittleEndian.PutUint16(data[6:8], 512)
	data[8] = 0
	binary.LittleEndian.PutUint16(data[9:11], 30)
	binary.LittleEndian.PutUint16(data[11:13], 40)
	data[13] = 5
	data[14] = 7
	binary.LittleEndian.PutUint16(data[15:17], 9)

	item, ok, err := ParseFloorItemEntry(Packet{ID: 0x009E, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if item.ID != 2002 || item.ItemID != 512 || item.Identified || item.X != 30 || item.Y != 40 || item.Amount != 9 || item.SubX != 5 || item.SubY != 7 || !item.Falling {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestParseFloorItemDisappear(t *testing.T) {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], 0x00A1)
	binary.LittleEndian.PutUint32(data[2:6], 3003)

	disappear, ok, err := ParseFloorItemDisappear(Packet{ID: 0x00A1, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || disappear.ID != 3003 {
		t.Fatalf("unexpected disappear: ok=%v value=%+v", ok, disappear)
	}
}

func TestParseItemPickupAckLegacy(t *testing.T) {
	data := make([]byte, 23)
	binary.LittleEndian.PutUint16(data[0:2], 0x00A0)
	binary.LittleEndian.PutUint16(data[2:4], 12)
	binary.LittleEndian.PutUint16(data[4:6], 3)
	binary.LittleEndian.PutUint16(data[6:8], 938)
	data[8] = 1
	binary.LittleEndian.PutUint16(data[19:21], 0x0002)
	data[21] = 5
	data[22] = 0

	ack, ok, err := ParseItemPickupAck(Packet{ID: 0x00A0, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if ack.Index != 12 || ack.Amount != 3 || ack.ItemID != 938 || ack.Location != 0x0002 || ack.Type != 5 || !ack.Identified || ack.Result != 0 {
		t.Fatalf("unexpected ack: %+v", ack)
	}
}

func TestParseItemPickupAckExtendedVariants(t *testing.T) {
	for _, tc := range []struct {
		id             uint16
		length         int
		locationOffset int
		locationSize   int
		typeOffset     int
		resultOffset   int
	}{
		{id: 0x029A, length: 27, locationOffset: 19, locationSize: 2, typeOffset: 21, resultOffset: 22},
		{id: 0x02D4, length: 29, locationOffset: 19, locationSize: 2, typeOffset: 21, resultOffset: 22},
		{id: 0x0990, length: 31, locationOffset: 19, locationSize: 4, typeOffset: 23, resultOffset: 24},
		{id: 0x0A0C, length: 56, locationOffset: 19, locationSize: 4, typeOffset: 23, resultOffset: 24},
		{id: 0x0A37, length: 59, locationOffset: 19, locationSize: 4, typeOffset: 23, resultOffset: 24},
	} {
		t.Run(fmt.Sprintf("0x%04X", tc.id), func(t *testing.T) {
			data := make([]byte, tc.length)
			binary.LittleEndian.PutUint16(data[0:2], tc.id)
			binary.LittleEndian.PutUint16(data[2:4], 31)
			binary.LittleEndian.PutUint16(data[4:6], 2)
			binary.LittleEndian.PutUint16(data[6:8], 909)
			data[8] = 1
			if tc.locationSize == 4 {
				binary.LittleEndian.PutUint32(data[tc.locationOffset:tc.locationOffset+4], 0x00000020)
			} else {
				binary.LittleEndian.PutUint16(data[tc.locationOffset:tc.locationOffset+2], 0x0020)
			}
			data[tc.typeOffset] = 5
			data[tc.resultOffset] = 0

			ack, ok, err := ParseItemPickupAck(Packet{ID: tc.id, Data: data})
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatal("not parsed")
			}
			if ack.Index != 31 || ack.Amount != 2 || ack.ItemID != 909 || ack.Location != 0x0020 || ack.Type != 5 || !ack.Identified || ack.Result != 0 {
				t.Fatalf("unexpected ack: %+v", ack)
			}
		})
	}
}

func TestParseUseItemAckLegacy(t *testing.T) {
	data := make([]byte, 7)
	binary.LittleEndian.PutUint16(data[0:2], 0x00A8)
	binary.LittleEndian.PutUint16(data[2:4], 12)
	binary.LittleEndian.PutUint16(data[4:6], 3)
	data[6] = 1

	ack, ok, err := ParseUseItemAck(Packet{ID: 0x00A8, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if ack.Index != 12 || ack.Amount != 3 || ack.Result != 1 {
		t.Fatalf("unexpected use ack: %+v", ack)
	}
}

func TestParseUseItemAck2(t *testing.T) {
	data := make([]byte, 13)
	binary.LittleEndian.PutUint16(data[0:2], 0x01C8)
	binary.LittleEndian.PutUint16(data[2:4], 12)
	binary.LittleEndian.PutUint16(data[4:6], 512)
	binary.LittleEndian.PutUint32(data[6:10], 0x11223344)
	binary.LittleEndian.PutUint16(data[10:12], 3)
	data[12] = 1

	ack, ok, err := ParseUseItemAck(Packet{ID: 0x01C8, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if ack.Index != 12 || ack.ItemID != 512 || ack.AID != 0x11223344 || ack.Amount != 3 || ack.Result != 1 {
		t.Fatalf("unexpected use ack: %+v", ack)
	}
}

func TestBuildItemPickupPacket(t *testing.T) {
	packet := BuildItemPickupPacket(0x11223344)
	if len(packet) != 6 {
		t.Fatalf("len = %d, want 6", len(packet))
	}
	if got := ID(packet); got != 0x009F {
		t.Fatalf("opcode = 0x%04X, want 0x009F", got)
	}
	if got := binary.LittleEndian.Uint32(packet[2:6]); got != 0x11223344 {
		t.Fatalf("gid = 0x%08X", got)
	}
}

func TestBuildItemPickupPacketForClientDate20080910(t *testing.T) {
	packet := BuildItemPickupPacketForClientDate(0x11223344, 20080910)
	if len(packet) != 8 {
		t.Fatalf("len = %d, want 8", len(packet))
	}
	if got := ID(packet); got != 0x00F5 {
		t.Fatalf("opcode = 0x%04X, want 0x00F5", got)
	}
	if got := binary.LittleEndian.Uint32(packet[4:8]); got != 0x11223344 {
		t.Fatalf("gid = 0x%08X", got)
	}
	if padding := binary.LittleEndian.Uint16(packet[2:4]); padding != 0 {
		t.Fatalf("padding = 0x%04X, want 0", padding)
	}
}

func TestBuildItemPickupPacketForClientDateShuffledBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		clientDate int
		opcode     uint16
		length     int
		offset     int
	}{
		{name: "legacy", clientDate: 20040712, opcode: 0x009F, length: 6, offset: 2},
		{name: "20040713", clientDate: 20040713, opcode: 0x009F, length: 10, offset: 6},
		{name: "20040726", clientDate: 20040726, opcode: 0x0094, length: 10, offset: 6},
		{name: "20070212", clientDate: 20070212, opcode: 0x00F5, length: 8, offset: 4},
		{name: "20101124", clientDate: 20101124, opcode: 0x0362, length: 6, offset: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet := BuildItemPickupPacketForClientDate(0x11223344, tc.clientDate)
			if len(packet) != tc.length {
				t.Fatalf("len = %d, want %d", len(packet), tc.length)
			}
			if got := ID(packet); got != tc.opcode {
				t.Fatalf("opcode = 0x%04X, want 0x%04X", got, tc.opcode)
			}
			if got := binary.LittleEndian.Uint32(packet[tc.offset : tc.offset+4]); got != 0x11223344 {
				t.Fatalf("gid = 0x%08X", got)
			}
		})
	}
}

func TestBuildUseAndEquipItemPackets(t *testing.T) {
	use := BuildUseInventoryItemPacketForClientDate(7, 0x11223344, 20080910)
	if len(use) != 8 || ID(use) != 0x0439 {
		t.Fatalf("unexpected use packet header: % X", use)
	}
	if got := binary.LittleEndian.Uint16(use[2:4]); got != 7 {
		t.Fatalf("use item index = %d, want 7", got)
	}
	if got := binary.LittleEndian.Uint32(use[4:8]); got != 0x11223344 {
		t.Fatalf("use target = 0x%08X", got)
	}

	useLegacy := BuildUseInventoryItemPacketForClientDate(7, 0x11223344, 20070212)
	if len(useLegacy) != 14 || ID(useLegacy) != 0x009F {
		t.Fatalf("unexpected legacy use packet header: % X", useLegacy)
	}
	if got := binary.LittleEndian.Uint16(useLegacy[4:6]); got != 7 {
		t.Fatalf("legacy use item index = %d, want 7", got)
	}
	if got := binary.LittleEndian.Uint32(useLegacy[10:14]); got != 0x11223344 {
		t.Fatalf("legacy use target = 0x%08X", got)
	}

	use2 := BuildUseInventoryItemPacketForClientDate(7, 0x11223344, 20180307)
	if len(use2) != 8 || ID(use2) != 0x0439 || binary.LittleEndian.Uint16(use2[2:4]) != 7 || binary.LittleEndian.Uint32(use2[4:8]) != 0x11223344 {
		t.Fatalf("unexpected modern use packet: % X", use2)
	}

	equip := BuildWearEquipPacket(11, 0x0002)
	if len(equip) != 6 || ID(equip) != 0x00A9 || binary.LittleEndian.Uint16(equip[2:4]) != 11 || binary.LittleEndian.Uint16(equip[4:6]) != 0x0002 {
		t.Fatalf("unexpected equip packet: % X", equip)
	}

	takeoff := BuildTakeoffEquipPacket(11)
	if len(takeoff) != 4 || ID(takeoff) != 0x00AB || binary.LittleEndian.Uint16(takeoff[2:4]) != 11 {
		t.Fatalf("unexpected takeoff packet: % X", takeoff)
	}
}

func TestParseInventoryItemListNormalLegacy(t *testing.T) {
	data := make([]byte, 4+10)
	binary.LittleEndian.PutUint16(data[0:2], 0x00A3)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 7)
	binary.LittleEndian.PutUint16(data[6:8], 938)
	data[8] = 3
	data[9] = 1
	binary.LittleEndian.PutUint16(data[10:12], 4)
	binary.LittleEndian.PutUint16(data[12:14], 0x0001)

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x00A3, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 7 || got.ItemID != 938 || got.Type != 3 || got.Location != 0x0001 || !got.Identified || got.Amount != 4 || got.Equip {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestParseInventoryItemListNormal2008(t *testing.T) {
	data := make([]byte, 4+22)
	binary.LittleEndian.PutUint16(data[0:2], 0x02E8)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 7)
	binary.LittleEndian.PutUint16(data[6:8], 938)
	data[8] = 3
	data[9] = 1
	binary.LittleEndian.PutUint16(data[10:12], 4)
	binary.LittleEndian.PutUint16(data[12:14], 0x0001)
	binary.LittleEndian.PutUint16(data[14:16], 4001)
	binary.LittleEndian.PutUint32(data[22:26], 123456)

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x02E8, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 7 || got.ItemID != 938 || got.Type != 3 || got.Location != 0x0001 || !got.Identified || got.Amount != 4 || got.Equip {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestParseInventoryItemListNormal4(t *testing.T) {
	data := make([]byte, 4+24)
	binary.LittleEndian.PutUint16(data[0:2], 0x0991)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 7)
	binary.LittleEndian.PutUint16(data[6:8], 938)
	data[8] = 3
	binary.LittleEndian.PutUint16(data[9:11], 4)
	binary.LittleEndian.PutUint32(data[11:15], 0x00000001)
	binary.LittleEndian.PutUint32(data[23:27], 123456)
	data[27] = 1

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x0991, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 7 || got.ItemID != 938 || got.Type != 3 || got.Location != 0x0001 || !got.Identified || got.Amount != 4 || got.Equip {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestParseInventoryItemListEquipmentLegacy(t *testing.T) {
	data := make([]byte, 4+20)
	binary.LittleEndian.PutUint16(data[0:2], 0x00A4)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 11)
	binary.LittleEndian.PutUint16(data[6:8], 1201)
	data[8] = 4
	data[9] = 1
	binary.LittleEndian.PutUint16(data[10:12], 0x0002)
	binary.LittleEndian.PutUint16(data[12:14], 0x0002)
	data[14] = 1
	data[15] = 5

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x00A4, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 11 || got.ItemID != 1201 || got.Type != 4 || got.Location != 0x0002 || !got.Identified || got.Amount != 1 || !got.Equip || !got.Equipped || !got.Damaged || got.Refine != 5 {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestParseInventoryItemListEquipment4(t *testing.T) {
	data := make([]byte, 4+31)
	binary.LittleEndian.PutUint16(data[0:2], 0x0992)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 11)
	binary.LittleEndian.PutUint16(data[6:8], 1201)
	data[8] = 4
	binary.LittleEndian.PutUint32(data[9:13], 0x00000002)
	binary.LittleEndian.PutUint32(data[13:17], 0x00000002)
	data[17] = 5
	data[34] = 3

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x0992, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 11 || got.ItemID != 1201 || got.Type != 4 || got.Location != 0x0002 || !got.Identified || got.Amount != 1 || !got.Equip || !got.Equipped || !got.Damaged || got.Refine != 5 {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestParseInventoryItemListEquipment5(t *testing.T) {
	data := make([]byte, 4+57)
	binary.LittleEndian.PutUint16(data[0:2], 0x0A0D)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 12)
	binary.LittleEndian.PutUint16(data[6:8], 1101)
	data[8] = 5
	binary.LittleEndian.PutUint32(data[9:13], 0x00000002)
	data[17] = 7
	data[60] = 1

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x0A0D, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 12 || got.ItemID != 1101 || got.Type != 5 || got.Location != 0x0002 || !got.Identified || got.Amount != 1 || !got.Equip || got.Equipped || got.Damaged || got.Refine != 7 {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestParseInventoryItemListEquipment2008(t *testing.T) {
	data := make([]byte, 4+26)
	binary.LittleEndian.PutUint16(data[0:2], 0x02D0)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 11)
	binary.LittleEndian.PutUint16(data[6:8], 1201)
	data[8] = 4
	data[9] = 1
	binary.LittleEndian.PutUint16(data[10:12], 0x0002)
	binary.LittleEndian.PutUint16(data[12:14], 0x0002)
	data[14] = 1
	data[15] = 5
	binary.LittleEndian.PutUint32(data[24:28], 123456)
	binary.LittleEndian.PutUint16(data[28:30], 1)

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x02D0, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 11 || got.ItemID != 1201 || got.Type != 4 || got.Location != 0x0002 || !got.Identified || got.Amount != 1 || !got.Equip || !got.Equipped || !got.Damaged || got.Refine != 5 {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestParseShopDealAndSellList(t *testing.T) {
	dealData := make([]byte, 6)
	binary.LittleEndian.PutUint16(dealData[0:2], 0x00C4)
	binary.LittleEndian.PutUint32(dealData[2:6], 0x11223344)
	deal, ok, err := ParseShopDealSelection(Packet{ID: 0x00C4, Data: dealData})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || deal.NPCID != 0x11223344 {
		t.Fatalf("unexpected deal: ok=%v value=%+v", ok, deal)
	}

	sellData := make([]byte, 4+10)
	binary.LittleEndian.PutUint16(sellData[0:2], 0x00C7)
	binary.LittleEndian.PutUint16(sellData[2:4], uint16(len(sellData)))
	binary.LittleEndian.PutUint16(sellData[4:6], 12)
	binary.LittleEndian.PutUint32(sellData[6:10], 10)
	binary.LittleEndian.PutUint32(sellData[10:14], 12)
	items, ok, err := ParseShopSellList(Packet{ID: 0x00C7, Data: sellData})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("sell items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 12 || got.Price != 10 || got.OverchargePrice != 12 {
		t.Fatalf("unexpected sell item: %+v", got)
	}
}

func TestBuildShopPackets(t *testing.T) {
	deal := BuildShopDealSelectionPacket(0x11223344, 1)
	if len(deal) != 7 || ID(deal) != 0x00C5 || binary.LittleEndian.Uint32(deal[2:6]) != 0x11223344 || deal[6] != 1 {
		t.Fatalf("unexpected deal packet: % X", deal)
	}

	buy := BuildBuyItemListPacket([]BuyRequestItem{{ItemID: 501, Amount: 2}, {ItemID: 502, Amount: 3}})
	if len(buy) != 12 || ID(buy) != 0x00C8 || binary.LittleEndian.Uint16(buy[2:4]) != 12 {
		t.Fatalf("unexpected buy packet header: % X", buy)
	}
	if binary.LittleEndian.Uint16(buy[4:6]) != 2 || binary.LittleEndian.Uint16(buy[6:8]) != 501 || binary.LittleEndian.Uint16(buy[8:10]) != 3 || binary.LittleEndian.Uint16(buy[10:12]) != 502 {
		t.Fatalf("unexpected buy packet items: % X", buy)
	}

	sell := BuildSellItemListPacket([]SellRequestItem{{Index: 2, Amount: 3}, {Index: 4, Amount: 5}})
	if len(sell) != 12 || ID(sell) != 0x00C9 || binary.LittleEndian.Uint16(sell[2:4]) != 12 {
		t.Fatalf("unexpected sell packet header: % X", sell)
	}
	if binary.LittleEndian.Uint16(sell[4:6]) != 2 || binary.LittleEndian.Uint16(sell[6:8]) != 3 || binary.LittleEndian.Uint16(sell[8:10]) != 4 || binary.LittleEndian.Uint16(sell[10:12]) != 5 {
		t.Fatalf("unexpected sell packet items: % X", sell)
	}
}
