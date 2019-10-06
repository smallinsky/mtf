package fswatch

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/smallinsky/mtf/proto/fswatch"
)

func Subscriber(port string, handler func(event *pb.EventRequest)) {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("[ERR] (fswatcher) failed to listen: %v", err)
	}
	s := grpc.NewServer()
	svc := subService{
		handler: handler,
	}
	pb.RegisterWatcherServer(s, &svc)

	if err := s.Serve(l); err != nil {
		log.Fatalf("[ERR] (fswatcher) server stopped with err: %v", err)
	}
}

type subService struct {
	handler func(event *pb.EventRequest)
}

func (ss *subService) Event(ctx context.Context, req *pb.EventRequest) (*empty.Empty, error) {
	if ss.handler != nil {
		ss.handler(req)
	}
	return &empty.Empty{}, nil
}
