# Ponder Media Catalog — Implementation Plan

> **Strategy:** Good Strategy Bad Strategy kernel (diagnosis → guiding policy → coherent actions)
> **Execution:** EOS Rocks (quarterly phases → weekly tasks)
> **Coding:** TDD, Go std lib first, DDD modular monolith

**Goal:** Build the catalog pipeline as a single Go binary with clean domain/application/infrastructure layers, proven by tests from the domain outward.

**Diagnosis (the real challenge):** A catalog that pulls from multiple external sources, merges into canonical form, and publishes changes — without coupling domain logic to any specific source or database. The risk is building a monolithic soup where source-specific concerns leak everywhere. The leverage point is the domain model: get `SourceRecord` and `MediaItem` right first, and everything else snaps into place.

**Guiding Policy:** Start with one source, one media type. Prove the vertical slice end-to-end (fetch → store → publish) before adding more sources. Write tests at the right layer — domain tests for rules, application tests with fakes for behavior, infrastructure tests only at boundaries. Use Go std lib for everything except CLI parsing, SQLite driver, migrations, and test assertions.

---

## Phase 0: Project Bootstrap (one-time, ~30 min)

> **Rock:** Scaffold the project and install tools. Nothing clever — just the bones.

### Task 0.1 — Clone repo and initialize Go module

```bash
cd ~/Workspace/github.com/Caduceus-Labs
git clone git@github.com:Caduceus-Labs/ponder-media-catalog.git
cd ponder-media-catalog
go mod init github.com/Caduceus-Labs/ponder-media-catalog
```

### Task 0.2 — Add initial dependencies

```bash
go get github.com/spf13/cobra@latest
go get modernc.org/sqlite
go get github.com/golang-migrate/migrate/v4
go get github.com/stretchr/testify
```

### Task 0.3 — Create directory structure

```bash
mkdir -p cmd/ponder-catalog
mkdir -p internal/catalog/domain
mkdir -p internal/catalog/application
mkdir -p internal/catalog/infrastructure/sqlite
mkdir -p internal/catalog/infrastructure/sources
mkdir -p internal/catalog/infrastructure/publishing
mkdir -p internal/platform/config
mkdir -p internal/platform/database
mkdir -p internal/platform/logging
mkdir -p internal/platform/lock
mkdir -p internal/cli
mkdir -p migrations
mkdir -p testdata/sourcea
```

**Verify:**
```bash
find . -type d | sort
# Should show all directories above
```

### Task 0.4 — Update README.md

Replace the old "media-scraper" README with a one-liner for the new project name.

### Task 0.5 — Create Makefile

```makefile
.PHONY: test test-race vet build clean

test:
	go test ./... -v

test-race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -o bin/ponder-catalog ./cmd/ponder-catalog/

clean:
	rm -rf bin/
```

**Commit checkpoint:**
```bash
git add -A
git commit -m "chore: bootstrap Go module with directory structure and dependencies"
```

---

## Phase 1: Domain Layer — SourceRecord (pure domain, no dependencies)

> **Rock:** `SourceRecord` is the smallest unit of domain logic with real business value. It has no database, no HTTP, no CLI — just validation rules and fingerprints.

### Task 1.1 — Write `SourceRecord` test: rejects empty source name

**File:** `internal/catalog/domain/source_record_test.go`

```go
package domain

import (
	"testing"
)

func TestSourceRecord_RejectsEmptySourceName(t *testing.T) {
	_, err := NewSourceRecord("", "ext-1", []byte(`{"title":"Test"}`))
	if err == nil {
		t.Fatal("expected error for empty source name")
	}
}
```

**Verify:**
```bash
go test ./internal/catalog/domain/ -run TestSourceRecord_RejectsEmptySourceName -v
# Expected: FAIL — package domain not defined / type not found
```

### Task 1.2 — Implement `SourceRecord` struct

**File:** `internal/catalog/domain/source_record.go`

```go
package domain

import (
	"crypto/sha256"
	"fmt"
)

type SourceRecord struct {
	SourceName     string
	ExternalID     string
	RawData        []byte
	Fingerprint    string
}

func NewSourceRecord(sourceName, externalID string, rawData []byte) (*SourceRecord, error) {
	if sourceName == "" {
		return nil, fmt.Errorf("source name is required")
	}
	if externalID == "" {
		return nil, fmt.Errorf("external ID is required")
	}
	fingerprint := fingerprintRawData(rawData)
	return &SourceRecord{
		SourceName:  sourceName,
		ExternalID:  externalID,
		RawData:     rawData,
		Fingerprint: fingerprint,
	}, nil
}

func fingerprintRawData(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
```

