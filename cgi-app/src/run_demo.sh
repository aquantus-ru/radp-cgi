#!/bin/sh
set -e

# Create cgi-bin directory
mkdir -p www/cgi-bin

# Build the static binary for CGI
echo "Building static binary for CGI (radp and asnmap)..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o www/cgi-bin/radp ./cmd/cgi
cp www/cgi-bin/radp www/cgi-bin/asnmap

# Build the standalone server
echo "Building standalone server..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

# Make them executable
chmod +x www/cgi-bin/radp
chmod +x www/cgi-bin/asnmap
chmod +x server

# Copy binaries to root for convenience
cp www/cgi-bin/radp ../radp
cp www/cgi-bin/asnmap ../asnmap
cp server ../server

echo "Build complete."
