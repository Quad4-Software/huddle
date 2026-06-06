// Command huddle runs the Huddle signaling server and embedded web UI.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"huddle/internal/config"
	"huddle/internal/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthCheck())
	}

	cfg := config.Load()
	srv := server.New(cfg)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func runHealthCheck() int {
	addr := os.Getenv("HUDDLE_HEALTH_ADDR")
	if addr == "" {
		addr = "http://127.0.0.1:8080"
	}

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(addr + "/api/health")
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %v\n", err)
		return 1
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck: status %d\n", resp.StatusCode)
		return 1
	}
	return 0
}
