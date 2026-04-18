BINARY  := kubectl-refx
BIN_DIR := ./bin

.PHONY: build install test lint clean

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) .

install:
	@test -f $(BIN_DIR)/$(BINARY) || (echo "Run 'make build' first"; exit 1)
	cp $(BIN_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed /usr/local/bin/$(BINARY)"

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
