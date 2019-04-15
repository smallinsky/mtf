run_e2e:
	cd service/echo && GOOS=linux GOARCH=amd64 go build
	cd service/oracle && GOOS=linux GOARCH=amd64 go build
	docker-compose up

proto:
	protoc --go_out=plugins=grpc:. proto/echo/*.proto
	protoc --go_out=plugins=grpc:. proto/oracle/*.proto


.PHONY: proto, run_e2e
