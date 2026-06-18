package cache

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOpenAppliesMigrationsAndPersistsState(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)

	db, err := sql.Open(driverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	version, err := store.SchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("schema version: %v", err)
	}

	if version != schemaVersion {
		t.Fatalf("expected schema version %d, got %d", schemaVersion, version)
	}

	state := map[string]any{
		"network": "testnet",
		"follow":  true,
	}
	if err := store.SetStateJSON(context.Background(), "session", state); err != nil {
		t.Fatalf("set state: %v", err)
	}

	var loaded map[string]any
	if err := store.GetStateJSON(context.Background(), "session", &loaded); err != nil {
		t.Fatalf("get state: %v", err)
	}

	if loaded["network"] != "testnet" {
		t.Fatalf("expected state network to round-trip, got %v", loaded["network"])
	}
}

func TestUpsertProfileAndListProfiles(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)

	db, err := sql.Open(driverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	err = store.UpsertProfile(context.Background(), Profile{
		ID:      "public",
		Name:    "Public",
		Network: "mainnet",
		RPCURL:  "https://rpc.example",
	})
	if err != nil {
		t.Fatalf("upsert first profile: %v", err)
	}

	err = store.UpsertProfile(context.Background(), Profile{
		ID:         "testnet",
		Name:       "A Testnet",
		Network:    "testnet",
		RPCURL:     "https://rpc.testnet",
		IndexerURL: "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("upsert second profile: %v", err)
	}

	profiles, err := store.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("list profiles: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	if profiles[0].ID != "testnet" {
		t.Fatalf("expected profiles sorted by name, got first profile %q", profiles[0].ID)
	}
}

func TestUpsertBookmarkAndListBookmarks(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)

	db, err := sql.Open(driverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	err = store.UpsertBookmark(context.Background(), Bookmark{
		ID:        "bookmark-2",
		ProfileID: "public",
		Kind:      "transaction",
		Target:    "tx:abc",
		Title:     "Zebra Tx",
		Notes:     "suspicious",
	})
	if err != nil {
		t.Fatalf("upsert first bookmark: %v", err)
	}

	err = store.UpsertBookmark(context.Background(), Bookmark{
		ID:        "bookmark-1",
		ProfileID: "public",
		Kind:      "contract",
		Target:    "contract:def",
		Title:     "Alpha Contract",
	})
	if err != nil {
		t.Fatalf("upsert second bookmark: %v", err)
	}

	bookmarks, err := store.ListBookmarks(context.Background())
	if err != nil {
		t.Fatalf("list bookmarks: %v", err)
	}

	if len(bookmarks) != 2 {
		t.Fatalf("expected 2 bookmarks, got %d", len(bookmarks))
	}

	if bookmarks[0].ID != "bookmark-1" {
		t.Fatalf("expected bookmarks sorted by title, got first bookmark %q", bookmarks[0].ID)
	}
}

func TestUpsertLabelAndListLabels(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)

	db, err := sql.Open(driverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	err = store.UpsertLabel(context.Background(), Label{
		ID:        "label-2",
		ProfileID: "public",
		Name:      "wallet",
		Color:     "#00aa88",
	})
	if err != nil {
		t.Fatalf("upsert first label: %v", err)
	}

	err = store.UpsertLabel(context.Background(), Label{
		ID:        "label-1",
		ProfileID: "public",
		Name:      "amm",
		Color:     "#ff6600",
	})
	if err != nil {
		t.Fatalf("upsert second label: %v", err)
	}

	labels, err := store.ListLabels(context.Background())
	if err != nil {
		t.Fatalf("list labels: %v", err)
	}

	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}

	if labels[0].ID != "label-1" {
		t.Fatalf("expected labels sorted by name, got first label %q", labels[0].ID)
	}
}

