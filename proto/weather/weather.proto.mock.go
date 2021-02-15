package weather

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

    
func NewWeatherClientMock(target string) WeatherClient {
    conn, err := grpc.Dial(target, grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    return &WeatherClientMock{
        cc: NewWeatherClient(conn),
    }
}

func NewWeatherServerMock(addr string) *WeatherServiceMock {
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        panic(err)
    }

    s := grpc.NewServer()

    mock := &WeatherServiceMock{
        addr: ln.Addr().String(),
        server: &MockWeatherServer{
            AskAboutWeatherHandler: make(chan HandleAskAboutWeather),
        },
    }

    RegisterWeatherServer(s, mock.server)
    go func() {
        if err := s.Serve(ln); err != nil {
            panic(err)
        }
    }()
    return mock
}

type MockWeatherServer struct {
    AskAboutWeatherHandler chan HandleAskAboutWeather
}


type HandleAskAboutWeather func(ctx context.Context, req *AskAboutWeatherRequest) (*AskAboutWeatherResponse, error)

func (s *MockWeatherServer) AskAboutWeather (ctx context.Context, req *AskAboutWeatherRequest) (*AskAboutWeatherResponse, error) {
    select {
        case handler := <-s.AskAboutWeatherHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*AskAboutWeatherRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*AskAboutWeatherResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "AskAboutWeather timeout").Err()
    }
}

type WeatherServiceMock struct {
    addr   string
    server *MockWeatherServer
}

func (s *WeatherServiceMock) Addr() string {
    return s.addr
}


func (s *WeatherServiceMock) AskAboutWeather(fn HandleAskAboutWeather) {
    go func() {
        select {
            case s.server.AskAboutWeatherHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}


type WeatherClientMock struct {
    cc WeatherClient
}

func (s *WeatherClientMock) AskAboutWeather(ctx context.Context, in *AskAboutWeatherRequest, opts ...grpc.CallOption) (*AskAboutWeatherResponse, error) {
    resp, err := s.cc.AskAboutWeather(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}


    
func NewScaleConvClientMock(target string) ScaleConvClient {
    conn, err := grpc.Dial(target, grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    return &ScaleConvClientMock{
        cc: NewScaleConvClient(conn),
    }
}

func NewScaleConvServerMock(addr string) *ScaleConvServiceMock {
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        panic(err)
    }

    s := grpc.NewServer()

    mock := &ScaleConvServiceMock{
        addr: ln.Addr().String(),
        server: &MockScaleConvServer{
            CelsiusToFahrenheitHandler: make(chan HandleCelsiusToFahrenheit),
        },
    }

    RegisterScaleConvServer(s, mock.server)
    go func() {
        if err := s.Serve(ln); err != nil {
            panic(err)
        }
    }()
    return mock
}

type MockScaleConvServer struct {
    CelsiusToFahrenheitHandler chan HandleCelsiusToFahrenheit
}


type HandleCelsiusToFahrenheit func(ctx context.Context, req *CelsiusToFahrenheitRequest) (*CelsiusToFahrenheitResponse, error)

func (s *MockScaleConvServer) CelsiusToFahrenheit (ctx context.Context, req *CelsiusToFahrenheitRequest) (*CelsiusToFahrenheitResponse, error) {
    select {
        case handler := <-s.CelsiusToFahrenheitHandler:
            fn  := func(ctx context.Context, req interface{}) (interface{}, error) {
                return handler(ctx, req.(*CelsiusToFahrenheitRequest))
            }
            resp, err := handle(ctx, req, fn)
            if err != nil {
                return nil, err
            }
            return resp.(*CelsiusToFahrenheitResponse), nil
        case <-time.Tick(time.Second * 5):
            return nil, status.New(codes.Unavailable, "CelsiusToFahrenheit timeout").Err()
    }
}

type ScaleConvServiceMock struct {
    addr   string
    server *MockScaleConvServer
}

func (s *ScaleConvServiceMock) Addr() string {
    return s.addr
}


func (s *ScaleConvServiceMock) CelsiusToFahrenheit(fn HandleCelsiusToFahrenheit) {
    go func() {
        select {
            case s.server.CelsiusToFahrenheitHandler <- fn:
        return
            case <-time.Tick(time.Second * 5):
        return
        }
    }()
}


type ScaleConvClientMock struct {
    cc ScaleConvClient
}

func (s *ScaleConvClientMock) CelsiusToFahrenheit(ctx context.Context, in *CelsiusToFahrenheitRequest, opts ...grpc.CallOption) (*CelsiusToFahrenheitResponse, error) {
    resp, err := s.cc.CelsiusToFahrenheit(ctx, in, grpc.WaitForReady(true))
    returnIfTestFailed(err)
    return resp, err
}


