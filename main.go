package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
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
	client := &rdap.Client{}

	switch queryType {
	case "domain":
		return client.QueryDomain(value)
	case "ip":
		return client.QueryIP(value)
	case "autnum":
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
