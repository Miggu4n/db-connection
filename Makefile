build:
	@go build -o bin/db-connection

run: build
	./bin/db-connection

test:
	go test -v ./...
