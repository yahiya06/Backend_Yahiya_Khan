package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

// Server holds all in-memory data (we'll fill this in step 2).
type Server struct{}

func NewServer() *Server {
	return &Server{}
}

// Routes registers every URL. Go 1.22 lets us put the method in the pattern.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

// writeJSON sends any value as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError sends {"error": "..."} (rule R4).
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// randomHex returns 16 random bytes as 32 hex characters (rule R7).
func randomHex() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil { // crypto/rand, NOT math/rand
		panic(err)
	}
	return hex.EncodeToString(b)
}

// nowUTC returns the current time as RFC 3339 in UTC (rule R10).
func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}