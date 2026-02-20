// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cpcloud/webcasa/internal/data"
)

// Server is the REST API server for webcasa.
type Server struct {
	handler http.Handler
	store   *data.Store
}

// NewServer creates a configured HTTP handler with all API routes and static
// file serving. webDir is the path to the web/ directory containing
// index.html; when empty, static serving is disabled.
func NewServer(store *data.Store, webDir string) *Server {
	mux := http.NewServeMux()
	a := &API{store: store}

	// House profile (singleton)
	mux.HandleFunc("GET /api/house", a.GetHouse)
	mux.HandleFunc("PUT /api/house", a.UpdateHouse)

	// Dashboard
	mux.HandleFunc("GET /api/dashboard", a.Dashboard)

	// Reference data
	mux.HandleFunc("GET /api/project-types", a.ListProjectTypes)
	mux.HandleFunc("GET /api/maintenance-categories", a.ListMaintenanceCategories)

	// Projects
	mux.HandleFunc("GET /api/projects", a.ListProjects)
	mux.HandleFunc("GET /api/projects/{id}", a.GetProject)
	mux.HandleFunc("POST /api/projects", a.CreateProject)
	mux.HandleFunc("PUT /api/projects/{id}", a.UpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", a.DeleteProject)
	mux.HandleFunc("POST /api/projects/{id}/restore", a.RestoreProject)
	mux.HandleFunc("GET /api/projects/{id}/quotes", a.ListQuotesByProject)

	// Quotes
	mux.HandleFunc("GET /api/quotes", a.ListQuotes)
	mux.HandleFunc("GET /api/quotes/{id}", a.GetQuote)
	mux.HandleFunc("POST /api/quotes", a.CreateQuote)
	mux.HandleFunc("PUT /api/quotes/{id}", a.UpdateQuote)
	mux.HandleFunc("DELETE /api/quotes/{id}", a.DeleteQuote)
	mux.HandleFunc("POST /api/quotes/{id}/restore", a.RestoreQuote)

	// Vendors
	mux.HandleFunc("GET /api/vendors", a.ListVendors)
	mux.HandleFunc("GET /api/vendors/{id}", a.GetVendor)
	mux.HandleFunc("POST /api/vendors", a.CreateVendor)
	mux.HandleFunc("PUT /api/vendors/{id}", a.UpdateVendor)
	mux.HandleFunc("DELETE /api/vendors/{id}", a.DeleteVendor)
	mux.HandleFunc("POST /api/vendors/{id}/restore", a.RestoreVendor)
	mux.HandleFunc("GET /api/vendors/{id}/quotes", a.ListQuotesByVendor)
	mux.HandleFunc("GET /api/vendors/{id}/service-logs", a.ListServiceLogsByVendor)

	// Maintenance
	mux.HandleFunc("GET /api/maintenance", a.ListMaintenance)
	mux.HandleFunc("GET /api/maintenance/{id}", a.GetMaintenance)
	mux.HandleFunc("POST /api/maintenance", a.CreateMaintenance)
	mux.HandleFunc("PUT /api/maintenance/{id}", a.UpdateMaintenance)
	mux.HandleFunc("DELETE /api/maintenance/{id}", a.DeleteMaintenance)
	mux.HandleFunc("POST /api/maintenance/{id}/restore", a.RestoreMaintenance)
	mux.HandleFunc("GET /api/maintenance/{id}/service-logs", a.ListServiceLogs)
	mux.HandleFunc("POST /api/maintenance/{id}/service-logs", a.CreateServiceLog)

	// Service logs
	mux.HandleFunc("GET /api/service-logs/{id}", a.GetServiceLog)
	mux.HandleFunc("PUT /api/service-logs/{id}", a.UpdateServiceLog)
	mux.HandleFunc("DELETE /api/service-logs/{id}", a.DeleteServiceLog)
	mux.HandleFunc("POST /api/service-logs/{id}/restore", a.RestoreServiceLog)

	// Appliances
	mux.HandleFunc("GET /api/appliances", a.ListAppliances)
	mux.HandleFunc("GET /api/appliances/{id}", a.GetAppliance)
	mux.HandleFunc("POST /api/appliances", a.CreateAppliance)
	mux.HandleFunc("PUT /api/appliances/{id}", a.UpdateAppliance)
	mux.HandleFunc("DELETE /api/appliances/{id}", a.DeleteAppliance)
	mux.HandleFunc("POST /api/appliances/{id}/restore", a.RestoreAppliance)
	mux.HandleFunc("GET /api/appliances/{id}/maintenance", a.ListMaintenanceByAppliance)

	// Incidents
	mux.HandleFunc("GET /api/incidents", a.ListIncidents)
	mux.HandleFunc("GET /api/incidents/{id}", a.GetIncident)
	mux.HandleFunc("POST /api/incidents", a.CreateIncident)
	mux.HandleFunc("PUT /api/incidents/{id}", a.UpdateIncident)
	mux.HandleFunc("DELETE /api/incidents/{id}", a.DeleteIncident)
	mux.HandleFunc("POST /api/incidents/{id}/restore", a.RestoreIncident)

	// Documents
	mux.HandleFunc("GET /api/documents", a.ListDocuments)
	mux.HandleFunc("GET /api/documents/{id}", a.GetDocument)
	mux.HandleFunc("GET /api/documents/{id}/download", a.DownloadDocument)
	mux.HandleFunc("POST /api/documents", a.UploadDocument)
	mux.HandleFunc("PUT /api/documents/{id}", a.UpdateDocument)
	mux.HandleFunc("DELETE /api/documents/{id}", a.DeleteDocument)
	mux.HandleFunc("POST /api/documents/{id}/restore", a.RestoreDocument)
	mux.HandleFunc("GET /api/documents/by/{kind}/{eid}", a.ListDocumentsByEntity)

	// Static files — serve web/ directory at root
	if webDir != "" {
		fs := http.FileServer(http.Dir(webDir))
		mux.Handle("/", fs)
	}

	handler := withMiddleware(mux)
	return &Server{handler: handler, store: store}
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// withMiddleware wraps the mux with recovery, CORS, and logging.
func withMiddleware(h http.Handler) http.Handler {
	return withRecovery(withLogging(withCORS(h)))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func withLogging(next http.Handler) http.Handler {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond))
	})
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Fprintf(os.Stderr, "panic: %v\n", err)
				jsonError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
