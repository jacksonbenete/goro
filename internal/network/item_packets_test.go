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
	data[22] = 0

	ack, ok, err := ParseItemPickupAck(Packet{ID: 0x00A0, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if ack.Index != 12 || ack.Amount != 3 || ack.ItemID != 938 || !ack.Identified || ack.Result != 0 {
		t.Fatalf("unexpected ack: %+v", ack)
	}
}

func TestParseItemPickupAckExtendedVariants(t *testing.T) {
	for _, tc := range []struct {
		id     uint16
		length int
	}{
		{id: 0x029A, length: 27},
		{id: 0x02D4, length: 29},
	} {
		t.Run(fmt.Sprintf("0x%04X", tc.id), func(t *testing.T) {
			data := make([]byte, tc.length)
			binary.LittleEndian.PutUint16(data[0:2], tc.id)
			binary.LittleEndian.PutUint16(data[2:4], 31)
			binary.LittleEndian.PutUint16(data[4:6], 2)
			binary.LittleEndian.PutUint16(data[6:8], 909)
			data[8] = 1
			data[22] = 0

			ack, ok, err := ParseItemPickupAck(Packet{ID: tc.id, Data: data})
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatal("not parsed")
			}
			if ack.Index != 31 || ack.Amount != 2 || ack.ItemID != 909 || !ack.Identified || ack.Result != 0 {
				t.Fatalf("unexpected ack: %+v", ack)
			}
		})
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

func TestParseInventoryItemListNormalLegacy(t *testing.T) {
	data := make([]byte, 4+10)
	binary.LittleEndian.PutUint16(data[0:2], 0x00A3)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 7)
	binary.LittleEndian.PutUint16(data[6:8], 938)
	data[8] = 3
	data[9] = 1
	binary.LittleEndian.PutUint16(data[10:12], 4)

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x00A3, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 7 || got.ItemID != 938 || got.Type != 3 || !got.Identified || got.Amount != 4 || got.Equip {
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
	binary.LittleEndian.PutUint16(data[14:16], 4001)
	binary.LittleEndian.PutUint32(data[22:26], 123456)

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x02E8, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 7 || got.ItemID != 938 || got.Type != 3 || !got.Identified || got.Amount != 4 || got.Equip {
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
	if got := items[0]; got.Index != 11 || got.ItemID != 1201 || got.Type != 4 || !got.Identified || got.Amount != 1 || !got.Equip || !got.Equipped || !got.Damaged || got.Refine != 5 {
		t.Fatalf("unexpected item: %+v", got)
	}
}

func TestParseInventoryItemListEquipment2008(t *testing.T) {
	data := make([]byte, 4+28)
	binary.LittleEndian.PutUint16(data[0:2], 0x02D0)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[4:6], 11)
	binary.LittleEndian.PutUint16(data[6:8], 1201)
	data[8] = 4
	data[9] = 1
	binary.LittleEndian.PutUint16(data[12:14], 0x0002)
	data[14] = 1
	data[15] = 5
	binary.LittleEndian.PutUint32(data[24:28], 123456)
	binary.LittleEndian.PutUint16(data[28:30], 1)
	binary.LittleEndian.PutUint16(data[30:32], 2)

	items, ok, err := ParseInventoryItemList(Packet{ID: 0x02D0, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(items) != 1 {
		t.Fatalf("items ok=%v len=%d", ok, len(items))
	}
	if got := items[0]; got.Index != 11 || got.ItemID != 1201 || got.Type != 4 || !got.Identified || got.Amount != 1 || !got.Equip || !got.Equipped || !got.Damaged || got.Refine != 5 {
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
