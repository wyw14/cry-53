.PHONY: build test race vet run web-test web-build
build:
	go build ./...
test:
	go test ./...
race:
	go test -race ./...
vet:
	go vet ./...
run:
	go run ./cmd/server
web-test:
	cd web && npm test
web-build:
	cd web && npm run build

