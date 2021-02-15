test:
	CGO_ENABLED=0 go test ./...

test-e2e:
	CGO_ENABLED=0 go test ./example/... -v -failfast -tags=mtf -count=1 -p 1 --rebuild_binary=true

gen-mtf-mock:
	protoc --go_out=plugins=grpc:. --go-fff_out=. ./proto/weather/*.proto
	protoc --go_out=plugins=grpc:. --go-fff_out=. ./proto/echo/*.proto
	protoc --go_out=plugins=grpc:. --go-fff_out=. ./proto/oracle/*.proto
	protoc --go_out=plugins=grpc:. --go-fff_out=. ./proto/fswatch/*.proto
	go test ./proto/...

.PHONY: test test-e2e gen-mtf-mock