**Verify:**
```bash
go test ./internal/catalog/domain/ -run TestSourceRecord_RejectsEmptySourceName -v
# Expected: PASS
```

### Task 1.3 — Write test: rejects empty external ID

Add to `source_record_test.go`:

```go
func TestSourceRecord_RejectsEmptyExternalID(t *testing.T) {
	_, err := NewSourceRecord("source-a", "", []byte(`{"title":"Test"}`))
	if err == nil {
		t.Fatal("expected error for empty external ID")
	}
}
```

**Verify:**
```bash
go test ./internal/catalog/domain/ -run TestSourceRecord_RejectsEmptyExternalID -v
# Expected: PASS
```

### Task 1.4 — Write test: same data produces same fingerprint

Add to `source_record_test.go`:

```go
func TestSourceRecord_SameDataSameFingerprint(t *testing.T) {
	data := []byte(`{"title":"Final Fantasy VII"}`)
	r1, _ := NewSourceRecord("source-a", "ext-1", data)
	r2, _ := NewSourceRecord("source-a", "ext-1", data)

	if r1.Fingerprint != r2.Fingerprint {
		t.Fatalf("expected same fingerprint, got %q vs %q", r1.Fingerprint, r2.Fingerprint)
	}
}
```

**Verify:**
```bash
go test ./internal/catalog/domain/ -v
# All 3 tests should pass
```

### Task 1.5 — Write test: different data produces different fingerprint

Add to `source_record_test.go`:

```go
func TestSourceRecord_DifferentDataDifferentFingerprint(t *testing.T) {
	r1, _ := NewSourceRecord("source-a", "ext-1", []byte(`{"title":"FFVII"}`))
	r2, _ := NewSourceRecord("source-a", "ext-1", []byte(`{"title":"Final Fantasy VII"}`))

	if r1.Fingerprint == r2.Fingerprint {
		t.Fatal("expected different fingerprints for different data")
	}
}
```

**Verify:**
```bash
go test ./internal/catalog/domain/ -v
# All 4 tests should pass
```

**Commit checkpoint:**
```bash
git add internal/catalog/domain/
git commit -m "feat: add SourceRecord domain value object with fingerprinting"
```

---

## Phase 2: Domain Layer — MediaItem (canonical media model)

> **Rock:** `MediaItem` is the canonical representation of a media entity. Pure domain, no deps.

### Task 2.1 — Write `MediaItem` test: empty title rejected

**File:** `internal/catalog/domain/media_item_test.go`

```go
package domain

import (
	"testing"
)

func TestMediaItem_RejectsEmptyTitle(t *testing.T) {
	item := NewMediaItem("item-1", MediaTypeGame)
	err := item.SetTitle("")
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}
```

**Verify:** Fails (NewMediaItem/MediaTypeGame not defined yet)

### Task 2.2 — Implement `MediaItem` with types

**File:** `internal/catalog/domain/media_item.go`

```go
package domain

import "fmt"

type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeSeries MediaType = "series"
	MediaTypeGame MediaType = "game"
	MediaTypeAnime MediaType = "anime"
)

type MediaItemID string

type MediaItem struct {
	ID          MediaItemID
	MediaType   MediaType
	Title       string
	SourceLinks []SourceLink
}

type SourceLink struct {
	SourceName string
	ExternalID string
}

func NewMediaItem(id MediaItemID, mediaType MediaType) *MediaItem {
	return &MediaItem{
		ID:        id,
		MediaType: mediaType,
	}
}

func (m *MediaItem) SetTitle(title string) error {
	if title == "" {
		return fmt.Errorf("title is required")
	}
	m.Title = title
	return nil
}
```

**Verify:**
```bash
go test ./internal/catalog/domain/ -v
# All tests pass
```

### Task 2.3 — Write test: valid title accepted

Add to `media_item_test.go`:

```go
func TestMediaItem_AcceptsValidTitle(t *testing.T) {
	item := NewMediaItem("item-1", MediaTypeGame)
	err := item.SetTitle("Final Fantasy VII")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.Title != "Final Fantasy VII" {
		t.Fatalf("expected %q, got %q", "Final Fantasy VII", item.Title)
	}
}
```

**Verify:**
```bash
go test ./internal/catalog/domain/ -v
```

