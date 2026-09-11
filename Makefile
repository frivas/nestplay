.PHONY: server run test project clean tools check check-server check-tvos

BUILD_DIR := .build/bin
SERVER_BINARY := $(BUILD_DIR)/nestplay-server
GOBIN := $(CURDIR)/.build/bin
export PATH := $(GOBIN):$(PATH)
export GOTOOLCHAIN := go1.26.6

XCODE_DESTINATION := generic/platform=tvOS
XCODE_FLAGS := CODE_SIGNING_ALLOWED=NO SWIFT_TREAT_WARNINGS_AS_ERRORS=YES

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

tools:
	mkdir -p $(GOBIN)
	GOBIN=$(GOBIN) go install honnef.co/go/tools/cmd/staticcheck@2026.1
	GOBIN=$(GOBIN) go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
	GOBIN=$(GOBIN) go install github.com/securego/gosec/v2/cmd/gosec@v2.29.0

check: check-server check-tvos

check-server: tools
	cd server && { \
		unformatted=$$(gofmt -l .); \
		if [ -n "$$unformatted" ]; then \
			echo "The following files need gofmt:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi; \
	}
	cd server && go mod tidy
	cd server && { \
		if [ -n "$$(git status --porcelain -- go.mod go.sum)" ]; then \
			git diff -- go.mod go.sum; \
			echo "Run 'go mod tidy' in server and commit the result."; \
			exit 1; \
		fi; \
	}
	cd server && go vet ./...
	cd server && staticcheck ./...
	mkdir -p .build
	cd server && go test -race -covermode=atomic -coverprofile=$(CURDIR)/.build/coverage.out ./...
	mkdir -p $(BUILD_DIR)
	cd server && go build -trimpath -o ../$(SERVER_BINARY) .
	cd server && govulncheck -test ./...
	cd server && gosec -exclude-generated ./...

check-tvos:
	xcrun swift-format lint --recursive --parallel apps/tvos packages/HealthClient
	swift test --package-path packages/HealthClient --parallel
	$(MAKE) project
	@if [ -n "$$(git status --porcelain -- apps/tvos/Nestplay.xcodeproj)" ]; then \
		git diff -- apps/tvos/Nestplay.xcodeproj; \
		echo "Run 'make project' and commit the generated Xcode project."; \
		exit 1; \
	fi
	plutil -lint apps/tvos/Nestplay/Info.plist
	xcodebuild analyze \
		-project apps/tvos/Nestplay.xcodeproj \
		-scheme Nestplay \
		-destination '$(XCODE_DESTINATION)' \
		-derivedDataPath "$(CURDIR)/.build/derived/analyze" \
		$(XCODE_FLAGS)
	xcodebuild build \
		-project apps/tvos/Nestplay.xcodeproj \
		-scheme Nestplay \
		-destination '$(XCODE_DESTINATION)' \
		-derivedDataPath "$(CURDIR)/.build/derived/build" \
		$(XCODE_FLAGS)

clean:
	rm -rf .build packages/HealthClient/.build apps/tvos/DerivedData
