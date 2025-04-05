# !/bin/bash

# Build for Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o build/wechat-gif-windows-amd64.exe --ldflags "-s -w" -trimpath

# Build for macOS (amd64)
GOOS=darwin GOARCH=amd64 go build -o build/wechat-gif-darwin-amd64 --ldflags "-s -w" -trimpath

# Build for Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o build/wechat-gif-linux-amd64 --ldflags "-s -w" -trimpath
