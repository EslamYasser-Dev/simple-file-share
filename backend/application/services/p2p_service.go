package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// P2P signal kinds exchanged over the signaling channel (SDP / ICE payloads).
const (
	P2PSignalOffer     = "offer"
	P2PSignalAnswer    = "answer"
	P2PSignalCandidate = "candidate"
	P2PSignalBye       = "bye"
)

// P2P event kinds delivered on the per-peer SSE stream.
const (
	P2PEventHello  = "hello"
	P2PEventPeers  = "peers"
	P2PEventSignal = "signal"
)

// ErrPeerOffline is returned when the destination peer has no live stream.
var ErrPeerOffline = errors.New("peer offline")

// ErrInvalidSignal is returned for a bad destination or unknown signal kind.
var ErrInvalidSignal = errors.New("invalid signal")

// P2PPeer is one connected browser session (tab) ready for WebRTC signaling.
type P2PPeer struct {
	ID   string `json:"id"`
	User string `json:"user,omitempty"`
}

// P2PSignal is one SDP/ICE/bye frame relayed between peers. Payload is opaque
// to the server (SDP string or RTCIceCandidateInit JSON).
type P2PSignal struct {
	From    string          `json:"from"`
	To      string          `json:"to"`
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload,omitempty"`
	At      time.Time       `json:"at"`
}

// P2PEvent is a frame on the peer's Server-Sent Events stream.
type P2PEvent struct {
	Type   string     `json:"type"`
	PeerID string     `json:"peerId,omitempty"`
	Peers  []P2PPeer  `json:"peers,omitempty"`
	Signal *P2PSignal `json:"signal,omitempty"`
	At     time.Time  `json:"at"`
}

const (
	p2pSubscriberBuffer = 64
)

type p2pSession struct {
	peer P2PPeer
	ch   chan P2PEvent
}

// P2PService is an in-memory presence + signaling hub. Bytes never pass
// through the server after signaling; WebRTC peers exchange file data
// directly (LAN or via host/STUN candidates).
type P2PService struct {
	mu       sync.Mutex
	sessions map[string]*p2pSession
}

func NewP2PService() *P2PService {
	return &P2PService{sessions: make(map[string]*p2pSession)}
}

// Subscribe registers a live peer session until ctx is cancelled or stop is
// called. The returned channel receives hello, peer-list, and signal events.
func (s *P2PService) Subscribe(ctx context.Context, user string) (peerID string, ch <-chan P2PEvent, stop func()) {
	if s == nil {
		closed := make(chan P2PEvent)
		close(closed)
		return "", closed, func() {}
	}

	id, err := randomPeerID()
	if err != nil {
		// Fallback keeps streaming available if the RNG fails (never expected).
		id = "peer-fallback-" + time.Now().UTC().Format("150405.000000000")
	}
	peer := P2PPeer{ID: id, User: user}
	out := make(chan P2PEvent, p2pSubscriberBuffer)
	sess := &p2pSession{peer: peer, ch: out}

	s.mu.Lock()
	s.sessions[id] = sess
	peers := s.peersLocked()
	s.mu.Unlock()

	// Hello first so the client learns its own id before any peer updates.
	safeSend(out, P2PEvent{Type: P2PEventHello, PeerID: id, Peers: peersFor(peers, id), At: time.Now().UTC()})
	s.broadcastPeers(peers)

	var once sync.Once
	unregister := func() {
		once.Do(func() {
			s.mu.Lock()
			if cur, ok := s.sessions[id]; ok && cur == sess {
				delete(s.sessions, id)
				close(out)
				s.broadcastPeersLocked(s.peersLocked())
			}
			s.mu.Unlock()
		})
	}

	if ctx != nil && ctx.Done() != nil {
		go func() {
			<-ctx.Done()
			unregister()
		}()
	}
	return id, out, unregister
}

// ListPeers returns every connected session except excludeID (usually self).
func (s *P2PService) ListPeers(excludeID string) []P2PPeer {
	if s == nil {
		return []P2PPeer{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return peersFor(s.peersLocked(), excludeID)
}

// OwnsPeer reports whether peerID is a live session for user. An empty user
// (auth disabled) only matches sessions that also registered anonymously.
func (s *P2PService) OwnsPeer(peerID, user string) bool {
	if s == nil || peerID == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[peerID]
	if !ok {
		return false
	}
	return sess.peer.User == user
}

// SendSignal relays one signaling frame from → to. The destination must have
// a live stream; otherwise ErrPeerOffline is returned. Payload must already
// be valid JSON.
func (s *P2PService) SendSignal(from, to, kind string, payload json.RawMessage) error {
	if s == nil {
		return ErrPeerOffline
	}
	if from == "" || to == "" || from == to {
		return ErrInvalidSignal
	}
	if !validP2PSignalKind(kind) {
		return ErrInvalidSignal
	}
	if len(payload) > 0 && !json.Valid(payload) {
		return ErrInvalidSignal
	}

	sig := P2PSignal{
		From:    from,
		To:      to,
		Kind:    kind,
		Payload: payload,
		At:      time.Now().UTC(),
	}
	ev := P2PEvent{Type: P2PEventSignal, Signal: &sig, At: sig.At}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[from]; !ok {
		return ErrInvalidSignal
	}
	sess, ok := s.sessions[to]
	if !ok {
		return ErrPeerOffline
	}
	return safeSend(sess.ch, ev)
}

// Stop tears down the hub (used by tests).
func (s *P2PService) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		delete(s.sessions, id)
		close(sess.ch)
	}
}

func (s *P2PService) peersLocked() []P2PPeer {
	out := make([]P2PPeer, 0, len(s.sessions))
	for _, sess := range s.sessions {
		out = append(out, sess.peer)
	}
	return out
}

func (s *P2PService) broadcastPeers(peers []P2PPeer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.broadcastPeersLocked(peers)
}

func (s *P2PService) broadcastPeersLocked(peers []P2PPeer) {
	// Each client filters self out of the shared list on receive; we still
	// send the full list so join/leave is consistent for every subscriber.
	for _, sess := range s.sessions {
		safeSend(sess.ch, P2PEvent{Type: P2PEventPeers, Peers: peers, At: time.Now().UTC()})
	}
}

func peersFor(all []P2PPeer, excludeID string) []P2PPeer {
	out := make([]P2PPeer, 0, len(all))
	for _, p := range all {
		if p.ID == excludeID {
			continue
		}
		out = append(out, p)
	}
	return out
}

func safeSend(ch chan P2PEvent, ev P2PEvent) error {
	select {
	case ch <- ev:
		return nil
	default:
		// Slow consumer: drop presence floods rather than stalling senders.
		// Signals are not dropped silently when the buffer is full — report.
		if ev.Type == P2PEventSignal {
			return ErrPeerOffline
		}
		return nil
	}
}

func validP2PSignalKind(kind string) bool {
	switch kind {
	case P2PSignalOffer, P2PSignalAnswer, P2PSignalCandidate, P2PSignalBye:
		return true
	default:
		return false
	}
}

func randomPeerID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