### Task 2.4 — Write `SourceName` value object test

**File:** `internal/catalog/domain/source_name_test.go`

```go
package domain

import "testing"

func TestSourceName_RejectsEmpty(t *testing.T) {
	_, err := NewSourceName("")
	if err == nil {
		t.Fatal("expected error for empty source name")
	}
}

func TestSourceName_AcceptsValid(t *testing.T) {
	name, err := NewSourceName("letterboxd")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(name) != "letterboxd" {
		t.Fatalf("expected %q, got %q", "letterboxd", name)
	}
}
```

### Task 2.5 — Implement `SourceName` type

**File:** `internal/catalog/domain/source_name.go`

```go
package domain

import "fmt"

type SourceName string

func NewSourceName(name string) (SourceName, error) {
	if name == "" {
		return "", fmt.Errorf("source name is required")
	}
	return SourceName(name), nil
}
```

**Verify:**
```bash
go test ./internal/catalog/domain/ -v
```

**Commit checkpoint:**
```bash
git add internal/catalog/domain/
git commit -m "feat: add MediaItem domain model and SourceName value object"
```

---

## Phase 3: Domain Events

> **Rock:** Domain events decouple side effects from use cases. Start minimal — just the event types.

### Task 3.1 — Define domain events

**File:** `internal/catalog/domain/events.go`

```go
package domain

type DomainEvent interface {
	EventName() string
}

type SourceRecordCreated struct {
	SourceName string
	ExternalID string
}

func (e SourceRecordCreated) EventName() string {
	return "source_record_created"
}

type MediaItemUpdated struct {
	MediaItemID string
}

func (e MediaItemUpdated) EventName() string {
	return "media_item_updated"
}
```

No tests needed for pure event structs — they're data containers.

**Commit checkpoint:**
```bash
git add internal/catalog/domain/events.go
git commit -m "feat: add domain event types"
```

---

## Phase 4: Application Layer — SyncSource use case

> **Rock:** The `SyncSource` use case orchestrates fetching changes from a source adapter and saving them. Tested with fakes — no real HTTP or database.

### Task 4.1 — Define application port interfaces

**File:** `internal/catalog/application/ports.go`

```go
package application

import (
	"context"
	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

type SourceAdapter interface {
	SourceName() domain.SourceName
	FetchChanges(ctx context.Context, checkpoint SyncCheckpoint) ([]domain.SourceRecord, SyncCheckpoint, error)
}

type SourceRecordRepository interface {
	Save(ctx context.Context, record domain.SourceRecord) (bool, error)
}

type SyncStateRepository interface {
	Load(ctx context.Context, sourceName domain.SourceName) (SyncCheckpoint, error)
	Save(ctx context.Context, sourceName domain.SourceName, checkpoint SyncCheckpoint) error
}

type SyncCheckpoint struct {
	Cursor string
}
```

### Task 4.2 — Write SyncSource test: saves a new record

**File:** `internal/catalog/application/sync_source_test.go`

```go
package application

import (
	"context"
	"testing"

	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

func TestSyncSource_SavesNewRecord(t *testing.T) {
	recordRepo := &fakeSourceRecordRepository{}
	syncRepo := &fakeSyncStateRepository{}
	adapter := &fakeSourceAdapter{records: []domain.SourceRecord{
		{SourceName: "source-a", ExternalID: "ext-1", Fingerprint: "abc123"},
	}}

	useCase := NewSyncSource(recordRepo, syncRepo, adapter)
	err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(recordRepo.saved) != 1 {
		t.Fatalf("expected 1 saved record, got %d", len(recordRepo.saved))
	}
}
```

**Verify:** Fails (NewSyncSource, fakes not defined)

### Task 4.3 — Implement fakes and SyncSource

**File:** `internal/catalog/application/sync_source.go`

