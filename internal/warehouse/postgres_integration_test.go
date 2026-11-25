//go:build integration

package warehouse_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"ttwms/internal/config"
	"ttwms/internal/warehouse"
	"ttwms/migrations"
)

// This test uses an isolated schema, never the application's own tables.
func TestPostgresSetupLifecycle(t *testing.T) {
	var cfg config.Config
	var err error
	if path := os.Getenv("TEST_ENV_FILE"); path != "" {
		cfg, err = config.LoadFile(path)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Environment != config.Development && cfg.Environment != config.Test {
		t.Fatal("integration tests require APP_ENV=development or test")
	}
	databaseURL := cfg.DatabaseURL
	if databaseURL == "" {
		t.Fatal("integration tests require DATABASE_URL; run bash scripts/check.sh to use the development .env file")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	admin, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal("cannot configure integration database")
	}
	defer admin.Close()
	if err := admin.PingContext(ctx); err != nil {
		t.Fatal("cannot connect to integration database")
	}
	// Install this shared extension in public, outside the disposable schema.
	if _, err := admin.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public`); err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("ttwms_test_%d", time.Now().UnixNano())
	if _, err := admin.ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if _, err := admin.ExecContext(cleanup, `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Errorf("clean up integration schema: %v", err)
		}
	}()
	u, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal("invalid integration database URL")
	}
	parameters := u.Query()
	parameters.Set("search_path", schema)
	u.RawQuery = parameters.Encode()
	open := func() *sql.DB {
		db, err := sql.Open("pgx", u.String())
		if err != nil {
			t.Fatal("cannot configure isolated schema")
		}
		db.SetMaxOpenConns(2)
		return db
	}
	db := open()
	defer func() { db.Close() }()
	if err := migrations.Check(ctx, db); err == nil {
		t.Fatal("empty schema was reported current")
	}

	// Both initial startup processes must safely agree on the migration state.
	results := make(chan error, 2)
	for range 2 {
		go func() { results <- migrations.Apply(ctx, db) }()
	}
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	t.Run("schema check is read only", func(t *testing.T) {
		readOnlyURL := *u
		parameters := readOnlyURL.Query()
		parameters.Set("default_transaction_read_only", "on")
		readOnlyURL.RawQuery = parameters.Encode()
		readOnly, err := sql.Open("pgx", readOnlyURL.String())
		if err != nil {
			t.Fatal("cannot configure read-only check")
		}
		defer readOnly.Close()
		if err := migrations.Check(ctx, readOnly); err != nil {
			t.Fatalf("schema verification required writes: %v", err)
		}
	})
	t.Run("newer database is rejected", func(t *testing.T) {
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(name,checksum) VALUES ('999_future.sql','future')`); err != nil {
			t.Fatal(err)
		}
		defer db.ExecContext(ctx, `DELETE FROM schema_migrations WHERE name='999_future.sql'`)
		if err := migrations.Check(ctx, db); err == nil {
			t.Fatal("database from a newer build was accepted")
		}
	})
	store := &warehouse.Postgres{DB: db}
	var location warehouse.Asset
	for _, test := range []struct {
		kind  warehouse.Kind
		asset warehouse.Asset
	}{
		{warehouse.Locations, warehouse.Asset{Code: "LOC-001", Barcode: "L001", LocationType: "pick", Active: true}},
		{warehouse.Trolleys, warehouse.Asset{Code: "TROLLEY-001", Barcode: "T001", Status: "active", CapacityOrders: 50}},
		{warehouse.PackingStations, warehouse.Asset{Code: "PACK-01", Status: "active", PrinterName: "Test printer", PrinterFormat: "100x150"}},
	} {
		t.Run(string(test.kind), func(t *testing.T) {
			created, err := store.Save(ctx, test.kind, test.asset)
			if err != nil {
				t.Fatal(err)
			}
			if created.ID == "" {
				t.Fatal("asset has no database identity")
			}
			if _, err := store.Save(ctx, test.kind, test.asset); !errors.Is(err, warehouse.ErrConflict) {
				t.Fatalf("duplicate code accepted: %v", err)
			}
			if test.kind != warehouse.PackingStations {
				duplicateBarcode := test.asset
				duplicateBarcode.Code += "-OTHER"
				if _, err := store.Save(ctx, test.kind, duplicateBarcode); !errors.Is(err, warehouse.ErrConflict) {
					t.Fatalf("duplicate barcode accepted: %v", err)
				}
			}
			created.Code += "-EDITED"
			updated, err := store.Save(ctx, test.kind, created)
			if err != nil {
				t.Fatal(err)
			}
			read, err := store.Get(ctx, test.kind, updated.ID)
			if err != nil || read != updated {
				t.Fatalf("saved values did not round-trip: %v", err)
			}
			items, err := store.List(ctx, test.kind)
			if err != nil || len(items) != 1 || items[0] != read {
				t.Fatalf("unexpected list: %v %#v", err, items)
			}
			var audits int
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM audit_log WHERE entity_id=$1`, created.ID).Scan(&audits); err != nil || audits != 2 {
				t.Fatalf("expected exactly two committed audit events: %d %v", audits, err)
			}
			if test.kind == warehouse.Locations {
				location = read
			}
		})
	}
	if t.Failed() {
		return
	}

	t.Run("warehouse boundary", func(t *testing.T) {
		var otherID string
		err := db.QueryRowContext(ctx, `WITH company AS (
			INSERT INTO companies(name) VALUES ('Other company') RETURNING id
		), warehouse AS (
			INSERT INTO warehouses(company_id,code,name) SELECT id,'OTHER','Other warehouse' FROM company RETURNING id
		) INSERT INTO locations(warehouse_id,code,barcode,type) SELECT id,'OTHER-LOC','OTHER-BAR','pick' FROM warehouse RETURNING id::text`).Scan(&otherID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.Get(ctx, warehouse.Locations, otherID); !errors.Is(err, warehouse.ErrNotFound) {
			t.Fatalf("other warehouse was readable: %v", err)
		}
		other := location
		other.ID = otherID
		if _, err := store.Save(ctx, warehouse.Locations, other); !errors.Is(err, warehouse.ErrNotFound) {
			t.Fatalf("other warehouse was editable: %v", err)
		}
	})

	t.Run("failed audit rolls back asset", func(t *testing.T) {
		if _, err := db.ExecContext(ctx, `ALTER TABLE audit_log ADD CONSTRAINT test_reject_audit CHECK(action <> 'setup.create') NOT VALID`); err != nil {
			t.Fatal(err)
		}
		defer db.ExecContext(ctx, `ALTER TABLE audit_log DROP CONSTRAINT test_reject_audit`)
		asset := warehouse.Asset{Code: "MUST-ROLL-BACK", Barcode: "ROLL-BACK", LocationType: "pick", Active: true}
		if _, err := store.Save(ctx, warehouse.Locations, asset); err == nil {
			t.Fatal("audit failure did not abort save")
		}
		var count int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM locations WHERE code='MUST-ROLL-BACK'`).Scan(&count); err != nil || count != 0 {
			t.Fatalf("failed save left an unaudited asset: %d %v", count, err)
		}
	})

	// Reopen the connection and reapply migrations, as a restarted server would.
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = open()
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	store = &warehouse.Postgres{DB: db}
	read, err := store.Get(ctx, warehouse.Locations, location.ID)
	if err != nil || read != location {
		t.Fatalf("asset was not persistent: %v", err)
	}
	var applied int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&applied); err != nil || applied != 2 {
		t.Fatalf("migrations were not applied exactly once: %d %v", applied, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE schema_migrations SET checksum='changed' WHERE name='001_foundation.sql'`); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Apply(ctx, db); err == nil || !strings.Contains(err.Error(), "has changed") {
		t.Fatalf("changed migration was not rejected: %v", err)
	}
	if err := migrations.Check(ctx, db); err == nil || !strings.Contains(err.Error(), "has changed") {
		t.Fatalf("schema check accepted a changed migration: %v", err)
	}
}
