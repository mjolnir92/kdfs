package network

import (
	"testing"
	"github.com/mjolnir92/kdfs/kademliaid"
	"github.com/mjolnir92/kdfs/contact"
)

func TestHello(t *testing.T) {
	t.Error("hello test")
}

func TestPing(t *testing.T) {
	id_server := &kademliaid.New("0000000000000000000000000000000000000000")
	nw_server := New(150, id_server)
	ct_server := &New(id_server, "localhost:12300")
	go nw_server.Listen("localhost", "12300")
	// TODO: clean up the Listen goroutine
	id_client := &kademliaid.New("1000000000000000000000000000000000000000")
	nw_client := New(150, id_client)
	err := nw_client.Ping(ct_server)
	if err != nil {
		t.Error("Ping returned an error:", v)
	}
	// TODO: was the routing table updated?
}