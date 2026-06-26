package res

import "testing"

func TestParseClientInfo(t *testing.T) {
	info, err := ParseClientInfo([]byte(`
<clientinfo>
  <servicetype>korea</servicetype>
  <connection>
    <display>TestRO</display>
    <address>127.0.0.1</address>
    <port>6900</port>
    <version>55</version>
    <langtype>0</langtype>
  </connection>
</clientinfo>`))
	if err != nil {
		t.Fatal(err)
	}
	if info.ServiceType != "korea" {
		t.Fatalf("service type = %q", info.ServiceType)
	}
	if len(info.Connections) != 1 {
		t.Fatalf("connections = %d", len(info.Connections))
	}
	conn := info.Connections[0]
	if conn.Display != "TestRO" || conn.Address != "127.0.0.1" || conn.Port != 6900 || conn.Version != 55 {
		t.Fatalf("unexpected connection: %+v", conn)
	}
}
