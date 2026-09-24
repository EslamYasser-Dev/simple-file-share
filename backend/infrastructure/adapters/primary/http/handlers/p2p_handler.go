package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// P2PHandler exposes WebRTC signaling (presence + SDP/ICE relay):
//
//	GET  /api/p2p/stream   SSE: hello, peer list, incoming signals
//	GET  /api/p2p/peers    connected peers (excluding self)
//	POST /api/p2p/signal   relay one offer/answer/candidate/bye to a peer
//
// File bytes never traverse this server after signaling; peers exchange
// data on a WebRTC DataChannel (direct on LAN when candidates allow).
type P2PHandler struct {
	svc *services.P2PService
}

func NewP2PHandler(svc *services.P2PService) *P2PHandler {
	return &P2PHandler{svc: svc}
}

func (h *P2PHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/p2p/stream":
		h.stream(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/p2p/peers":
		h.peers(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/p2p/signal":
		h.signal(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *P2PHandler) stream(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		respondError(w, http.StatusServiceUnavailable, "p2p unavailable")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		respondError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	user := ""
	if u := currentUser(r); u != nil {
		user = u.Username
	}

	peerID, ch, stop := h.svc.Subscribe(r.Context(), user)
	defer stop()

	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(sseHeartbeat)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case ev, open := <-ch:
			if !open {
				return
			}
			frame, err := json.Marshal(toP2PEventDTO(ev, peerID))
			if err != nil {
				continue
			}
			// Named events keep the browser client's EventSource routing simple.
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, frame); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *P2PHandler) peers(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		respondError(w, http.StatusServiceUnavailable, "p2p unavailable")
		return
	}
	raw := h.svc.ListPeers("")
	peers := make([]dto.P2PPeer, 0, len(raw))
	for _, p := range raw {
		peers = append(peers, dto.P2PPeer{ID: p.ID, User: p.User})
	}
	respondJSON(w, http.StatusOK, dto.P2PPeersResponse{Peers: peers})
}

func (h *P2PHandler) signal(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		respondError(w, http.StatusServiceUnavailable, "p2p unavailable")
		return
	}
	var req dto.P2PSignalRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	from := r.Header.Get("X-Peer-Id")
	if from == "" {
		respondError(w, http.StatusBadRequest, "X-Peer-Id header is required")
		return
	}
	user := ""
	if u := currentUser(r); u != nil {
		user = u.Username
	}
	if !h.svc.OwnsPeer(from, user) {
		respondError(w, http.StatusForbidden, "unknown peer session")
		return
	}
	if req.To == "" || req.Kind == "" {
		respondError(w, http.StatusBadRequest, "to and kind are required")
		return
	}
	if err := h.svc.SendSignal(from, req.To, req.Kind, req.Payload); err != nil {
		switch err {
		case services.ErrPeerOffline:
			respondError(w, http.StatusNotFound, "peer offline")
		case services.ErrInvalidSignal:
			respondError(w, http.StatusBadRequest, "invalid signal")
		default:
			respondWithError(w, err)
		}
		return
	}
	respondJSON(w, http.StatusAccepted, dto.MessageResponse{Message: "queued"})
}

func toP2PEventDTO(ev services.P2PEvent, selfID string) dto.P2PEvent {
	out := dto.P2PEvent{
		Type:   ev.Type,
		PeerID: ev.PeerID,
		At:     ev.At.UTC().Format(time.RFC3339Nano),
	}
	if ev.Type == services.P2PEventHello {
		out.PeerID = selfID
	}
	if ev.Peers != nil {
		out.Peers = make([]dto.P2PPeer, 0, len(ev.Peers))
		for _, p := range ev.Peers {
			if p.ID == selfID {
				continue
			}
			out.Peers = append(out.Peers, dto.P2PPeer{ID: p.ID, User: p.User})
		}
	}
	if ev.Signal != nil {
		out.Signal = &dto.P2PSignal{
			From:    ev.Signal.From,
			To:      ev.Signal.To,
			Kind:    ev.Signal.Kind,
			Payload: ev.Signal.Payload,
			At:      ev.Signal.At.UTC().Format(time.RFC3339Nano),
		}
	}
	return out
}
