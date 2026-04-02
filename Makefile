ROOT_DIR            := $(abspath .)
THIRD_PARTY_DIR     := $(ROOT_DIR)/third_party
BUILD_DIR           := $(ROOT_DIR)/.build
PREFIX_DIR          := $(ROOT_DIR)/.local

LIBYANG_REPO        := https://github.com/CESNET/libyang.git
LIBNETCONF2_REPO    := https://github.com/CESNET/libnetconf2.git

LIBYANG_REF         ?= master
LIBNETCONF2_REF     ?= master

LIBYANG_SRC         := $(THIRD_PARTY_DIR)/libyang
LIBNETCONF2_SRC     := $(THIRD_PARTY_DIR)/libnetconf2

LIBYANG_BUILD       := $(BUILD_DIR)/libyang
LIBNETCONF2_BUILD   := $(BUILD_DIR)/libnetconf2

CMAKE               := cmake
GIT                 := git
BUILD_TYPE          ?= Release

GO 					:= go
GOENV				:= CGO_ENABLED=1 CGO_LDFLAGS="-L$(PWD)/.local/lib -Wl,-rpath,$(CURDIR)/.local/lib"



COMMON_CMAKE_FLAGS := \
	-DCMAKE_BUILD_TYPE=$(BUILD_TYPE) \
	-DCMAKE_INSTALL_PREFIX=$(PREFIX_DIR) \
	-DCMAKE_PREFIX_PATH=$(PREFIX_DIR)

LIBYANG_CMAKE_FLAGS := \
	$(COMMON_CMAKE_FLAGS)

LIBNETCONF2_CMAKE_FLAGS := \
	$(COMMON_CMAKE_FLAGS) \
	-DENABLE_TESTS=OFF \
	-DENABLE_DNSSEC=OFF

all: libnetconf2 libyang

build-netconf-cli: all
	$(GOENV) $(GO) build -o netconf-client ./cmd/netconf-client

deps:
	@mkdir -p "$(THIRD_PARTY_DIR)" "$(BUILD_DIR)" "$(PREFIX_DIR)"

test:
	$(GOENV) $(GO) test -v ./...

run-nms:
	$(GOENV) $(GO) run ./cmd/nms-rc

run-netconf:
	$(GOENV) $(GO) run ./cmd/netconf-client

bootstrap: clone libnetconf2 libyang

clone: clone-libyang clone-libnetconf2

clone-libyang: deps
	@if [ ! -d "$(LIBYANG_SRC)/.git" ]; then \
		$(GIT) clone --depth 1 --branch "$(LIBYANG_REF)" "$(LIBYANG_REPO)" "$(LIBYANG_SRC)"; \
	else \
		echo "libyang already cloned: $(LIBYANG_SRC)"; \
	fi

clone-libnetconf2: deps
	@if [ ! -d "$(LIBNETCONF2_SRC)/.git" ]; then \
		$(GIT) clone --depth 1 --branch "$(LIBNETCONF2_REF)" "$(LIBNETCONF2_REPO)" "$(LIBNETCONF2_SRC)"; \
	else \
		echo "libnetconf2 already cloned: $(LIBNETCONF2_SRC)"; \
	fi

update-libyang:
	@test -d "$(LIBYANG_SRC)/.git" || (echo "Missing git repo: $(LIBYANG_SRC)"; exit 1)
	cd "$(LIBYANG_SRC)" && \
		$(GIT) fetch --tags origin && \
		$(GIT) checkout "$(LIBYANG_REF)" && \
		$(GIT) pull --ff-only origin "$(LIBYANG_REF)"

update-libnetconf2:
	@test -d "$(LIBNETCONF2_SRC)/.git" || (echo "Missing git repo: $(LIBNETCONF2_SRC)"; exit 1)
	cd "$(LIBNETCONF2_SRC)" && \
		$(GIT) fetch --tags origin && \
		$(GIT) checkout "$(LIBNETCONF2_REF)" && \
		$(GIT) pull --ff-only origin "$(LIBNETCONF2_REF)"

$(LIBYANG_BUILD)/CMakeCache.txt: clone-libyang
	mkdir -p "$(LIBYANG_BUILD)"
	$(CMAKE) -S "$(LIBYANG_SRC)" -B "$(LIBYANG_BUILD)" $(LIBYANG_CMAKE_FLAGS)

$(LIBNETCONF2_BUILD)/CMakeCache.txt: $(LIBYANG_BUILD)/CMakeCache.txt clone-libnetconf2
	mkdir -p "$(LIBNETCONF2_BUILD)"
	$(CMAKE) -S "$(LIBNETCONF2_SRC)" -B "$(LIBNETCONF2_BUILD)" $(LIBNETCONF2_CMAKE_FLAGS)

libyang: $(LIBYANG_BUILD)/CMakeCache.txt
	$(CMAKE) --build "$(LIBYANG_BUILD)" --parallel
	$(CMAKE) --install "$(LIBYANG_BUILD)"

libnetconf2: libyang $(LIBNETCONF2_BUILD)/CMakeCache.txt
	$(CMAKE) --build "$(LIBNETCONF2_BUILD)" --parallel
	$(CMAKE) --install "$(LIBNETCONF2_BUILD)"

clean:
	rm -rf "$(BUILD_DIR)"

distclean: clean
	rm -rf "$(PREFIX_DIR)"

.PHONY: all bootstrap deps clone clone-libyang clone-libnetconf2 \
	update-libyang update-libnetconf2 \
	libyang libnetconf2 clean distclean