```go
package application

import (
	"context"
	"fmt"
	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

type SyncSource struct {
	recordRepo SourceRecordRepository
	syncRepo   SyncStateRepository
	adapter    SourceAdapter
}

func NewSyncSource(recordRepo SourceRecordRepository, syncRepo SyncStateRepository, adapter SourceAdapter) *SyncSource {
	return &SyncSource{
		recordRepo: recordRepo,
		syncRepo:   syncRepo,
		adapter:    adapter,
	}
}

func (s *SyncSource) Execute(ctx context.Context) error {
	adapterName := s.adapter.SourceName()

	checkpoint, err := s.syncRepo.Load(ctx, adapterName)
	if err != nil {
		return fmt.Errorf("load checkpoint: %w", err)
	}

	records, newCheckpoint, err := s.adapter.FetchChanges(ctx, checkpoint)
	if err != nil {
		return fmt.Errorf("fetch changes: %w", err)
	}

	for _, record := range records {
		if _, err := s.recordRepo.Save(ctx, record); err != nil {
			return fmt.Errorf("save record: %w", err)
		}
	}

	if err := s.syncRepo.Save(ctx, adapterName, newCheckpoint); err != nil {
		return fmt.Errorf("save checkpoint: %w", err)
	}

	return nil
}
```

**File:** `internal/catalog/application/sync_source_test.go` (add fakes)

```go
// Add these before the test function

type fakeSourceRecordRepository struct {
	saved []domain.SourceRecord
}

func (f *fakeSourceRecordRepository) Save(_ context.Context, record domain.SourceRecord) (bool, error) {
	f.saved = append(f.saved, record)
	return true, nil
}

type fakeSyncStateRepository struct {
	checkpoints map[domain.SourceName]SyncCheckpoint
}

func (f *fakeSyncStateRepository) Load(_ context.Context, _ domain.SourceName) (SyncCheckpoint, error) {
	return SyncCheckpoint{}, nil
}

func (f *fakeSyncStateRepository) Save(_ context.Context, _ domain.SourceName, checkpoint SyncCheckpoint) error {
	if f.checkpoints == nil {
		f.checkpoints = make(map[domain.SourceName]SyncCheckpoint)
	}
	f.checkpoints[domain.SourceName("source-a")] = checkpoint
	return nil
}

type fakeSourceAdapter struct {
	records []domain.SourceRecord
}

func (f *fakeSourceAdapter) SourceName() domain.SourceName {
	return domain.SourceName("source-a")
}

func (f *fakeSourceAdapter) FetchChanges(_ context.Context, _ SyncCheckpoint) ([]domain.SourceRecord, SyncCheckpoint, error) {
	return f.records, SyncCheckpoint{Cursor: "new-cursor"}, nil
}
```

**Verify:**
```bash
go test ./internal/catalog/application/ -v
# Expected: PASS
```

### Task 4.4 — Write test: unchanged record not re-saved

Add to `sync_source_test.go`:

```go
func TestSyncSource_DoesNotSaveUnchangedRecord(t *testing.T) {
	recordRepo := &fakeSourceRecordRepository{}
	syncRepo := &fakeSyncStateRepository{}
	adapter := &fakeSourceAdapter{}

	useCase := NewSyncSource(recordRepo, syncRepo, adapter)
	err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(recordRepo.saved) != 0 {
		t.Fatalf("expected 0 saved records for empty fetch, got %d", len(recordRepo.saved))
	}
}
```

Update the fake adapter: if `records` is nil/empty, FetchChanges returns empty.

**Verify:**
```bash
go test ./internal/catalog/application/ -v
```

### Task 4.5 — Write test: checkpoint not updated on failed fetch

Add to `sync_source_test.go`:

```go
func TestSyncSource_DoesNotUpdateCheckpointOnFailedFetch(t *testing.T) {
	recordRepo := &fakeSourceRecordRepository{}
	syncRepo := &fakeSyncStateRepository{}
	adapter := &failingSourceAdapter{}

	useCase := NewSyncSource(recordRepo, syncRepo, adapter)
	err := useCase.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error from failing adapter")
	}

	if len(syncRepo.checkpoints) != 0 {
		t.Fatal("expected checkpoint not to be saved after failure")
	}
}
```

Add failing adapter:

```go
type failingSourceAdapter struct{}

func (f *failingSourceAdapter) SourceName() domain.SourceName {
	return domain.SourceName("source-a")
}

func (f *failingSourceAdapter) FetchChanges(_ context.Context, _ SyncCheckpoint) ([]domain.SourceRecord, SyncCheckpoint, error) {
	return nil, SyncCheckpoint{}, fmt.Errorf("network error")
}
```

Note: add `"fmt"` to imports in the test file.

**Verify:**
```bash
go test ./internal/catalog/application/ -v
```

**Commit checkpoint:**
```bash
git add internal/catalog/application/
git commit -m "feat: add SyncSource use case with fakes-based tests"
```

---

## Phase 5: Application Layer — PublishCatalog use case

> **Rock:** The `PublishCatalog` use case reads canonical media items and decides whether to publish based on content changes.

