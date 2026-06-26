package res

import "testing"

func TestParseMsgStringTable(t *testing.T) {
	table := parseMsgStringTable([]byte("// ignored\nDo you agree?#\r\nDisconnected from Server.#\n"))
	if got := table[0]; got != "Do you agree?" {
		t.Fatalf("table[0] = %q", got)
	}
	if got := table[1]; got != "Disconnected from Server." {
		t.Fatalf("table[1] = %q", got)
	}
}

func TestParseMsgStringCSV(t *testing.T) {
	table := parseMsgStringCSV([]byte("TVNJX0E=,SGVsbG8=\nTVNJX0I=,WW91IGdvdCAlZCBpdGVtcy4=\n"))
	if got := table[0]; got != "Hello" {
		t.Fatalf("table[0] = %q", got)
	}
	if got := table[1]; got != "You got %d items." {
		t.Fatalf("table[1] = %q", got)
	}
}
