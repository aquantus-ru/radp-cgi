# RDAP & ASN Map Tool - CGI & Self-Hosting Version

This application allows you to perform RDAP lookups for domains, IPs, and ASNs, as well as fetch ASN-to-IP prefix mapping from RADB.

## Contents
- **Precompiled Binaries:** Precompiled standalone binaries will be placed in the root of this folder upon build.
- **Source Code:** Available in the `src/` directory.

## Getting Started (Quick Start)

You can run this application in two main ways: using the self-hosted web server, or via CGI with an existing web server (like Apache or busybox).

### 1. Building the Binaries

To compile the binaries from source code, run the build script:

```bash
cd src
./run_demo.sh
```

This will compile the Go applications statically and copy the binaries to the `cgi-app/` directory for easy access.

### 2. Using the Self-Hosting Server

The easiest way to start using the tool is with the built-in standalone server. Note that you should execute the binary from within `cgi-app/src` so it can find the `www` folder.

```bash
cd src
./server
```

This starts a server on port `8080`. Open your browser and navigate to `http://localhost:8080`.

### 3. Using as a CGI script

You can host this tool using any CGI-compatible web server. The binaries `radp` and `asnmap` act as CGI scripts when executed in a CGI environment.

For a quick demo using Python:
```bash
cd src/www
python3 -m http.server --cgi 8080
```
Then visit `http://localhost:8080`.

### How to use the lookup tool
1. Open the web interface in your browser.
2. Select the **Query Type**:
   - **Domain (RDAP):** Get WHOIS/RDAP info for a domain (e.g., `example.com`).
   - **IP (RDAP):** Get ownership info for an IP address (e.g., `8.8.8.8`).
   - **Autnum (ASN) (RDAP):** Get registration info for an ASN (e.g., `AS15169`).
   - **ASN to IP Map (RADB):** Get all routed IP prefixes for an ASN (e.g., `AS15169`).
3. Enter the target in the **Value** field and click **Query**. The JSON results will be displayed below.
