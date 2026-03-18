Dependecy:

sudo apt install -y \
  libssl-dev \
  libssh-dev \
  libcurl4-openssl-dev \
  pkg-config


RUN:

clear && CGO_ENABLED=1 LD_LIBRARY_PATH=$PWD/.local/lib go run ./cmd/netconf-client/main.go