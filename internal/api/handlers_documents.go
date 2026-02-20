// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package api

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/cpcloud/micasa/internal/data"
)

// ── Documents ──────────────────────────────────────

func (a *API) ListDocuments(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListDocuments(boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) ListDocumentsByEntity(w http.ResponseWriter, r *http.Request) {
	entityKind := r.PathValue("kind")
	idStr := r.PathValue("eid")
	eid, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		jsonError(w, http.StatusBadRequest, fmt.Sprintf("invalid entity id %q", idStr))
		return
	}
	items, err := a.store.ListDocumentsByEntity(entityKind, uint(eid), boolQuery(r, "include_deleted"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, items)
}

func (a *API) GetDocument(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	doc, err := a.store.GetDocument(id)
	if err != nil {
		handleGetError(w, err, "document")
		return
	}
	jsonOK(w, doc)
}

// DownloadDocument streams the document BLOB with appropriate content headers.
func (a *API) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	doc, err := a.store.GetDocument(id)
	if err != nil {
		handleGetError(w, err, "document")
		return
	}
	if len(doc.Data) == 0 {
		jsonError(w, http.StatusNotFound, "document has no content")
		return
	}
	w.Header().Set("Content-Type", doc.MIMEType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(doc.SizeBytes, 10))
	w.WriteHeader(http.StatusOK)
	w.Write(doc.Data) //nolint:errcheck
}

// UploadDocument handles multipart form uploads. Fields:
//
//	file       - the file itself (required)
//	title      - optional title (auto-derived from filename if empty)
//	entityKind - entity type to link to (optional)
//	entityId   - entity ID to link to (optional)
//	notes      - optional notes
func (a *API) UploadDocument(w http.ResponseWriter, r *http.Request) {
	const maxUpload = 50 << 20 // 50 MiB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload+1024)

	if err := r.ParseMultipartForm(maxUpload); err != nil {
		jsonError(w, http.StatusBadRequest, fmt.Sprintf("parse form: %v -- max upload size is 50 MiB", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, http.StatusBadRequest, "missing 'file' field in multipart form")
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("read uploaded file: %v", err))
		return
	}

	title := r.FormValue("title")
	if title == "" {
		title = data.TitleFromFilename(header.Filename)
	}

	mime := header.Header.Get("Content-Type")
	if mime == "" || mime == "application/octet-stream" {
		mime = detectMIME(fileData, header.Filename)
	}

	checksum := fmt.Sprintf("%x", sha256.Sum256(fileData))

	doc := data.Document{
		Title:          title,
		FileName:       filepath.Base(header.Filename),
		EntityKind:     r.FormValue("entityKind"),
		MIMEType:       mime,
		SizeBytes:      int64(len(fileData)),
		ChecksumSHA256: checksum,
		Data:           fileData,
		Notes:          r.FormValue("notes"),
	}

	if eidStr := r.FormValue("entityId"); eidStr != "" {
		eid, err := strconv.ParseUint(eidStr, 10, 64)
		if err != nil {
			jsonError(w, http.StatusBadRequest, fmt.Sprintf("invalid entityId %q", eidStr))
			return
		}
		doc.EntityID = uint(eid)
	}

	if err := a.store.CreateDocument(&doc); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return without the BLOB data.
	doc.Data = nil
	jsonCreated(w, doc)
}

func (a *API) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Metadata-only update via JSON.
	body, err := decodeBody[data.Document](r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.ID = id
	if err := a.store.UpdateDocument(body); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, err := a.store.GetDocument(id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated.Data = nil
	jsonOK(w, updated)
}

func (a *API) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.DeleteDocument(id); err != nil {
		handleDeleteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) RestoreDocument(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.RestoreDocument(id); err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// detectMIME uses http.DetectContentType with extension fallback for types
// that content sniffing misses.
func detectMIME(data []byte, filename string) string {
	mime := http.DetectContentType(data)
	if mime != "application/octet-stream" {
		return mime
	}
	switch filepath.Ext(filename) {
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown"
	default:
		return mime
	}
}
