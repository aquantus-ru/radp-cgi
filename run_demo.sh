#!/bin/sh
set -e

# Create cgi-bin directory
mkdir -p www/cgi-bin

# Build the static binary for RADP
echo "Building static binary for radp..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o www/cgi-bin/radp main.go asnmap.go

# Build the static binary for ASNMAP
# It's the same binary but we'll copy/link it or rely on the same build if main.go handles both.
# But since we want to be explicit and match the requirement:
# "incorporate an asnmap utility... that allows asn to ip lookups, using cgi-bin/asnmap?asn=number"
echo "Building static binary for asnmap..."
cp www/cgi-bin/radp www/cgi-bin/asnmap

# Make it executable
chmod +x www/cgi-bin/radp
chmod +x www/cgi-bin/asnmap

# Try to start busybox httpd
if command -v busybox >/dev/null 2>&1; then
    echo "Starting busybox httpd on port 8080..."
    # -f: run in foreground
    # -p: port
    # -h: home directory
    busybox httpd -f -p 8080 -h www
else
    echo "Busybox not found. Trying python3 http.server..."
    if command -v python3 >/dev/null 2>&1; then
         echo "Starting python3 http.server on port 8080..."
         cd www
         python3 -m http.server --cgi 8080
    else
        echo "Neither busybox nor python3 found. Please install one to run the demo server."
        echo "You can still run the binary directly in CLI mode: ./www/cgi-bin/radp -help"
    fi
fi
