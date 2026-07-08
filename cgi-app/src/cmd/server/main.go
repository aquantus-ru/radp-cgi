package main

import (
	"cgi-rdap/pkg/core"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/cgi-bin/radp", handleRadp)
	http.HandleFunc("/cgi-bin/asnmap", handleAsnMap)

	fs := http.FileServer(http.Dir("www"))
	http.Handle("/", fs)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}

func handleRadp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	values := r.URL.Query()
	var result interface{}
	var queryErr error

	if domain := values.Get("domain"); domain != "" {
		result, queryErr = core.PerformQuery("domain", domain)
	} else if ip := values.Get("ip"); ip != "" {
		result, queryErr = core.PerformQuery("ip", ip)
	} else if autnum := values.Get("autnum"); autnum != "" {
		result, queryErr = core.PerformQuery("autnum", autnum)
	} else {
		respondJSON(w, nil, fmt.Errorf("missing parameter: expected domain, ip, or autnum"))
		return
	}

	respondJSON(w, result, queryErr)
}

func handleAsnMap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	values := r.URL.Query()
	asn := values.Get("asn")
	if asn == "" {
		respondJSON(w, nil, fmt.Errorf("missing parameter: expected asn"))
		return
	}

	result, err := core.QueryASN(asn)
	respondJSON(w, result, err)
}

func respondJSON(w http.ResponseWriter, data interface{}, err error) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // or maybe 400 depending on error, keeping simple
		encoder.Encode(core.ErrorResponse{Error: err.Error()})
	} else {
		encoder.Encode(data)
	}
}
