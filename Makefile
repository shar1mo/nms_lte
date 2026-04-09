ROOT_DIR            := $(abspath .)
THIRD_PARTY_DIR     := $(ROOT_DIR)/third_party
BUILD_DIR           := $(ROOT_DIR)/.build
PREFIX_DIR          := $(ROOT_DIR)/.local
FRONTEND_DIR		:= $(ROOT_DIR)/frontend
FRONTEND_EMBED_DIR	:= $(ROOT_DIR)/cmd/nms-lte/frontend
FRONTEND_EMBED_DIST	:= $(FRONTEND_EMBED_DIR)/dist
FRONTEND_FALLBACK_DIST := $(THIRD_PARTY_DIR)/nms-front/dist

LIBYANG_REPO        := https://github.com/CESNET/libyang.git
LIBNETCONF2_REPO    := https://github.com/CESNET/libnetconf2.git
FRONTEND_REPO		:= https://github.com/JoraBarjomi/nms-front.git

LIBYANG_REF         ?= master
LIBNETCONF2_REF     ?= master
FRONTEND_REF		?= main

LIBYANG_SRC         := $(THIRD_PARTY_DIR)/libyang
LIBNETCONF2_SRC     := $(THIRD_PARTY_DIR)/libnetconf2
FRONTEND_SRC		:= $(FRONTEND_DIR)/nms-front

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
	@mkdir -p "$(THIRD_PARTY_DIR)" "$(BUILD_DIR)" "$(PREFIX_DIR)" "$(FRONTEND_DIR)" "$(FRONTEND_EMBED_DIR)"

test: frontend-assets
	$(GOENV) $(GO) test -v ./...

run-nms: libnetconf2 libyang frontend-assets
	$(GOENV) $(GO) run ./cmd/nms-lte

build-nms: libnetconf2 libyang frontend-assets
	$(GOENV) $(GO) build -o nms-lte ./cmd/nms-lte

swagger:
	swag init -g main.go -d cmd/nms-lte,internal/httpapi -o docs/swagger --parseInternal
	
run-netconf:
	$(GOENV) $(GO) run ./cmd/netconf-client

bootstrap: clone libnetconf2 libyang

clone: clone-libyang clone-libnetconf2 clone-frontend

clone-frontend: deps
	@if [ ! -d "$(FRONTEND_SRC)/.git" ]; then \
		$(GIT) clone --depth 1 --branch "$(FRONTEND_REF)" "$(FRONTEND_REPO)" "$(FRONTEND_SRC)"; \
	else \
		echo "Frontend already cloned: $(FRONTEND_SRC)"; \
	fi

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

frontend: clone-frontend
	cd "$(FRONTEND_SRC)" && rm -rf node_modules package-lock.json && npm install --prefer-offline --no-audit --no-fund && npm run build
	$(MAKE) frontend-assets

frontend-assets: clone-frontend
	@if [ -f "$(FRONTEND_SRC)/dist/index.html" ]; then \
		src="$(FRONTEND_SRC)/dist"; \
	elif [ -f "$(FRONTEND_FALLBACK_DIST)/index.html" ]; then \
		src="$(FRONTEND_FALLBACK_DIST)"; \
		echo "Using cached frontend dist: $(FRONTEND_FALLBACK_DIST)"; \
	else \
		echo "Frontend dist not found. Run 'make frontend' after fixing npm."; \
		exit 1; \
	fi; \
	rm -rf "$(FRONTEND_EMBED_DIST)"; \
	mkdir -p "$(FRONTEND_EMBED_DIST)"; \
	cp -R "$$src/." "$(FRONTEND_EMBED_DIST)/"

libyang: $(LIBYANG_BUILD)/CMakeCache.txt
	$(CMAKE) --build "$(LIBYANG_BUILD)" --parallel
	$(CMAKE) --install "$(LIBYANG_BUILD)"

libnetconf2: libyang $(LIBNETCONF2_BUILD)/CMakeCache.txt
	$(CMAKE) --build "$(LIBNETCONF2_BUILD)" --parallel
	$(CMAKE) --install "$(LIBNETCONF2_BUILD)"

clean:
	rm -rf "$(BUILD_DIR)" "$(FRONTEND_EMBED_DIST)"

distclean: clean
	rm -rf "$(PREFIX_DIR)"

.PHONY: all bootstrap deps clone clone-frontend clone-libyang clone-libnetconf2 \
	update-libyang update-libnetconf2 \
	test run-nms build-nms swagger run-netconf frontend frontend-assets \
	libyang libnetconf2 clean distclean
