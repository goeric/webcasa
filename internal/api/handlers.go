// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/cpcloud/micasa/internal/data"
	"gorm.io/gorm"
)

// API holds the store reference for all handlers.
type API struct {
	store *data.Store
}

// ── House Profile ──────────────────────────────────

func (a *API) GetHouse(w http.ResponseWriter, r *http.Request) {
	profile, err := a.store.HouseProfile()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		jsonOK(w, map[string]any{})
		return
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, profile)
}

func (a *API) UpdateHouse(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody[data.HouseProfile](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Auto-create if no profile exists yet.
	_, getErr := a.store.HouseProfile()
	if errors.Is(getErr, gorm.ErrRecordNotFound) {
		if err := a.store.CreateHouseProfile(body); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else if getErr != nil {
		jsonError(w, http.StatusInternalServerError, getErr.Error())
		return
	} else {
		if err := a.store.UpdateHouseProfile(body); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	profile, err := a.store.HouseProfile()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, profile)
}

// ── Reference Data ─────────────────────────────────

func (a *API) ListProjectTypes(w http.ResponseWriter, _ *http.Request) {
	types, err := a.store.ProjectTypes()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, types)
}

func (a *API) ListMaintenanceCategories(w http.ResponseWriter, _ *http.Request) {
	cats, err := a.store.MaintenanceCategories()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, cats)
}

// ── Projects ───────────────────────────────────────

