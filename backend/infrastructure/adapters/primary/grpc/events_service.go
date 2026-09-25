package grpcapi

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// EventsService streams live application events to native subscribers,
// mirroring the HTTP Server-Sent Events handler for gRPC clients.
type EventsService struct {
	filesharev1.UnimplementedEventsServiceServer

	bus *events.Bus
}

func NewEventsService(bus *events.Bus) *EventsService {
	return &EventsService{bus: bus}
}

func (s *EventsService) Subscribe(req *filesharev1.SubscribeRequest, stream filesharev1.EventsService_SubscribeServer) error {
	if s.bus == nil {
		return status.Error(codes.Unavailable, "events unavailable")
	}
	subscriber := authctx.UserFromContext(stream.Context())
	ch, stop := s.bus.Subscribe(stream.Context())
	defer stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case e, open := <-ch:
			if !open {
				return nil
			}
			if !events.ShouldDeliver(e, subscriber) {
				continue
			}
			msg := &filesharev1.EventMessage{
				Type:  e.Type,
				Path:  e.Path,
				User:  e.User,
				Bytes: e.Bytes,
			}
			if !e.At.IsZero() {
				msg.At = e.At.UTC().Format(time.RFC3339)
			}
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}