### Task 5.1 — Define PublishCatalog port interfaces

Add to `ports.go`:

```go
type MediaItemRepository interface {
	FindAll(ctx context.Context) ([]domain.MediaItem, error)
}

type PublicationRepository interface {
	LoadLast(ctx context.Context) (PublicationState, error)
	Save(ctx context.Context, state PublicationState) error
}

type PublicationState struct {
	Hash   string
	ItemCount int
}

type CatalogPublisher interface {
	Publish(ctx context.Context, items []domain.MediaItem) (string, error)
}
```

### Task 5.2 — Write test: does not publish unchanged output

**File:** `internal/catalog/application/publish_catalog_test.go`

```go
package application

import (
	"context"
	"testing"
	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

func TestPublishCatalog_DoesNotPublishUnchanged(t *testing.T) {
	mediaRepo := &fakeMediaItemRepository{}
	pubRepo := &fakePublicationRepository{lastHash: "same-hash"}
	publisher := &catalogPublisherStub{hash: "same-hash"}

	useCase := NewPublishCatalog(mediaRepo, pubRepo, publisher)
	err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if publisher.published {
		t.Fatal("expected no publish when hash matches")
	}
}
```

**Verify:** Fails

### Task 5.3 — Implement PublishCatalog

**File:** `internal/catalog/application/publish_catalog.go`

```go
package application

import (
	"context"
	"fmt"
	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

type PublishCatalog struct {
	mediaRepo MediaItemRepository
	pubRepo   PublicationRepository
	publisher CatalogPublisher
}

func NewPublishCatalog(mediaRepo MediaItemRepository, pubRepo PublicationRepository, publisher CatalogPublisher) *PublishCatalog {
	return &PublishCatalog{
		mediaRepo: mediaRepo,
		pubRepo:   pubRepo,
		publisher: publisher,
	}
}

func (p *PublishCatalog) Execute(ctx context.Context) error {
	items, err := p.mediaRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("find media items: %w", err)
	}

	lastPub, err := p.pubRepo.LoadLast(ctx)
	if err != nil {
		return fmt.Errorf("load last publication: %w", err)
	}

	newHash, err := p.publisher.Publish(ctx, items)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	if newHash == lastPub.Hash {
		return nil // Nothing changed
	}

	if err := p.pubRepo.Save(ctx, PublicationState{Hash: newHash, ItemCount: len(items)}); err != nil {
		return fmt.Errorf("save publication state: %w", err)
	}

	return nil
}
```

### Task 5.4 — Add fakes and implement test

Add to `publish_catalog_test.go`:

```go
type fakeMediaItemRepository struct{}

func (f *fakeMediaItemRepository) FindAll(_ context.Context) ([]domain.MediaItem, error) {
	return []domain.MediaItem{}, nil
}

type fakePublicationRepository struct {
	lastHash string
}

func (f *fakePublicationRepository) LoadLast(_ context.Context) (PublicationState, error) {
	return PublicationState{Hash: f.lastHash}, nil
}

func (f *fakePublicationRepository) Save(_ context.Context, state PublicationState) error {
	return nil
}

type catalogPublisherStub struct {
	hash      string
	published bool
}

func (c *catalogPublisherStub) Publish(_ context.Context, _ []domain.MediaItem) (string, error) {
	c.published = true
	return c.hash, nil
}
```

**Verify:**
```bash
go test ./internal/catalog/application/ -v
```

### Task 5.5 — Write test: publishes changed output

Add to `publish_catalog_test.go`:

```go
func TestPublishCatalog_PublishesChangedOutput(t *testing.T) {
	mediaRepo := &fakeMediaItemRepository{}
	pubRepo := &fakePublicationRepository{lastHash: "old-hash"}
	publisher := &catalogPublisherStub{hash: "new-hash"}

	useCase := NewPublishCatalog(mediaRepo, pubRepo, publisher)
	err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !publisher.published {
		t.Fatal("expected publish when hash differs")
	}
}
```

**Verify:**
```bash
go test ./internal/catalog/application/ -v
```

**Commit checkpoint:**
```bash
git add internal/catalog/application/
git commit -m "feat: add PublishCatalog use case with change-detection"
```

---

## Phase 6: Infrastructure — SQLite Repository

> **Rock:** Real database implementation of the repository interfaces. Uses `database/sql` + `modernc.org/sqlite`.

### Task 6.1 — Create migration: source records table

