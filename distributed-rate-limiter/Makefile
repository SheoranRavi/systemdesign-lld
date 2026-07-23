GO             ?= go
BINARY         ?= drl
BIN_DIR        ?= bin
OUTPUT         := $(BIN_DIR)/$(BINARY)

.PHONY: build clean

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(OUTPUT) ./cmd

clean:
	rm -f $(OUTPUT)
