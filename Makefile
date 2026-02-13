.PHONY: build test install clean ui-check

BINARY_NAME=zk
CMD_PATH=./cmd/zk/main.go

build:
	go build -o $(BINARY_NAME) $(CMD_PATH)

test:
	@echo "Running Go tests..."
	go test -v ./internal/...
	@echo "Validating CUE configuration..."
	cue vet config.cue

install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

ui-check:
	@echo "Checking Lua syntax..."
	luacheck lua/

clean:
	rm -f $(BINARY_NAME)
	rm -rf .zk_index/