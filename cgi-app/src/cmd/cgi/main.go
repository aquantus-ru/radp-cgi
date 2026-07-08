package main

import (
	"cgi-rdap/pkg/core"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

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
		result, queryErr = core.PerformQuery("domain", domain)
	} else if ip := values.Get("ip"); ip != "" {
		result, queryErr = core.PerformQuery("ip", ip)
	} else if autnum := values.Get("autnum"); autnum != "" {
		result, queryErr = core.PerformQuery("autnum", autnum)
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
		result, err = core.PerformQuery("domain", *domain)
	} else if *ip != "" {
		result, err = core.PerformQuery("ip", *ip)
	} else if *autnum != "" {
		result, err = core.PerformQuery("autnum", *autnum)
	} else {
		fmt.Println("Usage: radp -domain <domain> | -ip <ip> | -autnum <asn>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err != nil {
		output, _ := json.MarshalIndent(core.ErrorResponse{Error: err.Error()}, "", "  ")
		fmt.Println(string(output))
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
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

	result, err := core.QueryASN(asn)
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

	result, err := core.QueryASN(asn)
	if err != nil {
		output, _ := json.MarshalIndent(core.ErrorResponse{Error: err.Error()}, "", "  ")
		fmt.Println(string(output))
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
}

func respondJSON(data interface{}, err error) {
	fmt.Printf("Content-Type: application/json\r\n\r\n")

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err != nil {
		encoder.Encode(core.ErrorResponse{Error: err.Error()})
	} else {
		encoder.Encode(data)
	}
}
