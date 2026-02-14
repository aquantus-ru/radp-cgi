package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// ASNMapResponse represents the response for ASN map query
type ASNMapResponse struct {
	ASN      string   `json:"asn"`
	Prefixes []string `json:"prefixes"`
}

func mainAsnMap() {
	// Check if running as CGI
	if os.Getenv("GATEWAY_INTERFACE") != "" || os.Getenv("QUERY_STRING") != "" {
		handleAsnMapCGI()
	} else {
		handleAsnMapCLI()
	}
}

func handleAsnMapCGI() {
	query := os.Getenv("QUERY_STRING")
	values, err := url.ParseQuery(query)
	if err != nil {
		respondJSON(nil, fmt.Errorf("failed to parse query string: %v", err))
		return
	}

	asn := values.Get("asn")
	if asn == "" {
		respondJSON(nil, fmt.Errorf("missing parameter: expected asn"))
		return
	}

	result, err := queryASN(asn)
	respondJSON(result, err)
}

func handleAsnMapCLI() {
	var asn string

	// Simple argument parsing
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-asn" && i+1 < len(os.Args) {
			asn = os.Args[i+1]
			i++
		} else if !strings.HasPrefix(arg, "-") {
			asn = arg
		}
	}

	if asn == "" {
		fmt.Println("Usage: asnmap <asn> | -asn <asn>")
		os.Exit(1)
	}

	result, err := queryASN(asn)
	if err != nil {
		output, _ := json.MarshalIndent(ErrorResponse{Error: err.Error()}, "", "  ")
		fmt.Println(string(output))
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
}

func queryASN(asn string) (*ASNMapResponse, error) {
	// Normalize ASN
	asn = strings.ToUpper(strings.TrimSpace(asn))

	// Validate format: must be AS<digits> or just <digits>
	// Using regex to check
	validASN := regexp.MustCompile(`^(AS)?\d+$`)
	if !validASN.MatchString(asn) {
		return nil, fmt.Errorf("invalid ASN format: expected AS<number> or <number>")
	}

	if !strings.HasPrefix(asn, "AS") {
		asn = "AS" + asn
	}
	// Extract the number part
	asnNum := strings.TrimPrefix(asn, "AS")

	prefixes, err := queryRadb(asnNum)
	if err != nil {
		return nil, err
	}

	return &ASNMapResponse{
		ASN:      asn,
		Prefixes: prefixes,
	}, nil
}

func queryRadb(asnNum string) ([]string, error) {
	var allPrefixes []string

	// Fetch IPv4
	conn, err := net.DialTimeout("tcp", "whois.radb.net:43", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to whois.radb.net: %v", err)
	}

	prefixesV4, err := fetchRadb(conn, "!gAS"+asnNum)
	conn.Close()

	if err == nil {
		allPrefixes = append(allPrefixes, prefixesV4...)
	}

	// Fetch IPv6
	conn2, err := net.DialTimeout("tcp", "whois.radb.net:43", 10*time.Second)
	if err != nil {
		if len(allPrefixes) > 0 {
			return allPrefixes, nil
		}
		return nil, fmt.Errorf("failed to connect to whois.radb.net for IPv6: %v", err)
	}
	defer conn2.Close()

	prefixesV6, err := fetchRadb(conn2, "!6AS"+asnNum)
	if err == nil {
		allPrefixes = append(allPrefixes, prefixesV6...)
	}

	if len(allPrefixes) == 0 && err != nil {
		// Only return error if we got nothing and last one was error
		// Or maybe validly empty?
		return nil, err
	}

	return allPrefixes, nil
}

func fetchRadb(conn net.Conn, query string) ([]string, error) {
	fmt.Fprintf(conn, "%s\n", query)

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)

	if line == "D" {
		return []string{}, nil
	}

	if strings.HasPrefix(line, "A") {
		dataLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		// Read 'C'
		_, _ = reader.ReadString('\n')

		prefixes := strings.Fields(dataLine)
		return prefixes, nil
	}

    if strings.HasPrefix(line, "F") {
        return nil, fmt.Errorf("radb error: %s", line)
    }

	return nil, fmt.Errorf("unexpected response from radb: %s", line)
}
