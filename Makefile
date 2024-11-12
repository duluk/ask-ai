# Makefile for the ask-ai project.

# GOPATH ?= $(HOME)/go
ifeq ($(GOPATH),)
    GOPATH := $(HOME)/go
    $(info GOPATH not set, using default: $(GOPATH))
endif

BINARY_DIR := bin
INSTALL_DIR := $(GOPATH)/bin
MAIN_BINARY := ask-ai

CMD_FILES := $(wildcard cmd/*.go)
BIN_FILES := $(patsubst cmd/%.go,%,$(CMD_FILES))
PKG_FILES := $(wildcard LLM/*.go)
TST_FILES := $(wildcard test/*.go)

CP := $(shell which cp)
GO := $(shell which go)

CPFLAGS := --preserve=mode,timestamps
# GOFLAGS := -asm
TESTFLAGS := -v -cover -coverprofile=coverage.out

$(shell mkdir -p $(BINARY_DIR))

all: $(addprefix $(BINARY_DIR)/,$(BIN_FILES))

$(BINARY_DIR)/%: cmd/%.go vet
	$(GO) build $(GOFLAGS) -o $@ $<

list:
	@echo "CMD_FILES: $(CMD_FILES)"
	@echo "BIN_FILES: $(BIN_FILES)"
	@echo "PKG_FILES: $(PKG_FILES)"
	@echo "TST_FILES: $(TST_FILES)"

clean:
	rm -rf $(BINARY_DIR)/* coverage.out

test: $(TST_FILES)
	$(GO) test $(TESTFLAGS) ./test || exit 1

fmt:
	$(GO) fmt ./...

vet: fmt
	$(GO) vet ./...

run:
	./$(BINARY_DIR)/$(MAIN_BINARY)

install: all
	@mkdir -p $(INSTALL_DIR)
	@for app in $(BIN_FILES); do \
		$(CP) $(CPFLAGS) $(BINARY_DIR)/$$app $(INSTALL_DIR); \
		echo "Installed $$app to $(INSTALL_DIR)"; \
	done

.PHONY: all list clean test fmt vet run install
