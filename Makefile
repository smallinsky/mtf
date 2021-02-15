test:
	CGO_ENABLED=0 go test ./...

test-e2e:
	CGO_ENABLED=0 go test ./example/... -v -failfast -tags=mtf -count=1 -p 1 --rebuild_binary=true

gen-mtf-mock:
	protoc --go_out=plugins=grpc:. --go-mtf-mock_out=. ./proto/weather/*.proto
	protoc --go_out=plugins=grpc:. --go-mtf-mock_out=. ./proto/echo/*.proto
	protoc --go_out=plugins=grpc:. --go-mtf-mock_out=. ./proto/oracle/*.proto
	go test ./proto/...

.PHONY: test test-e2e gen-mtf-mock
