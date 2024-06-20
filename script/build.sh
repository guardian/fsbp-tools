#!/usr/bin/env bash

set -e

mkdir -p bin

# Targeting AMD (x86) Mac 
env GOOS=darwin GOARCH=amd64 go build -o bucketblocker-darwin-amd64 main.go
mv bucketblocker-darwin-amd64 bin/

# Targeting ARM64 Mac (Apple silicon) 
env GOOS=darwin GOARCH=arm64 go build -o bucketblocker-darwin-arm64 main.go
mv bucketblocker-darwin-arm64 bin/