.PHONY: server run test project

server:
	cd server && go build -o ../bin/nestplay-server .

run: server
	./bin/nestplay-server

test:
	cd server && go test ./...
	cd HealthClient && swift test

project:
	cd tvos && xcodegen generate
