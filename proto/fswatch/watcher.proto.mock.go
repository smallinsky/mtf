package watcher

import (
    "runtime"
    "net"
    "time"
    "context"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/golang/protobuf/ptypes/empty"
)

type grpcHandler func(context.Context, interface{}) (interface{}, error)
type grpcResp struct {
    msg interface{}
    err error
}

func handle(ctx context.Context, req interface{}, handler grpcHandler) (interface{}, error) {
    resultC := make(chan grpcResp)
    exitC := make(chan struct{})
    go func() {
        defer func() { exitC <- struct{}{} }()
        resp, err := handler(ctx, req)
        resultC <- grpcResp{msg: resp, err: err}
    }()

    select {
        case r := <-resultC:
            return r.msg, r.err
        case <-exitC:
            return nil, status.New(codes.Internal, "test failed").Err()
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "test timeout").Err()
    }
}

func returnIfTestFailed(err error) {
    if err == nil {
        return
    }
    st, ok := status.FromError(err)
    if !ok {
        return
    }
    if st.Err().Error() == "rpc error: code = Internal desc = test failed" {
        runtime.Goexit()
    }
}

    
func NewWatcherClientMock(target string) WatcherClient {
    conn, err := grpc.Dial(target, grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    return &WatcherClientMock{
        cc: NewWatcherClient(conn),
    }
}

func NewWatcherServerMock() *WatcherServiceMock {
    ln, err := net.Listen("tcp", ":0")
    if err != nil {
        panic(err)
    }

    s := grpc.NewServer()

    mock := &WatcherServiceMock{
        addr: ln.Addr().String(),
        server: &MockWatcherServer{
            EventHandler: make(chan HandleEvent),
        },
    }

    RegisterWatcherServer(s, mock.server)
    go func() {
        if err := s.Serve(ln); err != nil {
            panic(err)
        }
    }()
    return mock
}

type MockWatcherServer struct {
    EventHandler chan HandleEvent
}


type HandleEvent func(ctx context.Context, req *EventRequest) (*empty.Empty, error)

func (s *MockWatcherServer) Event (ctx context.Context, req *EventRequest) (*empty.Empty, error) {
    select {
        case handler := <-s.EventHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*EventRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*empty.Empty), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "Event timeout").Err()
    }
}

type WatcherServiceMock struct {
    addr   string
    server *MockWatcherServer
}

func (s *WatcherServiceMock) Addr() string {
    return s.addr
}


func (s *WatcherServiceMock) Event(fn HandleEvent) {
    go func() {
        select {
            case s.server.EventHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}


type WatcherClientMock struct {
    cc WatcherClient
}

func (s *WatcherClientMock) Event(ctx context.Context, in *EventRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
    resp, err := s.cc.Event(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}


