package data

import (
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

func TestSeedDefaults(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	if err != nil {
		t.Fatalf("ProjectTypes error: %v", err)
	}
	if len(types) == 0 {
		t.Fatalf("ProjectTypes empty")
	}
	categories, err := store.MaintenanceCategories()
	if err != nil {
		t.Fatalf("MaintenanceCategories error: %v", err)
	}
	if len(categories) == 0 {
		t.Fatalf("MaintenanceCategories empty")
	}
}

func TestHouseProfileSingle(t *testing.T) {
	store := newTestStore(t)
	profile := HouseProfile{Nickname: "Primary Residence"}
	if err := store.CreateHouseProfile(profile); err != nil {
		t.Fatalf("CreateHouseProfile error: %v", err)
	}
	if _, err := store.HouseProfile(); err != nil {
		t.Fatalf("HouseProfile error: %v", err)
	}
	if err := store.CreateHouseProfile(profile); err == nil {
		t.Fatalf("expected error on second profile")
	}
}

func TestUpdateHouseProfile(t *testing.T) {
	store := newTestStore(t)
	profile := HouseProfile{Nickname: "Primary Residence", City: "Portland"}
	if err := store.CreateHouseProfile(profile); err != nil {
		t.Fatalf("CreateHouseProfile error: %v", err)
	}
	updated := HouseProfile{Nickname: "Primary Residence", City: "Seattle"}
	if err := store.UpdateHouseProfile(updated); err != nil {
		t.Fatalf("UpdateHouseProfile error: %v", err)
	}
	fetched, err := store.HouseProfile()
	if err != nil {
		t.Fatalf("HouseProfile error: %v", err)
	}
	if fetched.City != "Seattle" {
		t.Fatalf("expected city Seattle, got %q", fetched.City)
	}
}

func TestSoftDeleteRestoreProject(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	if err != nil {
		t.Fatalf("ProjectTypes error: %v", err)
	}
	project := Project{
		Title:         "Test Project",
		ProjectTypeID: types[0].ID,
		Status:        ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	projects, err := store.ListProjects(false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects expected 1, got %d err %v", len(projects), err)
	}
	if err := store.DeleteProject(projects[0].ID); err != nil {
		t.Fatalf("DeleteProject error: %v", err)
	}
	projects, err = store.ListProjects(false)
	if err != nil || len(projects) != 0 {
		t.Fatalf("ListProjects expected 0, got %d err %v", len(projects), err)
	}
	projects, err = store.ListProjects(true)
	if err != nil || len(projects) != 1 || !projects[0].DeletedAt.Valid {
		t.Fatalf("ListProjects expected deleted row")
	}
	if err := store.RestoreProject(projects[0].ID); err != nil {
		t.Fatalf("RestoreProject error: %v", err)
	}
	projects, err = store.ListProjects(false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects after restore expected 1, got %d err %v", len(projects), err)
	}
}

func TestLastDeletionRecord(t *testing.T) {
	store := newTestStore(t)
	types, err := store.ProjectTypes()
	if err != nil {
		t.Fatalf("ProjectTypes error: %v", err)
	}
	project := Project{
		Title:         "Test Project",
		ProjectTypeID: types[0].ID,
		Status:        ProjectStatusPlanned,
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	projects, err := store.ListProjects(false)
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects expected 1, got %d err %v", len(projects), err)
	}
	if err := store.DeleteProject(projects[0].ID); err != nil {
		t.Fatalf("DeleteProject error: %v", err)
	}
	record, err := store.LastDeletion(DeletionEntityProject)
	if err != nil {
		t.Fatalf("LastDeletion error: %v", err)
	}
	if record.TargetID != projects[0].ID {
		t.Fatalf("LastDeletion target %d != %d", record.TargetID, projects[0].ID)
	}
	if err := store.RestoreProject(record.TargetID); err != nil {
		t.Fatalf("RestoreProject error: %v", err)
	}
	_, err = store.LastDeletion(DeletionEntityProject)
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate error: %v", err)
	}
	if err := store.SeedDefaults(); err != nil {
		t.Fatalf("SeedDefaults error: %v", err)
	}
	return store
}
