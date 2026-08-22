all: fmt vet lint test build

build:
	go build .

install:
	go install .
	go run . completion --install

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test --timeout 10s ./...

generate:
	go generate ./...
