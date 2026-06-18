package cache

import (
	"context"
	"database/sql"
	"testing"
)

func TestMigrationsApplyIncrementally(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected migrations to be present")
	}

	for index, migration := range migrations {
		db, err := sql.Open(driverName, "migration-step-"+migration.name)
		if err != nil {
			t.Fatalf("open db for migration %d: %v", migration.version, err)
		}

		if err := ensureMigrationTable(context.Background(), db); err != nil {
			_ = db.Close()
			t.Fatalf("ensureMigrationTable() error = %v", err)
		}

		current, err := currentSchemaVersion(context.Background(), db)
		if err != nil {
			_ = db.Close()
			t.Fatalf("currentSchemaVersion() error = %v", err)
		}
		if current != 0 {
			_ = db.Close()
			t.Fatalf("expected fresh database version 0, got %d", current)
		}

		for step := 0; step <= index; step++ {
			if err := applyMigration(context.Background(), db, migrations[step]); err != nil {
				_ = db.Close()
				t.Fatalf("apply migration %d (%s): %v", migrations[step].version, migrations[step].name, err)
			}
		}

		version, err := currentSchemaVersion(context.Background(), db)
		if err != nil {
			_ = db.Close()
			t.Fatalf("schema version after migration %d: %v", migration.version, err)
		}
		if version != migration.version {
			_ = db.Close()
			t.Fatalf("expected version %d after step %d, got %d", migration.version, index, version)
		}
		_ = db.Close()
	}
}
