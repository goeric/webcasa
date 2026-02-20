// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cpcloud/webcasa/internal/api"
	"github.com/cpcloud/webcasa/internal/data"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address (host:port)")
	dbPath := flag.String("db", "", "SQLite database path (default: platform data dir)")
	demo := flag.Bool("demo", false, "seed demo data into an in-memory database")
	webDir := flag.String("web-dir", "web", "path to web/ directory for static files")
	flag.Parse()

	resolvedDB, err := resolveDB(*dbPath, *demo)
	if err != nil {
		fail("resolve db path", err)
	}

	store, err := data.Open(resolvedDB)
	if err != nil {
		fail("open database", err)
	}
	defer store.Close()

	if err := store.AutoMigrate(); err != nil {
		fail("migrate database", err)
	}
	if err := store.SeedDefaults(); err != nil {
		fail("seed defaults", err)
	}
	if *demo {
		if err := store.SeedDemoData(); err != nil {
			fail("seed demo data", err)
		}
		fmt.Fprintf(os.Stderr, "webcasa: demo data seeded\n")
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      api.NewServer(store, *webDir),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Fprintf(os.Stderr, "webcasa: listening on %s\n", *addr)
		if resolvedDB == ":memory:" {
			fmt.Fprintf(os.Stderr, "webcasa: using in-memory database (demo mode)\n")
		} else {
			fmt.Fprintf(os.Stderr, "webcasa: database at %s\n", resolvedDB)
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fail("listen", err)
		}
	}()

	<-ctx.Done()
	fmt.Fprintf(os.Stderr, "\nwebcasa: shutting down...\n")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fail("shutdown", err)
	}
}

func resolveDB(path string, demo bool) (string, error) {
	if path != "" {
		return path, nil
	}
	if demo {
		return ":memory:", nil
	}
	return data.DefaultDBPath()
}

func fail(context string, err error) {
	fmt.Fprintf(os.Stderr, "webcasa: %s: %v\n", context, err)
	os.Exit(1)
}
