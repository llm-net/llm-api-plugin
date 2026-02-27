GO := go
GOFLAGS := -trimpath
BIN_DIR := bin

TOOLS := gemini-cli ark-cli topview-cli jimeng-cli

.PHONY: all build clean $(TOOLS)

all: build

build: $(TOOLS)

gemini-cli:
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$@ ./cmd/gemini-cli/

ark-cli:
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$@ ./cmd/ark-cli/

topview-cli:
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$@ ./cmd/topview-cli/

jimeng-cli:
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$@ ./cmd/jimeng-cli/

# Cross-compile all tools for release
.PHONY: release
release:
	@for tool in $(TOOLS); do \
		for os in linux darwin windows; do \
			for arch in amd64 arm64; do \
				echo "Building $$tool-$$os-$$arch..."; \
				GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) -o dist/$$tool-$$os-$$arch ./cmd/$$tool/; \
			done; \
		done; \
	done

clean:
	rm -rf $(BIN_DIR)/ dist/
