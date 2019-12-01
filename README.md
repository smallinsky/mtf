[![CircleCI](https://circleci.com/gh/smallinsky/mtf.svg?style=svg)](https://circleci.com/gh/smallinsky/mtf)[![Go Report Card](https://goreportcard.com/badge/github.com/smallinsky/mtf)](https://goreportcard.com/report/github.com/smallinsky/mtf)
 # Microservice Test Framework
## Introduction
This Microservice Test Framework (MTF) allows in simple way to mock service dependencies and setup docker test environment   in a comprehensive way.

Supported dependencies:
* GRPC client/server communication
* Google Cloud Pubsub
* Google Cloud Storage (Partial - only bucket object Insert/Get operation)
* FTP
* HTTP/HTTPS integration
* MySQL
* Redis

The aim of the MTF framework is not only cut dependencies require to run your binary and execute tests but also focus on test readability.




## Getting started
To begin using `MTF` framework you will need to define `TestMain` function and setup required environment used for binary that you would like to test further called `SUT` (System Under Test). To setup prerequisite dependency for your SUT the `framework.TestEnv` function should be invoked. `With...()` functions method chain allows to set up and configure dependency for SUT. Finally the `.Run()` chain method function starts the test.
```go
func TestMain(m *testing.M) {
	framework.TestEnv(m).
		WithSUT(framework.SutSettings{
			Dir:   "./service", // dir with source files of system under test.
			Ports: []int{8001},
			Envs: []string{
				"ORACLE_ADDR=" + framework.GetDockerHostAddr(8002),
			}}).
		WithRedis(framework.RedisSettings{
			Password: "test",
		}).
		WithMySQL(framework.MysqlSettings{
			DatabaseName: "test_db",
			MigrationDir: "./service/migrations",
			Password:     "test",
		}).Run()
}
```

SuiteTest collects and groups collection of ports that allows to communicate with external mocked dependency.
```go
type SuiteTest struct {
	echoPort   *port.Port
	httpPort   *port.Port
	oraclePort *port.Port
}
```
Ports initialization should be done within Suite `Init(t *testing.T)` function.
```go
func (st *SuiteTest) Init(t *testing.T) {
	var err error
	if st.echoPort, err = port.NewGRPCClientPort((*pb.EchoClient)(nil), "localhost:8001"); err != nil {
		t.Fatalf("failed to init grpc client port")
	}
	st.httpPort = port.NewHTTPPort()
	if st.oraclePort, err = port.NewGRPCServerPort((*pbo.OracleServer)(nil), ":8002"); err != nil {
		t.Fatalf("failed to init grpc oracle server")
	}
}
```

## Ports
Port are used to communicate with dependencies by sending and receiving messages in a consistent way.
### GRPC Client/Server port `port.NewGRPCServerPort` `port.NewGRPCClientPort`
GRPC ports allows to mock whole grpc communication between client<->SUT<->Other GRPC service.

Server grpc port initialization:
```
oraclePort, err = port.NewGRPCServerPort((*pbo.OracleServer)(nil), ":8002")
```

Client port initialization:
```
echoPort, err = port.NewGRPCClientPort((*pb.EchoClient)(nil), "localhost:8001")
```

Example of sut gprc method handler:
```go
func (s *server) AskOracle(ctx context.Context, req *pb.AskOracleRequest) (*pb.AskOracleResponse, error) {
	resp, err := s.OracleClient.AskDeepThought(ctx, &pbo.AskDeepThoughtRequest{
		Data: req.GetData(),
	})
	if err != nil {
		switch status.Code(err) {
		case codes.FailedPrecondition:
			return &pb.AskOracleResponse{
				Data: "Come back after seven and a half million years",
			}, nil
		default:
			return nil, err
		}
	}
	return &pb.AskOracleResponse{
		Data: resp.GetData(),
	}, nil
}
```
Test GRPC message flow:
```go
func (st *SuiteTest) TestClientServerGRPC(t *testing.T) {
	st.echoPort.Send(t, &pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(t, &pbo.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(t, &pbo.AskDeepThoughtResponse{
		Data: "42",
	})
	st.echoPort.Receive(t, &pb.AskOracleResponse{
		Data: "42",
	})
}
```
```
--- PASS: TestEchoService (0.75s)
    --- PASS: TestEchoService/TestClientServerGRPC (0.01s)
PASS
```
If we change the handler body by adding additional text to AskOracleResponse.Date:
```diff
@@ -110,6 +110,6 @@ func (s *server) AskOracle(ctx context.Context, req *pb.AskOracleRequest) (*pb.A
                return nil, err
        }
        return &pb.AskOracleResponse{
-               Data: resp.GetData(),
+               Data: resp.GetData() + "!!!",
        }, nil
 }
```
The `echoPort.Receive(t, &pb.AskOracleResponse{...}` call will log port mismatch
```
--- FAIL: TestEchoService/TestClientServerGRPC (0.01s)
port.go:89: Failed to receive *echo.AskOracleResponse:
     deep equal:
     got:'{
     "data": "42!!!!"
    }'
     exp: '{
     "data": "42"
    }'
    : match not eq
```
Mock GRPC error response:
```go
func (st *SuiteTest) TestClientServerGRPCError(t *testing.T) {
	st.echoPort.Send(t, &pb.AskOracleRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Receive(t, &pbo.AskDeepThoughtRequest{
		Data: "Get answer for ultimate question of life the universe and everything",
	})
	st.oraclePort.Send(t, &port.GRPCErr{
		Err: status.Errorf(codes.FailedPrecondition, "Deepthought error"),
	})
	st.echoPort.Receive(t, &pb.AskOracleResponse{
		Data: "Come back after seven and a half million years",
	})
}
```
## HTTP/HTTPS Port `port.NewHTTPPort()`
HTTP port allows to test external http endpoint integration by matching SUT's http requests and sending back custom shape responses.

```go
func (st *SuiteTest) TestHTTP(t *testing.T) {
	st.httpPort.Receive(t, &port.HTTPRequest{
		Body:   []byte{},
		Method: "GET",
		Host:   "example.com",
		URL:    "/urlpath",
	})

	st.httpPort.Send(t, &port.HTTPResponse{
		Body:   []byte(`{"value":{"joke":"42"}}`),
		Status: http.StatusOK,
	})
}
```


### Metchers
Match only message type:
```go
echoPort.Receive(t, match.Type(&pb.AskOracleResponse{})
```
Match by custom function:
```go
echoPort.Receive(t, match.Fn(func(resp *pb.AskGoogleResponse) {
		if got, want := len(resp.GetData()), 2; got != want {
			t.Fatalf("data len mismatch, got: %v want: %v", got, want)
		}
	}))
```
Match GRPC Error status code:
```
echoPort.Receive(t, match.GRPCStatusCode(codes.Internal))
```
## MTF Tests execution
Right now MTF framework does not support parallel test execution and to prevent simultaneously test run passing the  `-p 1` flag to `go test` command is required.  
### Run tests examples:
```bash
go test ./example/... -p 1 --rebuild_binary=true  -tags=mtf
```

### Test Environment preparation phase
At first run the mtf will download docker images dependency needed to prepare and run test environment:
```
=== PREPARING TEST ENV
  - Starting [REDIS Component] -  1.601356005s
  - Starting [MYSQL Component] -  50.75908ms
  - Starting [MIGRATE Component] -  805.090778ms
  - Starting [SUT Component] -  5.895084273s
=== TEST RUN DONE - 11.863770827s
```
