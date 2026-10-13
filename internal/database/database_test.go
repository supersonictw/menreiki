package database

import (
	"path/filepath"
	"testing"
)

func TestDatabaseInitAndCommandRevision(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	appDB, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	botID := "1234567890"

	// Revision initially 0
	rev, err := appDB.GetCommandRevision(botID)
	if err != nil {
		t.Fatalf("GetCommandRevision failed: %v", err)
	}
	if rev != 0 {
		t.Errorf("expected revision 0, got %d", rev)
	}

	// Set revision to 1
	if err := appDB.SetCommandRevision(botID, 1); err != nil {
		t.Fatalf("SetCommandRevision failed: %v", err)
	}

	rev, err = appDB.GetCommandRevision(botID)
	if err != nil {
		t.Fatalf("GetCommandRevision failed: %v", err)
	}
	if rev != 1 {
		t.Errorf("expected revision 1, got %d", rev)
	}
}
