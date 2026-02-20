// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const maxBodySize = 1 << 20 // 1 MiB

func jsonOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

func jsonCreated(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, data)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Headers already sent; log but can't change status.
		fmt.Fprintf(w, `{"error":"encode: %s"}`, err)
	}
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

func parseID(r *http.Request) (uint, error) {
	raw := r.PathValue("id")
	if raw == "" {
		return 0, fmt.Errorf("missing id parameter")
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id %q: %w", raw, err)
	}
	if n == 0 {
		return 0, fmt.Errorf("id must be positive")
	}
	return uint(n), nil
}

func boolQuery(r *http.Request, key string) bool {
	return r.URL.Query().Get(key) == "true"
}

func decodeBody[T any](r *http.Request) (T, error) {
	var v T
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode request body: %w", err)
	}
	return v, nil
}