func TestUpsertLabelTargetAndListLabelTargets(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)

	db, err := sql.Open(driverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	err = store.UpsertLabelTarget(context.Background(), LabelTarget{
		ID:        "label-target-1",
		LabelID:   "label-1",
		ProfileID: "public",
		Kind:      "account",
		Target:    "GABC",
	})
	if err != nil {
		t.Fatalf("upsert label target: %v", err)
	}

	targets, err := store.ListLabelTargets(context.Background())
	if err != nil {
		t.Fatalf("list label targets: %v", err)
	}

	if len(targets) != 1 {
		t.Fatalf("expected 1 label target, got %d", len(targets))
	}

	if targets[0].LabelID != "label-1" || targets[0].Target != "GABC" {
		t.Fatalf("unexpected label target: %#v", targets[0])
	}
}

func TestUpsertNoteAndListNotes(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)

	db, err := sql.Open(driverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	err = store.UpsertNote(context.Background(), Note{
		ID:        "note-older",
		ProfileID: "public",
		Target:    "account:GABC",
		Title:     "Older",
		Body:      "first note",
		UpdatedAt: time.Unix(10, 0).UTC(),
		CreatedAt: time.Unix(10, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("upsert first note: %v", err)
	}

	err = store.UpsertNote(context.Background(), Note{
		ID:        "note-newer",
		ProfileID: "public",
		Target:    "contract:CDEF",
		Title:     "Newer",
		Body:      "second note",
		UpdatedAt: time.Unix(20, 0).UTC(),
		CreatedAt: time.Unix(20, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("upsert second note: %v", err)
	}

	notes, err := store.ListNotes(context.Background())
	if err != nil {
		t.Fatalf("list notes: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}

	if notes[0].ID != "note-newer" {
		t.Fatalf("expected notes sorted by updated desc, got first note %q", notes[0].ID)
	}
}

func TestUpsertLiveTransactionsAndListLiveTransactions(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)

	db, err := sql.Open(driverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	err = store.UpsertLiveTransactions(context.Background(), []LiveTransaction{
		{ProfileID: "public", Hash: "tx-older", LedgerSequence: 10, Account: "GOLD", CreatedAt: time.Unix(10, 0).UTC()},
		{ProfileID: "public", Hash: "tx-newer", LedgerSequence: 11, Account: "GNEW", IsSoroban: true, CreatedAt: time.Unix(20, 0).UTC()},
	})
	if err != nil {
		t.Fatalf("upsert live transactions: %v", err)
	}

	transactions, err := store.ListLiveTransactions(context.Background(), "public", 10)
	if err != nil {
		t.Fatalf("list live transactions: %v", err)
	}
	if len(transactions) != 2 {
		t.Fatalf("expected 2 live transactions, got %d", len(transactions))
	}
	if transactions[0].Hash != "tx-newer" || !transactions[0].IsSoroban {
		t.Fatalf("unexpected first live transaction: %#v", transactions[0])
	}
}

func TestUpsertEntityCacheAndListEntityCache(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)

	db, err := sql.Open(driverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	err = store.UpsertEntityCache(context.Background(), EntityCache{
		ProfileID: "public",
		Kind:      "transaction",
		Target:    "tx-1",
		Title:     "Tx 1",
		Summary:   "ledger 10",
		Payload:   `{"hash":"tx-1"}`,
	})
	if err != nil {
		t.Fatalf("upsert entity cache: %v", err)
	}

	entities, err := store.ListEntityCache(context.Background(), "public", 10)
	if err != nil {
		t.Fatalf("list entity cache: %v", err)
	}
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity cache row, got %d", len(entities))
	}
	if entities[0].Kind != "transaction" || entities[0].Target != "tx-1" {
		t.Fatalf("unexpected entity cache row: %#v", entities[0])
	}
}

type fakeSQLiteDriver struct{}

type fakeSQLiteConn struct {
	state *fakeSQLiteState
}

type fakeSQLiteState struct {
	mu            sync.Mutex
	versions      []int
	state         map[string]string
	profiles      map[string]Profile
	bookmarks     map[string]Bookmark
	labels        map[string]Label
	labelTargets  map[string]LabelTarget
	notes         map[string]Note
	liveTxs       map[string]LiveTransaction
	entities      map[string]EntityCache
	watchSettings map[string]WatchSetting
	executedSQL   []string
}

type fakeSQLiteTx struct{}

type fakeSQLiteRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func registerFakeSQLiteDriver(t *testing.T) string {
	t.Helper()

	name := fmt.Sprintf("fake-sqlite-%d", time.Now().UnixNano())
	sql.Register(name, fakeSQLiteDriver{})
	return name
}

func (d fakeSQLiteDriver) Open(string) (driver.Conn, error) {
	return &fakeSQLiteConn{
		state: &fakeSQLiteState{
			state:         make(map[string]string),
			profiles:      make(map[string]Profile),
			bookmarks:     make(map[string]Bookmark),
			labels:        make(map[string]Label),
			labelTargets:  make(map[string]LabelTarget),
			notes:         make(map[string]Note),
			liveTxs:       make(map[string]LiveTransaction),
			entities:      make(map[string]EntityCache),
			watchSettings: make(map[string]WatchSetting),
		},
	}, nil
}

func (c *fakeSQLiteConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *fakeSQLiteConn) Close() error                        { return nil }
func (c *fakeSQLiteConn) Begin() (driver.Tx, error)           { return fakeSQLiteTx{}, nil }
func (c *fakeSQLiteConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return fakeSQLiteTx{}, nil
}

func (c *fakeSQLiteConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()

	c.state.executedSQL = append(c.state.executedSQL, normalizeSQL(query))

	switch {
	case strings.Contains(query, "INSERT INTO schema_migrations"):
		if len(args) > 0 {
			c.state.versions = append(c.state.versions, int(asInt64(args[0].Value)))
		}
	case strings.Contains(query, "INSERT INTO app_state"):
		if len(args) >= 2 {
			key, _ := args[0].Value.(string)
			value, _ := args[1].Value.(string)
			c.state.state[key] = value
		}
	case strings.Contains(query, "INSERT INTO profiles"):
		if len(args) >= 7 {
			profile := Profile{
				ID:         asString(args[0].Value),
				Name:       asString(args[1].Value),
				Network:    asString(args[2].Value),
				RPCURL:     asString(args[3].Value),
				IndexerURL: asString(args[4].Value),
				CreatedAt:  asTime(args[5].Value),
				UpdatedAt:  asTime(args[6].Value),
			}
			c.state.profiles[profile.ID] = profile
		}
	case strings.Contains(query, "INSERT INTO bookmarks"):
		if len(args) >= 8 {
			bookmark := Bookmark{
				ID:        asString(args[0].Value),
				ProfileID: asString(args[1].Value),
				Kind:      asString(args[2].Value),
				Target:    asString(args[3].Value),
				Title:     asString(args[4].Value),
				Notes:     asString(args[5].Value),
				CreatedAt: asTime(args[6].Value),
				UpdatedAt: asTime(args[7].Value),
			}
			c.state.bookmarks[bookmark.ID] = bookmark
		}
	case strings.Contains(query, "INSERT INTO labels"):
		if len(args) >= 6 {
			label := Label{
				ID:        asString(args[0].Value),
				ProfileID: asString(args[1].Value),
				Name:      asString(args[2].Value),
				Color:     asString(args[3].Value),
				CreatedAt: asTime(args[4].Value),
				UpdatedAt: asTime(args[5].Value),
			}
			c.state.labels[label.ID] = label
		}
	case strings.Contains(query, "INSERT INTO label_targets"):
		if len(args) >= 7 {
			target := LabelTarget{
				ID:        asString(args[0].Value),
				LabelID:   asString(args[1].Value),
				ProfileID: asString(args[2].Value),
				Kind:      asString(args[3].Value),
				Target:    asString(args[4].Value),
				CreatedAt: asTime(args[5].Value),
				UpdatedAt: asTime(args[6].Value),
			}
			c.state.labelTargets[target.ID] = target
		}
	case strings.Contains(query, "INSERT INTO notes"):
		if len(args) >= 7 {
			note := Note{
				ID:        asString(args[0].Value),
				ProfileID: asString(args[1].Value),
				Target:    asString(args[2].Value),
				Title:     asString(args[3].Value),
				Body:      asString(args[4].Value),
				CreatedAt: asTime(args[5].Value),
				UpdatedAt: asTime(args[6].Value),
			}
			c.state.notes[note.ID] = note
		}
	case strings.Contains(query, "INSERT INTO live_transactions"):
		if len(args) >= 10 {
			tx := LiveTransaction{
				ProfileID:        asString(args[0].Value),
				Hash:             asString(args[1].Value),
				LedgerSequence:   uint32(asInt64(args[2].Value)),
				ApplicationOrder: int32(asInt64(args[3].Value)),
				Account:          asString(args[4].Value),
				OperationCount:   int32(asInt64(args[5].Value)),
				Status:           int16(asInt64(args[6].Value)),
				IsSoroban:        asInt64(args[7].Value) != 0,
				CreatedAt:        asTime(args[8].Value),
				CachedAt:         asTime(args[9].Value),
			}
			c.state.liveTxs[tx.ProfileID+":"+tx.Hash] = tx
		}
	case strings.Contains(query, "INSERT INTO entity_cache"):
		if len(args) >= 9 {
			entity := EntityCache{
				ProfileID:   asString(args[0].Value),
				Kind:        asString(args[1].Value),
				Target:      asString(args[2].Value),
				Title:       asString(args[3].Value),
				Summary:     asString(args[4].Value),
				Payload:     asString(args[5].Value),
				SourceLabel: asString(args[6].Value),
				CreatedAt:   asTime(args[7].Value),
				UpdatedAt:   asTime(args[8].Value),
			}
			c.state.entities[entity.ProfileID+":"+entity.Kind+":"+entity.Target] = entity
		}
	case strings.Contains(query, "INSERT INTO watch_settings"):
		if len(args) >= 8 {
			setting := WatchSetting{
				ID:          asString(args[0].Value),
				ProfileID:   asString(args[1].Value),
				Name:        asString(args[2].Value),
				FiltersJSON: asString(args[3].Value),
				Paused:      asInt64(args[4].Value) != 0,
				AutoApply:   asInt64(args[5].Value) != 0,
				CreatedAt:   asTime(args[6].Value),
				UpdatedAt:   asTime(args[7].Value),
			}
			c.state.watchSettings[setting.ProfileID+":"+setting.Name] = setting
		}
	case strings.Contains(query, "DELETE FROM watch_settings"):
		if len(args) >= 2 {
			delete(c.state.watchSettings, asString(args[0].Value)+":"+asString(args[1].Value))
		}
	}

	return driver.RowsAffected(1), nil
}

func (c *fakeSQLiteConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()

	switch {
	case strings.Contains(query, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations"):
		version := 0
		for _, current := range c.state.versions {
			if current > version {
				version = current
			}
		}
		return &fakeSQLiteRows{
			columns: []string{"version"},
			rows:    [][]driver.Value{{int64(version)}},
		}, nil
	case strings.Contains(query, "SELECT value FROM app_state WHERE key = ?"):
		key := asString(args[0].Value)
		value, ok := c.state.state[key]
		if !ok {
			return &fakeSQLiteRows{columns: []string{"value"}}, nil
		}
		return &fakeSQLiteRows{
			columns: []string{"value"},
			rows:    [][]driver.Value{{value}},
		}, nil
	case strings.Contains(query, "FROM profiles"):
		profiles := make([]Profile, 0, len(c.state.profiles))
		for _, profile := range c.state.profiles {
			profiles = append(profiles, profile)
		}

		sort.Slice(profiles, func(i, j int) bool {
			return profiles[i].Name < profiles[j].Name
		})

		rows := make([][]driver.Value, 0, len(profiles))
		for _, profile := range profiles {
			rows = append(rows, []driver.Value{
				profile.ID,
				profile.Name,
				profile.Network,
				profile.RPCURL,
				profile.IndexerURL,
				profile.CreatedAt,
				profile.UpdatedAt,
			})
		}

		return &fakeSQLiteRows{
			columns: []string{"id", "name", "network", "rpc_url", "indexer_url", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM bookmarks"):
		bookmarks := make([]Bookmark, 0, len(c.state.bookmarks))
		for _, bookmark := range c.state.bookmarks {
			bookmarks = append(bookmarks, bookmark)
		}

		sort.Slice(bookmarks, func(i, j int) bool {
			if bookmarks[i].Title == bookmarks[j].Title {
				return bookmarks[i].CreatedAt.Before(bookmarks[j].CreatedAt)
			}

			return bookmarks[i].Title < bookmarks[j].Title
		})

		rows := make([][]driver.Value, 0, len(bookmarks))
		for _, bookmark := range bookmarks {
			rows = append(rows, []driver.Value{
				bookmark.ID,
				bookmark.ProfileID,
				bookmark.Kind,
				bookmark.Target,
				bookmark.Title,
				bookmark.Notes,
				bookmark.CreatedAt,
				bookmark.UpdatedAt,
			})
		}

		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "kind", "target", "title", "notes", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM labels"):
		labels := make([]Label, 0, len(c.state.labels))
		for _, label := range c.state.labels {
			labels = append(labels, label)
		}

		sort.Slice(labels, func(i, j int) bool {
			if labels[i].Name == labels[j].Name {
				return labels[i].CreatedAt.Before(labels[j].CreatedAt)
			}

			return labels[i].Name < labels[j].Name
		})

		rows := make([][]driver.Value, 0, len(labels))
		for _, label := range labels {
			rows = append(rows, []driver.Value{
				label.ID,
				label.ProfileID,
				label.Name,
				label.Color,
				label.CreatedAt,
				label.UpdatedAt,
			})
		}

		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "name", "color", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM label_targets"):
		targets := make([]LabelTarget, 0, len(c.state.labelTargets))
		for _, target := range c.state.labelTargets {
			targets = append(targets, target)
		}

		sort.Slice(targets, func(i, j int) bool {
			if targets[i].LabelID == targets[j].LabelID {
				if targets[i].Kind == targets[j].Kind {
					return targets[i].Target < targets[j].Target
				}
				return targets[i].Kind < targets[j].Kind
			}
			return targets[i].LabelID < targets[j].LabelID
		})

		rows := make([][]driver.Value, 0, len(targets))
		for _, target := range targets {
			rows = append(rows, []driver.Value{
				target.ID,
				target.LabelID,
				target.ProfileID,
				target.Kind,
				target.Target,
				target.CreatedAt,
				target.UpdatedAt,
			})
		}

		return &fakeSQLiteRows{
			columns: []string{"id", "label_id", "profile_id", "kind", "target", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM notes"):
		notes := make([]Note, 0, len(c.state.notes))
		for _, note := range c.state.notes {
			notes = append(notes, note)
		}

		sort.Slice(notes, func(i, j int) bool {
			if notes[i].UpdatedAt.Equal(notes[j].UpdatedAt) {
				return notes[i].CreatedAt.After(notes[j].CreatedAt)
			}

			return notes[i].UpdatedAt.After(notes[j].UpdatedAt)
		})

		rows := make([][]driver.Value, 0, len(notes))
		for _, note := range notes {
			rows = append(rows, []driver.Value{
				note.ID,
				note.ProfileID,
				note.Target,
				note.Title,
				note.Body,
				note.CreatedAt,
				note.UpdatedAt,
			})
		}

		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "target", "title", "body", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM live_transactions"):
		profileID := asString(args[0].Value)
		limit := int(asInt64(args[1].Value))
		transactions := make([]LiveTransaction, 0, len(c.state.liveTxs))
		for _, tx := range c.state.liveTxs {
			if tx.ProfileID == profileID {
				transactions = append(transactions, tx)
			}
		}

		sort.Slice(transactions, func(i, j int) bool {
			if transactions[i].LedgerSequence == transactions[j].LedgerSequence {
				return transactions[i].ApplicationOrder > transactions[j].ApplicationOrder
			}
			return transactions[i].LedgerSequence > transactions[j].LedgerSequence
		})
		if limit > 0 && len(transactions) > limit {
			transactions = transactions[:limit]
		}

		rows := make([][]driver.Value, 0, len(transactions))
		for _, tx := range transactions {
			rows = append(rows, []driver.Value{
				tx.ProfileID,
				tx.Hash,
				int64(tx.LedgerSequence),
				int64(tx.ApplicationOrder),
				tx.Account,
				int64(tx.OperationCount),
				int64(tx.Status),
				boolToInt(tx.IsSoroban),
				tx.CreatedAt,
				tx.CachedAt,
			})
		}

		return &fakeSQLiteRows{
			columns: []string{"profile_id", "hash", "ledger_sequence", "application_order", "account", "operation_count", "status", "is_soroban", "created_at", "cached_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM entity_cache"):
		profileID := asString(args[0].Value)
		limit := int(asInt64(args[1].Value))
		entities := make([]EntityCache, 0, len(c.state.entities))
		for _, entity := range c.state.entities {
			if entity.ProfileID == profileID {
				entities = append(entities, entity)
			}
		}
		sort.Slice(entities, func(i, j int) bool {
			return entities[i].UpdatedAt.After(entities[j].UpdatedAt)
		})
		if limit > 0 && len(entities) > limit {
			entities = entities[:limit]
		}
		rows := make([][]driver.Value, 0, len(entities))
		for _, entity := range entities {
			rows = append(rows, []driver.Value{
				entity.ProfileID,
				entity.Kind,
				entity.Target,
				entity.Title,
				entity.Summary,
				entity.Payload,
				entity.SourceLabel,
				entity.CreatedAt,
				entity.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"profile_id", "kind", "target", "title", "summary", "payload", "source_label", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM watch_settings"):
		if strings.Contains(query, "auto_apply = 1") {
			profileID := asString(args[0].Value)
			var latest *WatchSetting
			for _, setting := range c.state.watchSettings {
				if setting.ProfileID == profileID && setting.AutoApply {
					if latest == nil || setting.UpdatedAt.After(latest.UpdatedAt) {
						copy := setting
						latest = &copy
					}
				}
			}
			if latest == nil {
				return &fakeSQLiteRows{
					columns: []string{"id", "profile_id", "name", "filters_json", "paused", "auto_apply", "created_at", "updated_at"},
				}, nil
			}
			paused := int64(0)
			if latest.Paused {
				paused = 1
			}
			autoApply := int64(0)
			if latest.AutoApply {
				autoApply = 1
			}
			return &fakeSQLiteRows{
				columns: []string{"id", "profile_id", "name", "filters_json", "paused", "auto_apply", "created_at", "updated_at"},
				rows: [][]driver.Value{{
					latest.ID,
					latest.ProfileID,
					latest.Name,
					latest.FiltersJSON,
					paused,
					autoApply,
					latest.CreatedAt,
					latest.UpdatedAt,
				}},
			}, nil
		}
		profileID := asString(args[0].Value)
		settings := make([]WatchSetting, 0, len(c.state.watchSettings))
		for _, setting := range c.state.watchSettings {
			if setting.ProfileID == profileID {
				settings = append(settings, setting)
			}
		}
		sort.Slice(settings, func(i, j int) bool { return settings[i].UpdatedAt.After(settings[j].UpdatedAt) })
		rows := make([][]driver.Value, 0, len(settings))
		for _, setting := range settings {
			paused := int64(0)
			if setting.Paused {
				paused = 1
			}
			autoApply := int64(0)
			if setting.AutoApply {
				autoApply = 1
			}
			rows = append(rows, []driver.Value{
				setting.ID,
				setting.ProfileID,
				setting.Name,
				setting.FiltersJSON,
				paused,
				autoApply,
				setting.CreatedAt,
				setting.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "name", "filters_json", "paused", "auto_apply", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported query: %s", query)
	}
}

func (fakeSQLiteTx) Commit() error   { return nil }
func (fakeSQLiteTx) Rollback() error { return nil }

func (r *fakeSQLiteRows) Columns() []string {
	return r.columns
}

func (r *fakeSQLiteRows) Close() error {
	return nil
}

func (r *fakeSQLiteRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}

	copy(dest, r.rows[r.index])
	r.index++
	return nil
}

func normalizeSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

func asInt64(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func asString(value any) string {
	text, _ := value.(string)
	return text
}

func asTime(value any) time.Time {
	if ts, ok := value.(time.Time); ok {
		return ts
	}

	return time.Time{}
}
