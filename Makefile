build:
	go build .

install:
	go install .
	go run . completion --install

lint:
	golangci-lint run
