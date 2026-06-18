package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestImportLabelsFromFile(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)
	store, err := cache.OpenSQLite(context.Background(), driverName, "labels-import")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	dir := t.TempDir()
	labelsPath := filepath.Join(dir, "labels.toml")
	content := `
[accounts]
"GABC123" = "Treasury"
`
	if err := os.WriteFile(labelsPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write labels: %v", err)
	}

	cfg := config.Default()
	if err := importLabelsFromFile(context.Background(), store, cfg, labelsPath); err != nil {
		t.Fatalf("importLabelsFromFile() error = %v", err)
	}

	labels, err := store.ListLabels(context.Background())
	if err != nil {
		t.Fatalf("ListLabels() error = %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0].Name != "Treasury" {
		t.Fatalf("label name = %q", labels[0].Name)
	}

	targets, err := store.ListLabelTargets(context.Background())
	if err != nil {
		t.Fatalf("ListLabelTargets() error = %v", err)
	}
	if len(targets) != 1 || targets[0].Target != "GABC123" {
		t.Fatalf("unexpected label targets: %+v", targets)
	}
}
