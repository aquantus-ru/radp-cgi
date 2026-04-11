package core

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/openrdap/rdap"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func PerformQuery(queryType, value string) (interface{}, error) {
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
