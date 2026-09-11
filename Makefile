.PHONY: server run test project clean

BUILD_DIR := .build/bin
SERVER_BINARY := $(BUILD_DIR)/nestplay-server

server:
	mkdir -p $(BUILD_DIR)
	cd server && go build -o ../$(SERVER_BINARY) .

run: server
	./$(SERVER_BINARY)

test:
	cd server && go test ./...
	cd packages/HealthClient && swift test

project:
	cd apps/tvos && xcodegen generate

clean:
	rm -rf .build packages/HealthClient/.build apps/tvos/DerivedData
