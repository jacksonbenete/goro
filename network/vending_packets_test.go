package network

import (
	"encoding/binary"
	"testing"
)

func TestParseVendingPackets(t *testing.T) {
	open := []byte{0x2D, 0x01, 0x04, 0x00}
	req, ok, err := ParseVendingOpenRequest(Packet{ID: PacketZCOpenStore, Data: open})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || req.MaxItems != 4 {
		t.Fatalf("open request ok=%v req=%+v", ok, req)
	}

	board := make([]byte, 86)
	binary.LittleEndian.PutUint16(board[0:2], PacketZCStoreEntry)
	binary.LittleEndian.PutUint32(board[2:6], 0x11223344)
	copy(board[6:], "Cheap pots")
	entry, ok, err := ParseVendingBoard(Packet{ID: PacketZCStoreEntry, Data: board})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || entry.OwnerAID != 0x11223344 || entry.Name != "Cheap pots" {
		t.Fatalf("board ok=%v entry=%+v", ok, entry)
	}

	disappear := []byte{0x32, 0x01, 0x44, 0x33, 0x22, 0x11}
	closed, ok, err := ParseVendingBoardDisappear(Packet{ID: PacketZCDisappearEntry, Data: disappear})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || closed.OwnerAID != 0x11223344 {
		t.Fatalf("disappear ok=%v value=%+v", ok, closed)
	}
}

func TestParseVendingItemLists(t *testing.T) {
	list := make([]byte, 8+22)
	binary.LittleEndian.PutUint16(list[0:2], PacketZCPCPurchaseItemListFromMC)
	binary.LittleEndian.PutUint16(list[2:4], uint16(len(list)))
	binary.LittleEndian.PutUint32(list[4:8], 0x11223344)
	binary.LittleEndian.PutUint32(list[8:12], 1500)
	binary.LittleEndian.PutUint16(list[12:14], 3)
	binary.LittleEndian.PutUint16(list[14:16], 9)
	list[16] = 5
	binary.LittleEndian.PutUint16(list[17:19], 2301)
	list[19] = 1
	list[21] = 7
	binary.LittleEndian.PutUint16(list[22:24], 4001)
	items, ok, err := ParseVendingItemList(Packet{ID: PacketZCPCPurchaseItemListFromMC, Data: list})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || items.OwnerAID != 0x11223344 || len(items.Items) != 1 || items.Items[0].Index != 9 || items.Items[0].Amount != 3 || items.Items[0].ItemID != 2301 || items.Items[0].Price != 1500 || items.Items[0].Cards[0] != 4001 {
		t.Fatalf("vendor list ok=%v items=%+v", ok, items)
	}

	own := make([]byte, 8+22)
	binary.LittleEndian.PutUint16(own[0:2], PacketZCPCPurchaseMyItemList)
	binary.LittleEndian.PutUint16(own[2:4], uint16(len(own)))
	binary.LittleEndian.PutUint32(own[4:8], 2000000)
	binary.LittleEndian.PutUint32(own[8:12], 900)
	binary.LittleEndian.PutUint16(own[12:14], 12)
	binary.LittleEndian.PutUint16(own[14:16], 5)
	own[16] = 4
	binary.LittleEndian.PutUint16(own[17:19], 501)
	ownItems, ok, err := ParseVendingItemList(Packet{ID: PacketZCPCPurchaseMyItemList, Data: own})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !ownItems.Own || ownItems.Items[0].Index != 12 || ownItems.Items[0].Amount != 5 {
		t.Fatalf("own list ok=%v items=%+v", ok, ownItems)
	}
}

func TestParseVendingResultPackets(t *testing.T) {
	resultData := []byte{0x35, 0x01, 0x09, 0x00, 0x02, 0x00, 0x04}
	result, ok, err := ParseVendingPurchaseResult(Packet{ID: PacketZCPCPurchaseResultFromMC, Data: resultData})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || result.Index != 9 || result.Amount != 2 || result.Result != 4 {
		t.Fatalf("result ok=%v value=%+v", ok, result)
	}

	soldData := []byte{0x37, 0x01, 0x0C, 0x00, 0x03, 0x00}
	sold, ok, err := ParseVendingSoldItem(Packet{ID: PacketZCDeleteItemFromMCStore, Data: soldData})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || sold.Index != 12 || sold.Amount != 3 {
		t.Fatalf("sold ok=%v value=%+v", ok, sold)
	}
}

func TestBuildVendingPackets(t *testing.T) {
	closeStore := BuildCloseVendingStorePacket()
	if len(closeStore) != 2 || ID(closeStore) != PacketCZReqCloseStore {
		t.Fatalf("close packet: % X", closeStore)
	}

	open := BuildOpenVendingStorePacket("Shop", []VendingOpenItem{{Index: 7, Amount: 2, Price: 1500}})
	if len(open) != 92 || ID(open) != PacketCZReqOpenStore || binary.LittleEndian.Uint16(open[2:4]) != 92 {
		t.Fatalf("open packet header: % X", open[:8])
	}
	if string(open[4:8]) != "Shop" || binary.LittleEndian.Uint16(open[84:86]) != 7 || binary.LittleEndian.Uint16(open[86:88]) != 2 || binary.LittleEndian.Uint32(open[88:92]) != 1500 {
		t.Fatalf("open packet body: % X", open)
	}

	req := BuildVendingListRequestPacket(0x11223344)
	if len(req) != 6 || ID(req) != PacketCZReqBuyFromMC || binary.LittleEndian.Uint32(req[2:6]) != 0x11223344 {
		t.Fatalf("list request: % X", req)
	}

	buy := BuildVendingPurchasePacket(0x11223344, []VendingPurchaseItem{{Index: 9, Amount: 2}})
	if len(buy) != 12 || ID(buy) != PacketCZPCPurchaseItemListFromMC || binary.LittleEndian.Uint16(buy[2:4]) != 12 || binary.LittleEndian.Uint32(buy[4:8]) != 0x11223344 {
		t.Fatalf("buy packet header: % X", buy)
	}
	if binary.LittleEndian.Uint16(buy[8:10]) != 2 || binary.LittleEndian.Uint16(buy[10:12]) != 9 {
		t.Fatalf("buy packet body: % X", buy)
	}
}
