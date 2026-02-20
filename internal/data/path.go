// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

const AppName = "webcasa"

func DefaultDBPath() (string, error) {
	if override := os.Getenv("WEBCASA_DB_PATH"); override != "" {
		return override, nil
	}
	// xdg.DataFile creates the parent directory and returns the full path.
	// On Linux/WSL: $XDG_DATA_HOME/webcasa/webcasa.db (default ~/.local/share)
	// On macOS:     ~/Library/Application Support/webcasa/webcasa.db
	// On Windows:   %LOCALAPPDATA%/webcasa/webcasa.db
	return xdg.DataFile(filepath.Join(AppName, AppName+".db"))
}

// DocumentCacheDir returns the directory used for extracted document BLOBs.
// On Linux: $XDG_CACHE_HOME/webcasa/documents (default ~/.cache/webcasa/documents)
func DocumentCacheDir() (string, error) {
	dir := filepath.Join(xdg.CacheHome, AppName, "documents")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}
