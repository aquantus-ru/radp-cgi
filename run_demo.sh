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

# Decide which version to run for the demo.
# Default to busybox/python CGI, but we could run the server.
echo "To run the standalone server, execute: ./server"

# Try to start busybox httpd
if command -v busybox >/dev/null 2>&1; then
    echo "Starting busybox httpd (CGI version) on port 8080..."
    # -f: run in foreground
    # -p: port
    # -h: home directory
    busybox httpd -f -p 8080 -h www
else
    echo "Busybox not found. Trying python3 http.server..."
    if command -v python3 >/dev/null 2>&1; then
         echo "Starting python3 http.server (CGI version) on port 8080..."
         cd www
         python3 -m http.server --cgi 8080
    else
        echo "Neither busybox nor python3 found."
        echo "Starting the standalone web server instead..."
        ./server
    fi
fi
