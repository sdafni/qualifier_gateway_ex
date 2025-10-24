package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	// Custom header name to determine routing
	routingHeader = "X-Route-To"

	// Gateway endpoint
	gatewayPath = "/gateway_endpoint"

	// Default port
	defaultPort = "8080"
)

type Gateway struct {
	url1 string
	url2 string
}

func NewGateway(url1, url2 string) *Gateway {
	return &Gateway{
		url1: url1,
		url2: url2,
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle the gateway endpoint
	if r.URL.Path != gatewayPath {
		http.NotFound(w, r)
		return
	}

	// Get the routing header value
	routeTo := r.Header.Get(routingHeader)

	var targetURL string
	switch strings.ToLower(routeTo) {
	case "url1":
		targetURL = g.url1
	case "url2":
		targetURL = g.url2
	default:
		http.Error(w, fmt.Sprintf("Invalid or missing %s header. Must be 'url1' or 'url2'", routingHeader), http.StatusBadRequest)
		log.Printf("Invalid routing header: %s", routeTo)
		return
	}

	// Construct the target URL with /endpoint path
	target, err := url.Parse(targetURL + "/endpoint")
	if err != nil {
		http.Error(w, "Invalid target URL configuration", http.StatusInternalServerError)
		log.Printf("Error parsing target URL: %v", err)
		return
	}

	log.Printf("Routing request to: %s (route: %s)", target.String(), routeTo)

	// Create a new request to forward
	proxyReq, err := http.NewRequest(r.Method, target.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		log.Printf("Error creating proxy request: %v", err)
		return
	}

	// Copy headers from original request (excluding routing header)
	for name, values := range r.Header {
		if name != routingHeader {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}
	}

	// Forward the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		log.Printf("Error forwarding request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}

	log.Printf("Request completed with status: %d", resp.StatusCode)
}

func main() {
	// Get configuration from environment variables or use defaults
	url1 := os.Getenv("GATEWAY_URL1")
	url2 := os.Getenv("GATEWAY_URL2")
	port := os.Getenv("GATEWAY_PORT")

	if url1 == "" {
		url1 = "http://localhost:8081"
		log.Printf("GATEWAY_URL1 not set, using default: %s", url1)
	}

	if url2 == "" {
		url2 = "http://localhost:8084"
		log.Printf("GATEWAY_URL2 not set, using default: %s", url2)
	}

	if port == "" {
		port = defaultPort
	}

	gateway := NewGateway(url1, url2)

	log.Printf("Starting gateway server on port %s", port)
	log.Printf("URL1: %s", url1)
	log.Printf("URL2: %s", url2)
	log.Printf("Gateway endpoint: %s", gatewayPath)
	log.Printf("Routing header: %s", routingHeader)
	log.Println("Routes:")
	log.Printf("  %s: url1 -> %s/endpoint", routingHeader, url1)
	log.Printf("  %s: url2 -> %s/endpoint", routingHeader, url2)

	if err := http.ListenAndServe(":"+port, gateway); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