func (a *API) ListProjects(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListProjects(boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) GetProject(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.GetProject(id)
	if err != nil {
		handleGetError(w, err, "project")
		return
	}
	jsonOK(w, item)
}

func (a *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody[data.Project](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.CreateProject(&body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonCreated(w, body)
}

func (a *API) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := decodeBody[data.Project](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.ID = id
	if err := a.store.UpdateProject(body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := a.store.GetProject(id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, updated)
}

func (a *API) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DeleteProject(id); err != nil {
		handleDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) RestoreProject(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.RestoreProject(id); err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Quotes ─────────────────────────────────────────

// quoteRequest wraps a Quote with an optional inline Vendor for
// find-or-create behavior on create/update.
type quoteRequest struct {
	data.Quote
	Vendor data.Vendor `json:"Vendor"`
}

func (a *API) ListQuotes(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListQuotes(boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) GetQuote(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.GetQuote(id)
	if err != nil {
		handleGetError(w, err, "quote")
		return
	}
	jsonOK(w, item)
}

func (a *API) CreateQuote(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody[quoteRequest](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.CreateQuote(&body.Quote, body.Vendor); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	created, _ := a.store.GetQuote(body.Quote.ID)
	jsonCreated(w, created)
}

func (a *API) UpdateQuote(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := decodeBody[quoteRequest](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.Quote.ID = id
	if err := a.store.UpdateQuote(body.Quote, body.Vendor); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, _ := a.store.GetQuote(id)
	jsonOK(w, updated)
}

func (a *API) DeleteQuote(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DeleteQuote(id); err != nil {
		handleDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) RestoreQuote(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.RestoreQuote(id); err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) ListQuotesByProject(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := a.store.ListQuotesByProject(id, boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) ListQuotesByVendor(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := a.store.ListQuotesByVendor(id, boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

// ── Vendors ────────────────────────────────────────

func (a *API) ListVendors(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListVendors(boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) GetVendor(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.GetVendor(id)
	if err != nil {
		handleGetError(w, err, "vendor")
		return
	}
	jsonOK(w, item)
}

func (a *API) CreateVendor(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody[data.Vendor](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.CreateVendor(&body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonCreated(w, body)
}

func (a *API) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := decodeBody[data.Vendor](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.ID = id
	if err := a.store.UpdateVendor(body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := a.store.GetVendor(id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, updated)
}

func (a *API) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DeleteVendor(id); err != nil {
		handleDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) RestoreVendor(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.RestoreVendor(id); err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) ListServiceLogsByVendor(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := a.store.ListServiceLogsByVendor(id, boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

// ── Maintenance ────────────────────────────────────

func (a *API) ListMaintenance(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListMaintenance(boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) GetMaintenance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.GetMaintenance(id)
	if err != nil {
		handleGetError(w, err, "maintenance item")
		return
	}
	jsonOK(w, item)
}

func (a *API) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody[data.MaintenanceItem](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.CreateMaintenance(&body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonCreated(w, body)
}

func (a *API) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := decodeBody[data.MaintenanceItem](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.ID = id
	if err := a.store.UpdateMaintenance(body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := a.store.GetMaintenance(id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, updated)
}

func (a *API) DeleteMaintenance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DeleteMaintenance(id); err != nil {
		handleDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) RestoreMaintenance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.RestoreMaintenance(id); err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) ListMaintenanceByAppliance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := a.store.ListMaintenanceByAppliance(id, boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

// ── Service Logs ───────────────────────────────────

// serviceLogRequest wraps a ServiceLogEntry with an optional inline Vendor.
type serviceLogRequest struct {
	data.ServiceLogEntry
	Vendor data.Vendor `json:"Vendor"`
}

func (a *API) ListServiceLogs(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := a.store.ListServiceLog(id, boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) GetServiceLog(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.GetServiceLog(id)
	if err != nil {
		handleGetError(w, err, "service log")
		return
	}
	jsonOK(w, item)
}

func (a *API) CreateServiceLog(w http.ResponseWriter, r *http.Request) {
	maintID, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := decodeBody[serviceLogRequest](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.ServiceLogEntry.MaintenanceItemID = maintID
	if err := a.store.CreateServiceLog(&body.ServiceLogEntry, body.Vendor); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	created, _ := a.store.GetServiceLog(body.ServiceLogEntry.ID)
	jsonCreated(w, created)
}

func (a *API) UpdateServiceLog(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := decodeBody[serviceLogRequest](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.ServiceLogEntry.ID = id
	if err := a.store.UpdateServiceLog(body.ServiceLogEntry, body.Vendor); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, _ := a.store.GetServiceLog(id)
	jsonOK(w, updated)
}

func (a *API) DeleteServiceLog(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DeleteServiceLog(id); err != nil {
		handleDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) RestoreServiceLog(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.RestoreServiceLog(id); err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Appliances ─────────────────────────────────────

func (a *API) ListAppliances(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListAppliances(boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) GetAppliance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.GetAppliance(id)
	if err != nil {
		handleGetError(w, err, "appliance")
		return
	}
	jsonOK(w, item)
}

func (a *API) CreateAppliance(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody[data.Appliance](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.CreateAppliance(&body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonCreated(w, body)
}

func (a *API) UpdateAppliance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := decodeBody[data.Appliance](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.ID = id
	if err := a.store.UpdateAppliance(body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := a.store.GetAppliance(id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, updated)
}

func (a *API) DeleteAppliance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DeleteAppliance(id); err != nil {
		handleDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) RestoreAppliance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.RestoreAppliance(id); err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Incidents ──────────────────────────────────────

func (a *API) ListIncidents(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListIncidents(boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) GetIncident(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.GetIncident(id)
	if err != nil {
		handleGetError(w, err, "incident")
		return
	}
	jsonOK(w, item)
}

func (a *API) CreateIncident(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody[data.Incident](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.CreateIncident(&body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonCreated(w, body)
}

func (a *API) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := decodeBody[data.Incident](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.ID = id
	if err := a.store.UpdateIncident(body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := a.store.GetIncident(id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, updated)
}

func (a *API) DeleteIncident(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DeleteIncident(id); err != nil {
		handleDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) RestoreIncident(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.RestoreIncident(id); err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Error helpers ──────────────────────────────────

func handleGetError(w http.ResponseWriter, err error, entity string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		jsonError(w, http.StatusNotFound, entity+" not found")
		return
	}
	jsonError(w, http.StatusInternalServerError, err.Error())
}

func handleDeleteError(w http.ResponseWriter, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	// Store returns user-facing messages for FK constraint violations
	// (e.g. "vendor has 3 active quote(s) -- delete them first").
	if strings.Contains(err.Error(), "delete") || strings.Contains(err.Error(), "active") {
		jsonError(w, http.StatusConflict, err.Error())
		return
	}
	jsonError(w, http.StatusInternalServerError, err.Error())
}

