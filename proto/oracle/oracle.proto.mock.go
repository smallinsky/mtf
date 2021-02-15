package oracle

import (
    "runtime"
    "net"
    "time"
    "context"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    
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

    
func NewOracleClientMock(target string) OracleClient {
    conn, err := grpc.Dial(target, grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    return &OracleClientMock{
        cc: NewOracleClient(conn),
    }
}

func NewOracleServerMock() *OracleServiceMock {
    ln, err := net.Listen("tcp", ":0")
    if err != nil {
        panic(err)
    }

    s := grpc.NewServer()

    mock := &OracleServiceMock{
        addr: ln.Addr().String(),
        server: &MockOracleServer{
            AskDeepThoughtHandler: make(chan HandleAskDeepThought),
        },
    }

    RegisterOracleServer(s, mock.server)
    go func() {
        if err := s.Serve(ln); err != nil {
            panic(err)
        }
    }()
    return mock
}

type MockOracleServer struct {
    AskDeepThoughtHandler chan HandleAskDeepThought
}


type HandleAskDeepThought func(ctx context.Context, req *AskDeepThoughtRequest) (*AskDeepThoughtResponse, error)

func (s *MockOracleServer) AskDeepThought (ctx context.Context, req *AskDeepThoughtRequest) (*AskDeepThoughtResponse, error) {
    select {
        case handler := <-s.AskDeepThoughtHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*AskDeepThoughtRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*AskDeepThoughtResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "AskDeepThought timeout").Err()
    }
}

type OracleServiceMock struct {
    addr   string
    server *MockOracleServer
}

func (s *OracleServiceMock) Addr() string {
    return s.addr
}


func (s *OracleServiceMock) AskDeepThought(fn HandleAskDeepThought) {
    go func() {
        select {
            case s.server.AskDeepThoughtHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}


type OracleClientMock struct {
    cc OracleClient
}

func (s *OracleClientMock) AskDeepThought(ctx context.Context, in *AskDeepThoughtRequest, opts ...grpc.CallOption) (*AskDeepThoughtResponse, error) {
    resp, err := s.cc.AskDeepThought(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}