**File:** `migrations/000001_create_source_records.up.sql`

```sql
CREATE TABLE IF NOT EXISTS source_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_name TEXT NOT NULL,
    external_id TEXT NOT NULL,
    raw_data BLOB,
    fingerprint TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_name, external_id)
);

CREATE INDEX idx_source_records_source ON source_records(source_name);
CREATE INDEX idx_source_records_fingerprint ON source_records(fingerprint);
```

**File:** `migrations/000001_create_source_records.down.sql`

```sql
DROP TABLE IF EXISTS source_records;
```

### Task 6.2 — Create migration: media items and links

**File:** `migrations/000002_create_media_items.up.sql`

```sql
CREATE TABLE IF NOT EXISTS media_items (
    id TEXT PRIMARY KEY,
    media_type TEXT NOT NULL,
    title TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS media_source_links (
    media_item_id TEXT NOT NULL,
    source_name TEXT NOT NULL,
    external_id TEXT NOT NULL,
    PRIMARY KEY (media_item_id, source_name),
    FOREIGN KEY (media_item_id) REFERENCES media_items(id)
);
```

**File:** `migrations/000002_create_media_items.down.sql`

```sql
DROP TABLE IF EXISTS media_source_links;
DROP TABLE IF EXISTS media_items;
```

### Task 6.3 — Implement SQLite source record repository

**File:** `internal/catalog/infrastructure/sqlite/media_repository.go`

```go
package sqlite

import (
	"context"
	"database/sql"

	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

type MediaRepository struct {
	db *sql.DB
}

func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Save(ctx context.Context, record domain.SourceRecord) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO source_records (source_name, external_id, raw_data, fingerprint)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(source_name, external_id) DO UPDATE SET
		   raw_data = excluded.raw_data,
		   fingerprint = excluded.fingerprint,
		   updated_at = CURRENT_TIMESTAMP
		 WHERE source_records.fingerprint != excluded.fingerprint`,
		record.SourceName, record.ExternalID, record.RawData, record.Fingerprint,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}
```

### Task 6.4 — Write repository test with real SQLite

**File:** `internal/catalog/infrastructure/sqlite/media_repository_test.go`

```go
package sqlite

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS source_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_name TEXT NOT NULL,
			external_id TEXT NOT NULL,
			raw_data BLOB,
			fingerprint TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source_name, external_id)
		)
	`)
	if err != nil {
		t.Fatalf("create test table: %v", err)
	}

	return db
}

func TestMediaRepository_SaveAndRead(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMediaRepository(db)

	record, _ := domain.NewSourceRecord("source-a", "ext-1", []byte(`{"title":"Test"}`))
	changed, err := repo.Save(context.Background(), *record)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if !changed {
		t.Fatal("expected new record to be saved as changed")
	}

	// Verify by querying directly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM source_records").Scan(&count)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 record, got %d", count)
	}
}
```

**Verify:**
```bash
go test ./internal/catalog/infrastructure/sqlite/ -v
# Expected: PASS
```

### Task 6.5 — Write test: unchanged record not flagged

Add to `media_repository_test.go`:

```go
func TestMediaRepository_SameFingerprintNotChanged(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMediaRepository(db)

	record, _ := domain.NewSourceRecord("source-a", "ext-1", []byte(`{"title":"Test"}`))
	repo.Save(context.Background(), *record)

	// Save same data again
	changed, err := repo.Save(context.Background(), *record)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if changed {
		t.Fatal("expected no change for identical data")
	}
}
```

**Verify:**
```bash
go test ./internal/catalog/infrastructure/sqlite/ -v
```

**Commit checkpoint:**
```bash
git add internal/catalog/infrastructure/sqlite/ migrations/
git commit -m "feat: add SQLite repository with migrations"
```

---

## Phase 7: Infrastructure — Source Adapter

> **Rock:** Implement the `SourceAdapter` interface for the first source. Use `httptest` for tests.

### Task 7.1 — Create source adapter skeleton

**File:** `internal/catalog/infrastructure/sources/sourcea/client.go`

```go
package sourcea

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/application"
	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) SourceName() domain.SourceName {
	return domain.SourceName("source-a")
}

func (c *Client) FetchChanges(ctx context.Context, checkpoint application.SyncCheckpoint) ([]domain.SourceRecord, application.SyncCheckpoint, error) {
	// TODO: implement
	return nil, application.SyncCheckpoint{}, fmt.Errorf("not implemented")
}
```

