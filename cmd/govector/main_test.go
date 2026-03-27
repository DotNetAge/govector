package main

import (
	"os"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	printUsage()
}

func TestRunCommand_MissingFlags(t *testing.T) {
	// Should return nil but print missing flags
	err := runCommand([]string{"govector", "upsert"})
	if err != nil {
		t.Errorf("expected no error from runCommand itself, got %v", err)
	}
}

func TestRunCommand_UnknownCommand(t *testing.T) {
	// Unknown commands are now treated as TUI arguments.
	// We expect it to run TUI and exit immediately on EOF, returning no error.
	err := runCommand([]string{"govector", "Invalid", "-c=test"})
	if err != nil {
		t.Errorf("expected no error for TUI fallback, got %v", err)
	}
	defer os.Remove("Invalid") // remove the created file
}

func TestRunCommand_BasicFlow(t *testing.T) {
	dbFile := "test_tmp.db"
	defer os.Remove(dbFile)

	// Test Upsert with short flags
	err := runCommand([]string{"govector", "upsert", dbFile, "-c=test", "-j=[{\"id\":\"1\",\"vector\":[0.1,0.2,0.3]}]"})
	if err != nil {
		t.Errorf("Upsert failed: %v", err)
	}

	// Test ls
	err = runCommand([]string{"govector", "ls", dbFile})
	if err != nil {
		t.Errorf("ls failed: %v", err)
	}

	// Test Count with short flags
	err = runCommand([]string{"govector", "count", dbFile, "-c=test"})
	if err != nil {
		t.Errorf("Count failed: %v", err)
	}

	// Test Search with short flags
	err = runCommand([]string{"govector", "search", dbFile, "-c=test", "-v=[0.1,0.2,0.3]", "-l=5"})
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	// Test Delete with short flags
	err = runCommand([]string{"govector", "delete", dbFile, "-c=test", "-i=[\"1\"]"})
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Test rm
	err = runCommand([]string{"govector", "rm", dbFile, "-c=test"})
	if err != nil {
		t.Errorf("Rm failed: %v", err)
	}
}
