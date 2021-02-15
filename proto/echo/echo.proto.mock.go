package echo

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

    
func NewEchoClientMock(target string) EchoClient {
    conn, err := grpc.Dial(target, grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    return &EchoClientMock{
        cc: NewEchoClient(conn),
    }
}

func NewEchoServerMock() *EchoServiceMock {
    ln, err := net.Listen("tcp", ":0")
    if err != nil {
        panic(err)
    }

    s := grpc.NewServer()

    mock := &EchoServiceMock{
        addr: ln.Addr().String(),
        server: &MockEchoServer{
            RepeatHandler: make(chan HandleRepeat),
            ScreamHandler: make(chan HandleScream),
            AskGoogleHandler: make(chan HandleAskGoogle),
            AskDBHandler: make(chan HandleAskDB),
            AskRedisHandler: make(chan HandleAskRedis),
            AskOracleHandler: make(chan HandleAskOracle),
        },
    }

    RegisterEchoServer(s, mock.server)
    go func() {
        if err := s.Serve(ln); err != nil {
            panic(err)
        }
    }()
    return mock
}

type MockEchoServer struct {
    RepeatHandler chan HandleRepeat
    ScreamHandler chan HandleScream
    AskGoogleHandler chan HandleAskGoogle
    AskDBHandler chan HandleAskDB
    AskRedisHandler chan HandleAskRedis
    AskOracleHandler chan HandleAskOracle
}


type HandleRepeat func(ctx context.Context, req *RepeatRequest) (*RepeatResponse, error)

func (s *MockEchoServer) Repeat (ctx context.Context, req *RepeatRequest) (*RepeatResponse, error) {
    select {
        case handler := <-s.RepeatHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*RepeatRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*RepeatResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "Repeat timeout").Err()
    }
}
type HandleScream func(ctx context.Context, req *ScreamRequest) (*ScreamResponse, error)

func (s *MockEchoServer) Scream (ctx context.Context, req *ScreamRequest) (*ScreamResponse, error) {
    select {
        case handler := <-s.ScreamHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*ScreamRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*ScreamResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "Scream timeout").Err()
    }
}
type HandleAskGoogle func(ctx context.Context, req *AskGoogleRequest) (*AskGoogleResponse, error)

func (s *MockEchoServer) AskGoogle (ctx context.Context, req *AskGoogleRequest) (*AskGoogleResponse, error) {
    select {
        case handler := <-s.AskGoogleHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*AskGoogleRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*AskGoogleResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "AskGoogle timeout").Err()
    }
}
type HandleAskDB func(ctx context.Context, req *AskDBRequest) (*AskDBResponse, error)

func (s *MockEchoServer) AskDB (ctx context.Context, req *AskDBRequest) (*AskDBResponse, error) {
    select {
        case handler := <-s.AskDBHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*AskDBRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*AskDBResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "AskDB timeout").Err()
    }
}
type HandleAskRedis func(ctx context.Context, req *AskRedisRequest) (*AskRedisResponse, error)

func (s *MockEchoServer) AskRedis (ctx context.Context, req *AskRedisRequest) (*AskRedisResponse, error) {
    select {
        case handler := <-s.AskRedisHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*AskRedisRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*AskRedisResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "AskRedis timeout").Err()
    }
}
type HandleAskOracle func(ctx context.Context, req *AskOracleRequest) (*AskOracleResponse, error)

func (s *MockEchoServer) AskOracle (ctx context.Context, req *AskOracleRequest) (*AskOracleResponse, error) {
    select {
        case handler := <-s.AskOracleHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*AskOracleRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*AskOracleResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "AskOracle timeout").Err()
    }
}

type EchoServiceMock struct {
    addr   string
    server *MockEchoServer
}

func (s *EchoServiceMock) Addr() string {
    return s.addr
}


func (s *EchoServiceMock) Repeat(fn HandleRepeat) {
    go func() {
        select {
            case s.server.RepeatHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}

func (s *EchoServiceMock) Scream(fn HandleScream) {
    go func() {
        select {
            case s.server.ScreamHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}

func (s *EchoServiceMock) AskGoogle(fn HandleAskGoogle) {
    go func() {
        select {
            case s.server.AskGoogleHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}

func (s *EchoServiceMock) AskDB(fn HandleAskDB) {
    go func() {
        select {
            case s.server.AskDBHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}

func (s *EchoServiceMock) AskRedis(fn HandleAskRedis) {
    go func() {
        select {
            case s.server.AskRedisHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}

func (s *EchoServiceMock) AskOracle(fn HandleAskOracle) {
    go func() {
        select {
            case s.server.AskOracleHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}


type EchoClientMock struct {
    cc EchoClient
}

func (s *EchoClientMock) Repeat(ctx context.Context, in *RepeatRequest, opts ...grpc.CallOption) (*RepeatResponse, error) {
    resp, err := s.cc.Repeat(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}

func (s *EchoClientMock) Scream(ctx context.Context, in *ScreamRequest, opts ...grpc.CallOption) (*ScreamResponse, error) {
    resp, err := s.cc.Scream(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}

func (s *EchoClientMock) AskGoogle(ctx context.Context, in *AskGoogleRequest, opts ...grpc.CallOption) (*AskGoogleResponse, error) {
    resp, err := s.cc.AskGoogle(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}

func (s *EchoClientMock) AskDB(ctx context.Context, in *AskDBRequest, opts ...grpc.CallOption) (*AskDBResponse, error) {
    resp, err := s.cc.AskDB(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}

func (s *EchoClientMock) AskRedis(ctx context.Context, in *AskRedisRequest, opts ...grpc.CallOption) (*AskRedisResponse, error) {
    resp, err := s.cc.AskRedis(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}

func (s *EchoClientMock) AskOracle(ctx context.Context, in *AskOracleRequest, opts ...grpc.CallOption) (*AskOracleResponse, error) {
    resp, err := s.cc.AskOracle(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}


