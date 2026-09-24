package dto

import "encoding/json"

// P2PSignalRequest is the body of POST /api/p2p/signal.
type P2PSignalRequest struct {
	To      string          `json:"to"`
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// P2PPeer is one connected peer session exposed to the UI.
type P2PPeer struct {
	ID   string `json:"id"`
	User string `json:"user,omitempty"`
}

// P2PSignal is a relayed SDP/ICE/bye frame.
type P2PSignal struct {
	From    string          `json:"from"`
	To      string          `json:"to"`
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload,omitempty"`
	At      string          `json:"at,omitempty"`
}

// P2PEvent is one Server-Sent Events frame on /api/p2p/stream.
type P2PEvent struct {
	Type   string     `json:"type"`
	PeerID string     `json:"peerId,omitempty"`
	Peers  []P2PPeer  `json:"peers,omitempty"`
	Signal *P2PSignal `json:"signal,omitempty"`
	At     string     `json:"at,omitempty"`
}

// P2PPeersResponse is the body of GET /api/p2p/peers.
type P2PPeersResponse struct {
	Peers []P2PPeer `json:"peers"`
	Self  string    `json:"self,omitempty"`
}