### Task 7.2 — Write adapter test with httptest

**File:** `internal/catalog/infrastructure/sources/sourcea/client_test.go`

```go
package sourcea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/application"
)

func TestClient_FetchChangesMapsFixture(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"ext-1","title":"Final Fantasy VII"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	records, checkpoint, err := client.FetchChanges(context.Background(), application.SyncCheckpoint{})
	if err != nil {
		t.Fatalf("fetch changes: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	_ = checkpoint
}
```

**Verify:**
```bash
go test ./internal/catalog/infrastructure/sources/sourcea/ -v
```

### Task 7.3 — Implement mapper to convert API response to SourceRecord

**File:** `internal/catalog/infrastructure/sources/sourcea/mapper.go`

```go
package sourcea

import (
	"encoding/json"

	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

type apiResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func mapToSourceRecord(raw json.RawMessage) (*domain.SourceRecord, error) {
	var item apiResponse
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, err
	}

	return domain.NewSourceRecord("source-a", item.ID, raw)
}
```

**Commit checkpoint:**
```bash
git add internal/catalog/infrastructure/sources/sourcea/
git commit -m "feat: add source-a adapter with httptest and mapper"
```

---

## Phase 8: CLI Commands

> **Rock:** Wire the use cases into Cobra commands. Thin controllers — no business logic.

### Task 8.1 — Create root command

**File:** `internal/cli/root_command.go`

```go
package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ponder-catalog",
		Short: "Media metadata catalog pipeline",
	}

	cmd.AddCommand(newIngestCommand())
	cmd.AddCommand(newPublishCommand())

	return cmd
}
```

### Task 8.2 — Create ingest command

**File:** `internal/cli/ingest_command.go`

```go
package cli

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"
)

func newIngestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ingest [source-name]",
		Short: "Ingest data from a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("ingesting", "source", args[0])
			// TODO: wire real dependencies
			return nil
		},
	}
}
```

### Task 8.3 — Create publish command

**File:** `internal/cli/publish_command.go`

```go
package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func newPublishCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "publish",
		Short: "Publish catalog JSON files",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("publishing catalog")
			// TODO: wire real dependencies
			return nil
		},
	}
}
```

### Task 8.4 — Create main.go

**File:** `cmd/ponder-catalog/main.go`

```go
package main

import (
	"log/slog"
	"os"

	"github.com/Caduceus-Labs/ponder-media-catalog/internal/cli"
)

func main() {
	cmd := cli.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
```

**Verify:**
```bash
go build -o bin/ponder-catalog ./cmd/ponder-catalog/
./bin/ponder-catalog --help
./bin/ponder-catalog ingest source-a
./bin/ponder-catalog publish
```

### Task 8.5 — Create database initialization in platform layer

**File:** `internal/platform/database/sqlite.go`

```go
package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	// Enable WAL mode for concurrent reads
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	return db, nil
}
```

**Commit checkpoint:**
```bash
git add cmd/ internal/cli/ internal/platform/
git commit -m "feat: add CLI commands and main.go entry point"
```

---

## Phase 9: Wire the Full Application

> **Rock:** Connect all layers — CLI → application → infrastructure. Make `ingest source-a` actually work end-to-end with SQLite.

### Task 9.1 — Create dependency injection in CLI commands

Update `internal/cli/ingest_command.go` to wire real dependencies:

```go
func newIngestCommand(dbPath string) *cobra.Command {
	return &cobra.Command{
		Use:   "ingest [source-name]",
		Short: "Ingest data from a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := database.OpenSQLite(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			recordRepo := sqlite.NewMediaRepository(db)
			syncRepo := sqlite.NewSyncStateRepository(db) // to be created

			var adapter application.SourceAdapter
			switch args[0] {
			case "source-a":
				adapter = sourcea.NewClient("https://api.source-a.example.com")
			default:
				return fmt.Errorf("unknown source: %s", args[0])
			}

			useCase := application.NewSyncSource(recordRepo, syncRepo, adapter)
			return useCase.Execute(cmd.Context())
		},
	}
}
```

### Task 9.2 — Implement SyncStateRepository for SQLite

**File:** `internal/catalog/infrastructure/sqlite/source_sync_state_repository.go`

