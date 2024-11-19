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
TST_FILES := $(wildcard test/*.go)

CP := $(shell which cp)
GO := $(shell which go)

CPFLAGS := -p
GOFLAGS := -ldflags "-X 'main.commit=$(shell git rev-parse --short HEAD)' -X 'main.date=$(shell date -u '+%Y-%m-%d %H:%M:%S')'"
TESTFLAGS := -v -cover -coverprofile=coverage.out

$(shell mkdir -p $(BINARY_DIR))

all: check build

build: $(addprefix $(BINARY_DIR)/,$(BIN_FILES))

$(BINARY_DIR)/%: cmd/%.go
	$(GO) build $(GOFLAGS) -o $@ $<

list:
	@echo "CMD_FILES: $(CMD_FILES)"
	@echo "BIN_FILES: $(BIN_FILES)"
	@echo "TST_FILES: $(TST_FILES)"

clean:
	rm -f $(addprefix $(BINARY_DIR)/,$(BIN_FILES)) coverage.out
	# rm -rf $(BINARY_DIR)/* coverage.out

test: $(TST_FILES)
	$(GO) test $(TESTFLAGS) ./test || exit 1

check: fmt vet

fmt: $(CMD_FILES)
	$(GO) fmt ./...

vet: $(CMD_FILES) fmt
	$(GO) vet ./...

run: $(BINARY_DIR)/$(MAIN_BINARY)
	./$(BINARY_DIR)/$(MAIN_BINARY)

install: all
	@mkdir -p $(INSTALL_DIR)
	@for app in $(BIN_FILES); do \
		$(CP) $(CPFLAGS) $(BINARY_DIR)/$$app $(INSTALL_DIR); \
		echo "Installed $$app to $(INSTALL_DIR)"; \
	done

# Futurte self: This ensures that make treats the targets as labels and not
# files. This is important because if a file of the same name actually exists,
# it may not be executed if the timestamp hasn't changed. That's not what we
# want for these.
.PHONY: all list check clean test fmt vet run install
