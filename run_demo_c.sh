#!/bin/sh
set -e

# Build the C binaries
echo "Building C binaries..."
cd c_port
make
cd ..

# Create cgi-bin directory
mkdir -p www/cgi-bin

# Copy the binaries to the CGI directory
cp c_port/radp www/cgi-bin/radp
cp c_port/asnmap www/cgi-bin/asnmap

# Make them executable
chmod +x www/cgi-bin/radp
chmod +x www/cgi-bin/asnmap

# Try to start busybox httpd
if command -v busybox >/dev/null 2>&1; then
    echo "Starting busybox httpd (C version) on port 8080..."
    # -f: run in foreground
    # -p: port
    # -h: home directory
    busybox httpd -f -p 8080 -h www
else
    echo "Busybox not found. Trying python3 http.server..."
    if command -v python3 >/dev/null 2>&1; then
         echo "Starting python3 http.server (C version) on port 8080..."
         cd www
         python3 -m http.server --cgi 8080
    else
        echo "Neither busybox nor python3 found."
        echo "Please install one to run the C CGI server."
    fi
fi
