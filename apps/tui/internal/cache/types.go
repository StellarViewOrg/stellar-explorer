package cache

import "time"

// Profile stores the local connection profile used by the terminal app.
type Profile struct {
	ID         string
	Name       string
	Network    string
	RPCURL     string
	IndexerURL string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Bookmark stores a saved entity or view the user wants to revisit.
type Bookmark struct {
	ID        string
	ProfileID string
	Kind      string
	Target    string
	Title     string
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Label stores a user-defined tag for an entity or workflow grouping.
type Label struct {
	ID        string
	ProfileID string
	Name      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LabelTarget attaches a label definition to one executable entity.
type LabelTarget struct {
	ID        string
	LabelID   string
	ProfileID string
	Kind      string
	Target    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Note stores free-form user notes attached to a target entity or workflow item.
type Note struct {
	ID        string
	ProfileID string
	Target    string
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LiveTransaction stores one recent live-feed transaction for local revisit flows.
type LiveTransaction struct {
	ProfileID        string
	Hash             string
	LedgerSequence   uint32
	ApplicationOrder int32
	Account          string
	OperationCount   int32
	Status           int16
	IsSoroban        bool
	CreatedAt        time.Time
	CachedAt         time.Time
}

// EntityCache stores one visited lookup payload for quick local revisit flows.
type EntityCache struct {
	ProfileID   string
	Kind        string
	Target      string
	Title       string
	Summary     string
	Payload     string
	SourceLabel string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
