// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package api

import (
	"net/http"
	"time"

	"github.com/cpcloud/webcasa/internal/data"
)

// dashboardResponse is the aggregated JSON returned by GET /api/dashboard.
type dashboardResponse struct {
	Incidents          []data.Incident        `json:"incidents"`
	Maintenance        []data.MaintenanceItem `json:"maintenance"`
	ActiveProjects     []data.Project         `json:"activeProjects"`
	ExpiringWarranties []data.Appliance       `json:"expiringWarranties"`
	House              *data.HouseProfile     `json:"house,omitempty"`
	RecentServiceLogs  []data.ServiceLogEntry `json:"recentServiceLogs"`
	YTDServiceSpend    int64                  `json:"ytdServiceSpendCents"`
	TotalProjectSpend  int64                  `json:"totalProjectSpendCents"`
}

func (a *API) Dashboard(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()

	incidents, err := a.store.ListOpenIncidents()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	maintenance, err := a.store.ListMaintenanceWithSchedule()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	projects, err := a.store.ListActiveProjects()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	warranties, err := a.store.ListExpiringWarranties(now, 30*24*time.Hour, 90*24*time.Hour)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var house *data.HouseProfile
	h, err := a.store.HouseProfile()
	if err == nil {
		house = &h
	}

	recentLogs, err := a.store.ListRecentServiceLogs(5)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	ytdSpend, err := a.store.YTDServiceSpendCents(yearStart)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	projectSpend, err := a.store.TotalProjectSpendCents()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Ensure non-nil slices for clean JSON output.
	if incidents == nil {
		incidents = []data.Incident{}
	}
	if maintenance == nil {
		maintenance = []data.MaintenanceItem{}
	}
	if projects == nil {
		projects = []data.Project{}
	}
	if warranties == nil {
		warranties = []data.Appliance{}
	}
	if recentLogs == nil {
		recentLogs = []data.ServiceLogEntry{}
	}

	jsonOK(w, dashboardResponse{
		Incidents:          incidents,
		Maintenance:        maintenance,
		ActiveProjects:     projects,
		ExpiringWarranties: warranties,
		House:              house,
		RecentServiceLogs:  recentLogs,
		YTDServiceSpend:    ytdSpend,
		TotalProjectSpend:  projectSpend,
	})
}
