package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/openrdap/rdap"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	// Check if we are running as "asnmap" (via symlink or rename)
	// or if the first argument is "asnmap"
	if strings.Contains(os.Args[0], "asnmap") {
		mainAsnMap()
		return
	}

	// Check if running as CGI
	if os.Getenv("GATEWAY_INTERFACE") != "" || os.Getenv("QUERY_STRING") != "" {
		handleCGI()
	} else {
		handleCLI()
	}
}

func handleCGI() {
	query := os.Getenv("QUERY_STRING")
	values, err := url.ParseQuery(query)
	if err != nil {
		respondJSON(nil, fmt.Errorf("failed to parse query string: %v", err))
		return
	}

	var result interface{}
	var queryErr error

	if domain := values.Get("domain"); domain != "" {
		result, queryErr = performQuery("domain", domain)
	} else if ip := values.Get("ip"); ip != "" {
		result, queryErr = performQuery("ip", ip)
	} else if autnum := values.Get("autnum"); autnum != "" {
		result, queryErr = performQuery("autnum", autnum)
	} else {
		respondJSON(nil, fmt.Errorf("missing parameter: expected domain, ip, or autnum"))
		return
	}

	respondJSON(result, queryErr)
}

func handleCLI() {
	domain := flag.String("domain", "", "Domain to query")
	ip := flag.String("ip", "", "IP address to query")
	autnum := flag.String("autnum", "", "ASN to query")
	flag.Parse()

	var result interface{}
	var err error

	if *domain != "" {
		result, err = performQuery("domain", *domain)
	} else if *ip != "" {
		result, err = performQuery("ip", *ip)
	} else if *autnum != "" {
		result, err = performQuery("autnum", *autnum)
	} else {
		fmt.Println("Usage: radp -domain <domain> | -ip <ip> | -autnum <asn>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err != nil {
		output, _ := json.MarshalIndent(ErrorResponse{Error: err.Error()}, "", "  ")
		fmt.Println(string(output))
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
}

func performQuery(queryType, value string) (interface{}, error) {
	// Sanitize and validate input
	value = strings.TrimSpace(value)

	switch queryType {
	case "domain":
		// Basic domain validation: alphanumeric, hyphens, dots.
		// No strict regex for all TLDs, but prevent obvious injection or garbage.
		// Regex: ^[a-zA-Z0-9.-]+$
		validDomain := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
		if !validDomain.MatchString(value) {
			return nil, fmt.Errorf("invalid domain format")
		}

		client := &rdap.Client{}
		return client.QueryDomain(value)

	case "ip":
		// Check if it's a valid IP or CIDR
		// rdap.QueryIP takes an IP address string.
		// We can use net.ParseIP to check.

		// If it has a CIDR mask, we might need to strip it or handle it.
		// OpenRDAP QueryIP expects just the IP usually, but let's check.
		// Actually openrdap QueryIP handles parsing.
		// But we want to pre-validate.

		ip, _, err := net.ParseCIDR(value)
		if err != nil {
			// Try as raw IP
			ip = net.ParseIP(value)
			if ip == nil {
				return nil, fmt.Errorf("invalid IP address format")
			}
		}
		// If valid, use the original string (openrdap handles both)

		client := &rdap.Client{}
		return client.QueryIP(value)

	case "autnum":
		// Validate ASN format: AS<digits> or <digits>
		validASN := regexp.MustCompile(`^(AS|as)?\d+$`)
		if !validASN.MatchString(value) {
			return nil, fmt.Errorf("invalid ASN format")
		}

		// OpenRDAP expects full ASN string maybe?
		// Let's ensure consistency or pass as is if library handles it.
		// But library probably expects e.g. "AS15169" or "15169".

		client := &rdap.Client{}
		return client.QueryAutnum(value)

	default:
		return nil, fmt.Errorf("unknown query type: %s", queryType)
	}
}

func respondJSON(data interface{}, err error) {
	fmt.Printf("Content-Type: application/json\r\n\r\n")

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err != nil {
		encoder.Encode(ErrorResponse{Error: err.Error()})
	} else {
		encoder.Encode(data)
	}
}
