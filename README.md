# Microservice Test Framework [![CircleCI](https://circleci.com/gh/smallinsky/mtf.svg?style=svg)](https://circleci.com/gh/smallinsky/mtf)[![Go Report Card](https://goreportcard.com/badge/github.com/smallinsky/mtf)](https://goreportcard.com/report/github.com/smallinsky/mtf)
## Introduction
This Microservice Test Framework allows in simple way to mock "all" dependency by microservice and test it in a comprehensive way.    

Supported dependencies:
* GRPC client/server communication
* Google Cloud Pubsub 
* Google Cloud Storage (Partial - only bucket object Insert/Get operation)
* FTP 
* HTTP/HTTPS integration
* MySQL
* Redis

## Getting started
To begin using `MTF` framework you will need to define `TestMain` function and setup required environment used for binary that you would like to test further called `SUT` (System Under Test). To setup prerequisite dependency for your SUT the `framework.TestEnv` function should be invoked. `With...()` functions method chain allows to set up and configure dependency for SUT. Finally the `.Run()` chain method function starts the test.
```go
func TestMain(m *testing.M) {
	framework.TestEnv(m).
		WithSUT(framework.SutSettings{
			Dir:   "./service", // dir with source file of system under test.
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

Suite collects ports that allows to communicate with external mocked dependency by calling port.`{send receive}` functions. 
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
	st.httpPort = port.NewHTTP2Port()
	if st.oraclePort, err = port.NewGRPCServerPort((*pbo.OracleServer)(nil), ":8002"); err != nil {
		t.Fatalf("failed to init grpc oracle server")
	}
}
```

Tests function needs to be implemented on suite object.
```go
func (st *SuiteTest) TestRedis(t *testing.T) {
	st.echoPort.Send(t, &pb.AskRedisRequest{
		Data: "make me sandwitch",
	})
	st.echoPort.Receive(t, &pb.AskRedisResponse{
		Data: "what? make it yourself",
	})
	st.echoPort.Send(t, &pb.AskRedisRequest{
		Data: "sudo make me sandwitch",
	})
	st.echoPort.Receive(t, &pb.AskRedisResponse{
		Data: "okey",
	})
}
```
```go
func (st *SuiteTest) TestClientServerGRPC(t *testing.T) {
	time.Sleep(time.Second * 2)
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
```go
func (st *SuiteTest) TestFetchDataFromDB(t *testing.T) {
	st.echoPort.Send(t, &pb.AskDBRequest{
		Data: "the dirty fork",
	})
	st.echoPort.Receive(t, &pb.AskDBResponse{
		Data: "Lucky we didn't say anything about the dirty knife",
	})
}
```
