#!/bin/bash

platforms=(
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "darwin/amd64"
  "darwin/arm64"
)

BUILD_BIN_DIR="./bin"
BIN_SRC_DIR="./cmd"
MYAPP_SRC="ask-ai.go"
MYAPP=$(basename "${MYAPP_SRC}" .go)

for platform in "${platforms[@]}"; do
  OS="${platform%/*}"
  ARCH="${platform#*/}"
  echo "Building $MYAPP for $OS/$ARCH"
  output_name="${BUILD_BIN_DIR}/${MYAPP}_${OS}_${ARCH}"
  if [[ "$OS" == "windows" ]]; then
    output_name="${output_name}.exe"
  fi
  GOOS="$OS" GOARCH="$ARCH" go build -o "$output_name" "${BIN_SRC_DIR}/${MYAPP_SRC}"
done
