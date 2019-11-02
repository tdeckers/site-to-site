package main

import (
	"testing"
)

func TestMerakiAPI(t *testing.T) {
	debug = true
	origPeers, err := GetPeers()
	if err != nil {
		t.Fatalf("failed to get peers: %v", err)
	}

	testPeer := NewPeer("1.1.1.1", []string{"192.168.111.0/24"}, "verysecret")
	testPeer.Name = "test-peer"
	testPeers := append(origPeers, *testPeer)

	if err = UpdatePeers(testPeers); err != nil {
		t.Fatalf("failed to update peers: %v", err)
	}

	newPeers, err := GetPeers()
	if err != nil {
		t.Fatalf("failed to get peers (2): %v", err)
	}

	if len(newPeers) != len(origPeers)+1 {
		t.Fatal("expected an extra peer now.")
	}

	if err = UpdatePeers(origPeers); err != nil {
		t.Fatalf("failed to update back to original peers: %v", err)
	}
}

func TestConfigureMeraki(t *testing.T) {
	debug = true

	testPeerName := "test-peer"
	//
	origPeers, err := GetPeers()
	for _, p := range origPeers {
		if p.Name == testPeerName {
			t.Fatal("test peer already exist.  clean up first.")
		}
	}

	peer := NewPeer("10.1.1.1", []string{"192.168.66.0/24"}, "verysecret")
	peer.Name = "test-peer"
	if err = ConfigureMeraki(peer); err != nil {
		t.Fatal(err)
	}

	peers, err := GetPeers()
	found := false
	for _, p := range peers {
		if p.Name == testPeerName {
			found = true
		}
	}
	if !found {
		t.Fatal("test peer not found after adding")
	}

	if err := UpdatePeers(origPeers); err != nil {
		t.Fatalf("unable to reset original peers: %v", err)
	}
}
