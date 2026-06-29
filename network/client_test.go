package network

import (
	"net"
	"testing"
	"time"
)

func TestClientCloseDoesNotReportReadLoopError(t *testing.T) {
	client := NewClient(20080910, false)
	local, remote := net.Pipe()
	defer remote.Close()

	client.mu.Lock()
	client.conn = local
	client.mu.Unlock()
	done := make(chan struct{})
	go func() {
		client.readLoop(local)
		close(done)
	}()

	client.Close()
	select {
	case <-done:
		if errs := client.DrainErrors(); len(errs) > 0 {
			t.Fatalf("intentional close produced errors: %v", errs)
		}
	case <-time.After(time.Second):
		t.Fatal("read loop did not exit")
	}
}