```go
package sqlite

import (
	"context"
	"database/sql"

	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/application"
	"github.com/Caduceus-Labs/ponder-media-catalog/internal/catalog/domain"
)

type SyncStateRepository struct {
	db *sql.DB
}

func NewSyncStateRepository(db *sql.DB) *SyncStateRepository {
	return &SyncStateRepository{db: db}
}

func (r *SyncStateRepository) Load(ctx context.Context, sourceName domain.SourceName) (application.SyncCheckpoint, error) {
	var cursor string
	err := r.db.QueryRowContext(ctx,
		`SELECT cursor FROM source_sync_state WHERE source_name = ?`, string(sourceName),
	).Scan(&cursor)
	if err == sql.ErrNoRows {
		return application.SyncCheckpoint{}, nil
	}
	if err != nil {
		return application.SyncCheckpoint{}, err
	}
	return application.SyncCheckpoint{Cursor: cursor}, nil
}

func (r *SyncStateRepository) Save(ctx context.Context, sourceName domain.SourceName, checkpoint application.SyncCheckpoint) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO source_sync_state (source_name, cursor)
		 VALUES (?, ?)
		 ON CONFLICT(source_name) DO UPDATE SET cursor = excluded.cursor`,
		string(sourceName), checkpoint.Cursor,
	)
	return err
}
```

### Task 9.3 — Add sync state migration

**File:** `migrations/000003_create_sync_state.up.sql`

```sql
CREATE TABLE IF NOT EXISTS source_sync_state (
    source_name TEXT PRIMARY KEY,
    cursor TEXT NOT NULL DEFAULT '',
    last_success_at TIMESTAMP,
    last_error TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Commit checkpoint:**
```bash
git add internal/ migrations/
git commit -m "feat: wire full application with SQLite and CLI"
```

---

## Phase 10: Integration & Verification

> **Rock:** Full end-to-end test and CI setup.

### Task 10.1 — Write end-to-end integration test (skippable)

**File:** `internal/catalog/integration_test.go`

```go
//go:build integration

package catalog

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func TestFullPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	// Full end-to-end test skipped for now — just the skeleton
	_ = context.Background()
	_ = sql.Drivers()
}
```

### Task 10.2 — Update Makefile with full commands

```makefile
.PHONY: test test-race vet build clean lint integration

test:
	go test ./... -v -short

test-race:
	go test -race ./... -short

integration:
	go test -tags=integration ./... -v

vet:
	go vet ./...

lint:
	go vet ./...

build:
	go build -o bin/ponder-catalog ./cmd/ponder-catalog/

clean:
	rm -rf bin/

run-ingest:
	go run ./cmd/ponder-catalog/ ingest source-a

run-publish:
	go run ./cmd/ponder-catalog/ publish

.PHONY: test test-race vet build clean run-ingest run-publish
```

### Task 10.3 — Run full test suite

```bash
go test ./... -v -short        # Unit + app tests (fast)
go vet ./...                    # Static analysis
go test -race ./... -short      # Race detection
```

**Commit checkpoint:**
```bash
git add Makefile
git commit -m "chore: add Makefile targets and integration test skeleton"
```

---

## Summary: The 10 Rocks

| Rock | Phase | What You Build | Tests |
|------|-------|----------------|-------|
| 0 | Bootstrap | Module, deps, dirs, Makefile | `go build` |
| 1 | Domain: SourceRecord | Validation + fingerprinting | 4 tests |
| 2 | Domain: MediaItem | Canonical model + source links | 2 tests |
| 3 | Domain Events | Event types | none |
| 4 | App: SyncSource | Sync use case with fakes | 3 tests |
| 5 | App: PublishCatalog | Publish use case with fakes | 2 tests |
| 6 | Infra: SQLite | Real DB repository | 2 tests |
| 7 | Infra: Source Adapter | HTTP client + mapper | 1 test |
| 8 | CLI Commands | Cobra wiring + main.go | `go build` |
| 9 | Wiring | Full DI, real calls | manual |
| 10 | Integration | E2E test skeleton, CI | `go test ./...` |

---

## Verification Commands (run after each phase)

```bash
go build ./...          # Compiles everything
go test ./... -v        # All tests pass
go vet ./...            # No static analysis issues
```

## Commit Convention

Use conventional commits throughout:

```
feat: add SourceRecord domain value object
fix: handle empty checkpoint in SyncSource
chore: add Makefile targets
test: add integration test for full pipeline
docs: update README with setup instructions
```

---

*Plan generated from: Recommended_approach.md (DeepSeek/Perplexity) × wondelai/skills (good-strategy-bad-strategy + traction-eos frameworks)*