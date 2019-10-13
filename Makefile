test:
	CGO_ENABLED=0 go test ./...

test-e2e:
	CGO_ENABLED=0 go test ./example/... -v -failfast -tags=mtf -count=1 -p 1 --rebuild_binary=true
