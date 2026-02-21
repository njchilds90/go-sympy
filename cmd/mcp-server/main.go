// cmd/mcp-server/main.go — Standalone HTTP MCP server for go-sympy
//
// Exposes go-sympy tools as an HTTP endpoint for AI agent frameworks.
//
// Usage:
//   go run cmd/mcp-server/main.go -port 8080
//
// Tool call endpoint: POST /tool
// Schema endpoint:    GET  /schema
// Health endpoint:    GET  /health
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	gosympy "github.com/njchilds90/go-sympy"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	// POST /tool — handle a tool call
	mux.HandleFunc("/tool", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req gosympy.ToolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		resp := gosympy.HandleToolCall(req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /schema — return tool schema for agent registration
	mux.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, gosympy.MCPToolSpec())
	})

	// GET /health — liveness check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("go-sympy MCP server listening on %s", addr)
	log.Printf("  POST /tool   — execute a tool call")
	log.Printf("  GET  /schema — tool schema for agent registration")
	log.Printf("  GET  /health — health check")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
