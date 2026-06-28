package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/searchinfer"
	"github.com/stellar/go-stellar-sdk/strkey"
)

// Screen identifies a top-level application view.
type Screen string

const (
	ScreenHome     Screen = "home"
	ScreenLiveFeed Screen = "live-feed"
	ScreenLookup   Screen = "lookup"
	ScreenSettings Screen = "settings"
)

var validScreens = map[Screen]struct{}{
	ScreenHome:     {},
	ScreenLiveFeed: {},
	ScreenLookup:   {},
	ScreenSettings: {},
}

// StatusLevel is a small presentation hint for the renderer.
type StatusLevel string

const (
	StatusInfo  StatusLevel = "info"
	StatusWarn  StatusLevel = "warn"
	StatusError StatusLevel = "error"
)

// Status describes the current outcome of the last user action.
type Status struct {
	Level   StatusLevel
	Message string
}

// ViewState identifies a coarse loading lifecycle for renderer-facing panes.
type ViewState string

const (
	ViewStateIdle    ViewState = "idle"
	ViewStateLoading ViewState = "loading"
	ViewStateReady   ViewState = "ready"
	ViewStateEmpty   ViewState = "empty"
	ViewStateError   ViewState = "error"
)

// Snapshot is a renderer-friendly view of the current app state.
type Snapshot struct {
	ProductName string
	ConfigPath  string
	Current     Screen
	History     []Screen
	Forward     []Screen
	Status      Status
	Focus       FocusArea
	HelpVisible bool
	Profile     config.Profile
	Cache       CacheSnapshot
	LiveFeed    LiveFeedSnapshot
	Lookup      LookupSnapshot
	LookupRoute []LookupRouteStep
	Selection   SelectionSnapshot
	Command     CommandPaletteSnapshot
	Help        []string
}

// SourceMetadata describes how a given payload was resolved.
type SourceMetadata struct {
	Mode           string
	Operation      string
	Policy         string
	Preferred      string
	Actual         string
	Label          string
	CacheState     string
	FallbackUsed   bool
	Degraded       bool
	DegradedReason string
}

// FocusArea identifies which major region currently owns keyboard focus.
type FocusArea string

const (
	FocusMain    FocusArea = "main"
	FocusSidebar FocusArea = "sidebar"
	FocusStatus  FocusArea = "status"
)

// SelectionSnapshot contains renderer-facing selection state.
type SelectionSnapshot struct {
	LiveFeedIndex        int
	LiveFeedScrollOffset int
}

// TransactionSummarySelection is a renderer/runtime-friendly selected transaction.
type TransactionSummarySelection struct {
	Index      int
	Hash       string
	Ledger     uint32
	Account    string
	Ops        int32
	HasAccount bool
	IsSoroban  bool
	CreatedAt  time.Time
	StatusCode int16
}

// CommandPaletteSnapshot contains renderer-facing command/search overlay state.
type CommandPaletteSnapshot struct {
	Visible       bool
	Prompt        string
	Input         string
	Results       []SearchResult
	SelectedIndex int
	SearchLimit   int
}

// SearchResult is a locally inferred search result shown in the command palette.
type SearchResult struct {
	Kind        string
	Title       string
	Description string
	Command     string
	Enabled     bool
	Source      string
}

// LookupKind identifies a typed entity lookup.
type LookupKind string

const (
	LookupLedger      LookupKind = "ledger"
	LookupTransaction LookupKind = "transaction"
	LookupOperation   LookupKind = "operation"
	LookupAccount     LookupKind = "account"
	LookupAsset       LookupKind = "asset"
	LookupContract    LookupKind = "contract"
	LookupEvent       LookupKind = "event"
	LookupStorage     LookupKind = "storage"
)

type LookupExplorerKind string

const (
	LookupExplorerTransactions LookupExplorerKind = "transactions"
	LookupExplorerOperations   LookupExplorerKind = "operations"
	LookupExplorerInvocations  LookupExplorerKind = "invocations"
	LookupExplorerHolders      LookupExplorerKind = "holders"
	LookupExplorerEvents       LookupExplorerKind = "events"
	LookupExplorerStorage      LookupExplorerKind = "storage"
	LookupExplorerResults      LookupExplorerKind = "results"
	LookupExplorerTimeline     LookupExplorerKind = "timeline"
)

type LookupExplorerSnapshot struct {
	Kind         LookupExplorerKind
	Title        string
	ParentLabel  string
	BackCommand  string
	NextCommand  string
	ListLimit    int
	ListOffset   int
	Transactions []backendclient.TransactionSummary
	Operations   []backendclient.OperationSummary
	Holders      []backendclient.AssetHolderSummary
	Events       []backendclient.ContractEventSummary
	Storage      []backendclient.ContractStorageSummary
	Results      []SearchResult
}

type LookupMetadataSnapshot struct {
	Labels    []LookupLabelSnapshot
	Bookmarks []LookupBookmarkSnapshot
	Notes     []LookupNoteSnapshot
	Cached    *LookupCacheSnapshot
}

type LookupLabelSnapshot struct {
	Name  string
	Color string
}

type LookupBookmarkSnapshot struct {
	Title string
	Notes string
}

type LookupNoteSnapshot struct {
	Title string
	Body  string
}

type LookupCacheSnapshot struct {
	UpdatedAt   time.Time
	SourceLabel string
	Summary     string
}

type ContractDecodeMode string

const (
	ContractDecodeModeDecoded ContractDecodeMode = "decoded"
	ContractDecodeModeRaw     ContractDecodeMode = "raw"
)

// LookupSnapshot is a renderer-friendly view of the current lookup result.
type LookupSnapshot struct {
	Query           string
	Kind            LookupKind
	BackendURL      string
	Source          SourceMetadata
	State           ViewState
	Error           string
	SelectedSection int
	ScrollOffset    int
	Explorer        *LookupExplorerSnapshot
	Ledger          *backendclient.LedgerLookupResponse
	Transaction     *backendclient.TransactionLookupResponse
	Operation       *backendclient.OperationLookupSnapshot
	Account         *backendclient.AccountLookupResponse
	Asset           *backendclient.AssetLookupResponse
	Contract        *backendclient.ContractLookupResponse
	Event           *backendclient.ContractEventLookupSnapshot
	StorageEntry    *backendclient.ContractStorageLookupSnapshot
	DecodeMode      ContractDecodeMode
	ContractTab     ContractWorkspaceTab
	ExpandAll       bool
	VisualMode      bool
	Metadata        LookupMetadataSnapshot
	ReturnContext   *LookupReturnContext
}

// LiveFeedService loads live feed data from the configured backend.
type LiveFeedService interface {
	LiveFeedSummary(ctx context.Context) (backendclient.LiveFeedSummaryResponse, error)
}

const (
	LiveFeedFilterAll     = "all"
	LiveFeedFilterSoroban = "soroban"
	LiveFeedFilterClassic = "classic"

	defaultLiveFeedScrollback = 100
)

// LiveFeedSnapshot is a renderer-friendly view of backend live data.
type LiveFeedSnapshot struct {
	Configured         bool
	Available          bool
	State              ViewState
	BackendURL         string
	Source             SourceMetadata
	SourceMode         string
	LastIngestedLedger uint32
	LatestLedger       *backendclient.LedgerSummary
	RecentTransactions []backendclient.TransactionSummary
	Scrollback         []backendclient.TransactionSummary
	Paused             bool
	Filter             string
	ReplayOffset       int
	ScrollbackCount    int
	MaxScrollback      int
	LastUpdateCount    int
	Error              string
}

// CacheSnapshot is a renderer-friendly summary of local persistence state.
type CacheSnapshot struct {
	Enabled      bool
	Available    bool
	Path         string
	Schema       int
	Profiles     int
	LastScreen   string
	Status       string
	DefaultID    string
	DefaultLabel string
}

// Model tracks the terminal client's top-level state machine.
type Model struct {
	config                config.Config
	configPath            string
	current               Screen
	history               []Screen
	forward               []Screen
	status                Status
	focus                 FocusArea
	helpVisible           bool
	cache                 CacheSnapshot
	liveFeed              LiveFeedSnapshot
	lookup                LookupSnapshot
	lookupHistory         []LookupSnapshot
	lookupForward         []LookupSnapshot
	selection             SelectionSnapshot
	command               CommandPaletteSnapshot
	liveFeedAll           []backendclient.TransactionSummary
	liveFeedPendingStream []LiveFeedStreamUpdate
	liveFeedFilterSpec    LiveFeedFilterSpec
	liveFeedMonitoring    LiveFeedMonitoringContext
	defaultWatchApplied   bool
}

// NewModel creates a runnable app model.
func NewModel(cfg config.Config, configPath string, cacheSnapshot CacheSnapshot) *Model {
	profile, ok := cfg.Profile(cfg.DefaultProfile)
	status := Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded profile %q on %s via %s.", profile.Name, profile.Network, profile.BackendMode),
	}
	if !ok {
		status = Status{
			Level:   StatusWarn,
			Message: "No default profile available. Check local configuration.",
		}
	}

	return &Model{
		config:     cfg,
		configPath: configPath,
		current:    ScreenHome,
		focus:      FocusMain,
		status:     status,
		cache:      cacheSnapshot,
		liveFeed: LiveFeedSnapshot{
			Configured:    profileHasDataBackend(profile),
			State:         ViewStateIdle,
			BackendURL:    activeBackendLabel(profile),
			Source:        DefaultSourceMetadata(profile, "live-feed"),
			SourceMode:    LiveFeedSourcePoll,
			Filter:        LiveFeedFilterAll,
			MaxScrollback: defaultLiveFeedScrollback,
		},
		liveFeedFilterSpec: defaultLiveFeedFilterSpec(),
	}
}

// Snapshot returns an immutable view for rendering.
func (m *Model) Snapshot() Snapshot {
	profile, _ := m.config.Profile(m.config.DefaultProfile)

	return Snapshot{
		ProductName: m.config.ProductName,
		ConfigPath:  m.configPath,
		Current:     m.current,
		History:     append([]Screen(nil), m.history...),
		Forward:     append([]Screen(nil), m.forward...),
		Status:      m.status,
		Focus:       m.focus,
		HelpVisible: m.helpVisible,
		Profile:     profile,
		Cache:       m.cache,
		LiveFeed:    m.liveFeed,
		Lookup:      m.lookup,
		LookupRoute: m.buildLookupRoute(),
		Selection:   m.selection,
		Command:     m.command,
		Help: []string{
			"h home",
			"l live feed",
			"u lookup",
			"s settings",
			"b back (screen or lookup route)",
			"f forward (screen or lookup route)",
			"r refresh",
			"/ search",
			"tab focus",
			"y copy current entity",
			"ctrl+v paste clipboard in search",
			"? help",
			"j/k or arrows move",
			"enter follow selection",
			"route line shows lookup breadcrumbs",
			"q quit",
		},
	}
}

// RefreshLiveFeed fetches the backend summary used by the live screen.
func (m *Model) RefreshLiveFeed(ctx context.Context, service LiveFeedService, source ...SourceMetadata) error {
	return m.refreshLiveFeed(ctx, service, true, source...)
}

// RefreshLiveFeedMetadata fetches ledger metadata without merging polled transactions.
func (m *Model) RefreshLiveFeedMetadata(ctx context.Context, service LiveFeedService, source ...SourceMetadata) error {
	return m.refreshLiveFeed(ctx, service, false, source...)
}

func (m *Model) refreshLiveFeed(ctx context.Context, service LiveFeedService, mergeTransactions bool, source ...SourceMetadata) error {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	m.liveFeed.Configured = profileHasDataBackend(profile)
	m.liveFeed.BackendURL = activeBackendLabel(profile)
	m.liveFeed.Source = firstSourceMetadata(source, profile, "live-feed")

	if !m.liveFeed.Configured {
		m.liveFeed.Available = false
		m.liveFeed.State = ViewStateError
		m.liveFeed.Error = "no indexer_url or rpc_endpoint is configured for the current profile"
		m.liveFeed.LatestLedger = nil
		m.liveFeed.RecentTransactions = nil
		m.liveFeed.Source.Degraded = true
		m.liveFeed.Source.DegradedReason = m.liveFeed.Error
		return fmt.Errorf("no backend is configured")
	}

	if service == nil {
		m.liveFeed.Available = false
		m.liveFeed.State = ViewStateError
		m.liveFeed.Error = "backend client is unavailable"
		m.liveFeed.LatestLedger = nil
		m.liveFeed.RecentTransactions = nil
		m.liveFeed.Source.Degraded = true
		m.liveFeed.Source.DegradedReason = m.liveFeed.Error
		return fmt.Errorf("backend client is unavailable")
	}

	summary, err := service.LiveFeedSummary(ctx)
	if err != nil {
		m.liveFeed.Available = false
		m.liveFeed.State = ViewStateError
		m.liveFeed.Error = err.Error()
		m.liveFeed.LatestLedger = nil
		m.liveFeed.RecentTransactions = nil
		m.liveFeed.Source.Degraded = true
		if m.liveFeed.Source.DegradedReason == "" {
			m.liveFeed.Source.DegradedReason = err.Error()
		}
		m.status = Status{
			Level:   StatusError,
			Message: fmt.Sprintf("Live feed unavailable from %s: %v", valueOrFallbackSourceLabel(m.liveFeed.Source, m.liveFeed.BackendURL), err),
		}
		return err
	}

	m.liveFeed.Available = true
	m.liveFeed.Error = ""
	m.liveFeed.LastIngestedLedger = summary.LastIngestedLedger
	m.liveFeed.LatestLedger = summary.LatestLedger
	added := 0
	if mergeTransactions {
		added = m.mergeLiveFeedTransactions(summary.RecentTransactions)
		m.applyLiveFeedView(added)
		if m.liveFeed.SourceMode != LiveFeedSourceStream {
			m.liveFeed.SourceMode = LiveFeedSourcePoll
		}
	}

	message := fmt.Sprintf("Live feed synced from %s.", m.liveFeed.BackendURL)
	if summary.LatestLedger != nil {
		if mergeTransactions {
			message = fmt.Sprintf("Live feed synced at ledger %d with %d new item(s).", summary.LatestLedger.Sequence, added)
		} else {
			message = fmt.Sprintf("Live feed metadata synced at ledger %d.", summary.LatestLedger.Sequence)
		}
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: message,
	}

	return nil
}

// RestoreLiveFeedScrollback seeds the live feed from locally cached transactions.
func (m *Model) RestoreLiveFeedScrollback(transactions []backendclient.TransactionSummary) {
	m.liveFeedAll = nil
	_ = m.mergeLiveFeedTransactions(transactions)
	m.applyLiveFeedView(0)
	if len(m.liveFeed.RecentTransactions) > 0 {
		m.liveFeed.Available = true
		m.liveFeed.State = ViewStateReady
	}
}

func (m *Model) mergeLiveFeedTransactions(transactions []backendclient.TransactionSummary) int {
	if len(transactions) == 0 {
		return 0
	}
	seen := make(map[string]struct{}, len(m.liveFeedAll)+len(transactions))
	for _, tx := range m.liveFeedAll {
		if key := liveFeedTransactionKey(tx); key != "" {
			seen[key] = struct{}{}
		}
	}
	added := 0
	merged := make([]backendclient.TransactionSummary, 0, len(m.liveFeedAll)+len(transactions))
	for _, tx := range transactions {
		key := liveFeedTransactionKey(tx)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, tx)
		added++
	}
	merged = append(merged, m.liveFeedAll...)
	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].LedgerSequence != merged[j].LedgerSequence {
			return merged[i].LedgerSequence > merged[j].LedgerSequence
		}
		if !merged[i].CreatedAt.Equal(merged[j].CreatedAt) {
			return merged[i].CreatedAt.After(merged[j].CreatedAt)
		}
		if merged[i].ApplicationOrder != merged[j].ApplicationOrder {
			return merged[i].ApplicationOrder > merged[j].ApplicationOrder
		}
		return merged[i].Hash < merged[j].Hash
	})
	if len(merged) > defaultLiveFeedScrollback {
		merged = merged[:defaultLiveFeedScrollback]
	}
	m.liveFeedAll = merged
	return added
}

func liveFeedTransactionKey(tx backendclient.TransactionSummary) string {
	return strings.TrimSpace(tx.Hash)
}

func normalizeLiveFeedFilter(filter string) string {
	switch strings.ToLower(strings.TrimSpace(filter)) {
	case LiveFeedFilterSoroban:
		return LiveFeedFilterSoroban
	case LiveFeedFilterClassic:
		return LiveFeedFilterClassic
	default:
		return LiveFeedFilterAll
	}
}

// SetLiveFeedPaused controls whether automatic live-feed refreshes should update the scrollback.
func (m *Model) SetLiveFeedPaused(paused bool) {
	m.liveFeed.Paused = paused
	if !paused {
		m.liveFeed.ReplayOffset = 0
		m.flushPendingLiveFeedStreamUpdates()
		m.applyLiveFeedView(0)
	}
	state := "resumed"
	if paused {
		state = "paused"
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: "Live feed " + state + ".",
	}
}

func (m *Model) LiveFeedPaused() bool {
	return m.liveFeed.Paused
}

// SetLiveFeedFilter applies a transaction-class filter to the retained live scrollback.
func (m *Model) SetLiveFeedFilter(filter string) error {
	spec, err := parseLiveFeedFilterCommand([]string{"filter", filter})
	if err != nil {
		return err
	}
	m.setLiveFeedFilterSpec(spec)
	return nil
}

// SetLiveFeedFilterCommand applies a parsed live filter command.
func (m *Model) SetLiveFeedFilterCommand(fields []string) error {
	spec, err := parseLiveFeedFilterCommand(fields)
	if err != nil {
		return err
	}
	m.setLiveFeedFilterSpec(spec)
	return nil
}

func (m *Model) CycleLiveFeedFilter() {
	spec := m.liveFeedFilterSpec
	spec.Class = nextLiveFeedFilterClass(spec.Class)
	m.setLiveFeedFilterSpec(spec)
}

// UpdateLookupTransaction stores the latest transaction lookup result.
func (m *Model) UpdateLookupTransaction(query string, response backendclient.TransactionLookupResponse, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(LookupTransaction))
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:       strings.TrimSpace(query),
		Kind:        LookupTransaction,
		BackendURL:  strings.TrimSpace(resolvedSource.Label),
		Source:      resolvedSource,
		State:       ViewStateReady,
		Explorer:    nil,
		Transaction: &response,
	}
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded transaction %s from %s.", strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// UpdateLookupOperation stores a dedicated operation detail view.
func (m *Model) UpdateLookupOperation(query string, snapshot backendclient.OperationLookupSnapshot, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(LookupOperation))
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:      strings.TrimSpace(query),
		Kind:       LookupOperation,
		BackendURL: strings.TrimSpace(resolvedSource.Label),
		Source:     resolvedSource,
		State:      ViewStateReady,
		Explorer:   nil,
		Operation:  &snapshot,
	}
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded operation %s from %s.", strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// UpdateLookupLedger stores the latest ledger lookup result.
func (m *Model) UpdateLookupLedger(query string, response backendclient.LedgerLookupResponse, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(LookupLedger))
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:      strings.TrimSpace(query),
		Kind:       LookupLedger,
		BackendURL: strings.TrimSpace(resolvedSource.Label),
		Source:     resolvedSource,
		State:      ViewStateReady,
		Explorer:   nil,
		Ledger:     &response,
	}
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded ledger %s from %s.", strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// UpdateLookupAccount stores the latest account lookup result.
func (m *Model) UpdateLookupAccount(query string, response backendclient.AccountLookupResponse, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(LookupAccount))
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:      strings.TrimSpace(query),
		Kind:       LookupAccount,
		BackendURL: strings.TrimSpace(resolvedSource.Label),
		Source:     resolvedSource,
		State:      ViewStateReady,
		Explorer:   nil,
		Account:    &response,
	}
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded account %s from %s.", strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// UpdateLookupAsset stores the latest asset lookup result.
func (m *Model) UpdateLookupAsset(query string, response backendclient.AssetLookupResponse, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(LookupAsset))
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:      strings.TrimSpace(query),
		Kind:       LookupAsset,
		BackendURL: strings.TrimSpace(resolvedSource.Label),
		Source:     resolvedSource,
		State:      ViewStateReady,
		Explorer:   nil,
		Asset:      &response,
	}
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded asset %s from %s.", strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// UpdateLookupContract stores the latest contract lookup result.
func (m *Model) UpdateLookupContract(query string, response backendclient.ContractLookupResponse, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(LookupContract))
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:       strings.TrimSpace(query),
		Kind:        LookupContract,
		BackendURL:  strings.TrimSpace(resolvedSource.Label),
		Source:      resolvedSource,
		State:       ViewStateReady,
		Explorer:    nil,
		Contract:    &response,
		DecodeMode:  ContractDecodeModeDecoded,
		ContractTab: ContractWorkspaceTabOverview,
	}
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded contract %s from %s.", strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// SetContractDecodeMode switches contract-heavy lookup views between decoded and raw payloads.
func (m *Model) SetContractDecodeMode(mode ContractDecodeMode) error {
	switch m.lookup.Kind {
	case LookupContract:
		if m.lookup.Contract == nil {
			return errors.New("decode mode is only available from a contract lookup")
		}
	case LookupEvent:
		if m.lookup.Event == nil {
			return errors.New("decode mode is only available from a contract event detail view")
		}
	case LookupStorage:
		if m.lookup.StorageEntry == nil {
			return errors.New("decode mode is only available from a contract storage detail view")
		}
	case LookupOperation:
		if m.lookup.Operation == nil {
			return errors.New("decode mode is only available from a Soroban operation detail view")
		}
		if strings.TrimSpace(m.lookup.Operation.Operation.TypeName) != "invoke_host_function" {
			return errors.New("decode mode is only available for invoke_host_function operations")
		}
	default:
		return errors.New("decode mode is only available from contract, event, storage, or Soroban operation views")
	}
	switch mode {
	case ContractDecodeModeDecoded, ContractDecodeModeRaw:
	default:
		return fmt.Errorf("unknown decode mode %q", mode)
	}
	m.lookup.DecodeMode = mode
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Contract decode mode: %s.", mode),
	}
	return nil
}

// SetLookupMetadata attaches local labels, bookmarks, notes, and cache status to the current lookup.
func (m *Model) SetLookupMetadata(metadata LookupMetadataSnapshot) {
	m.lookup.Metadata = metadata
}

// SetLookupLoading records the requested lookup before a blocking backend call.
func (m *Model) SetLookupLoading(kind LookupKind, query string, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(kind))
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:      strings.TrimSpace(query),
		Kind:       kind,
		BackendURL: strings.TrimSpace(resolvedSource.Label),
		Source:     resolvedSource,
		State:      ViewStateLoading,
		Explorer:   nil,
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loading %s %s via %s.", kind, strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// SetLookupError keeps the lookup screen active while surfacing a backend or parse failure.
func (m *Model) SetLookupError(kind LookupKind, query string, err error, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(kind))
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query: strings.TrimSpace(query),
		Kind:  kind,
		Source: SourceMetadata{
			Mode:           resolvedSource.Mode,
			Operation:      resolvedSource.Operation,
			Policy:         resolvedSource.Policy,
			Preferred:      resolvedSource.Preferred,
			Actual:         resolvedSource.Actual,
			Label:          resolvedSource.Label,
			FallbackUsed:   resolvedSource.FallbackUsed,
			Degraded:       true,
			DegradedReason: valueOrDefault(resolvedSource.DegradedReason, err.Error()),
		},
		State:    ViewStateError,
		Error:    err.Error(),
		Explorer: nil,
	}
	m.status = Status{
		Level:   StatusError,
		Message: fmt.Sprintf("Lookup failed for %s %s: %v", kind, strings.TrimSpace(query), err),
	}
}

// OpenLookupTransactionExplorer switches the lookup pane into a dedicated transaction list mode.
func (m *Model) OpenLookupTransactionExplorer(title string, backCommand string, transactions []backendclient.TransactionSummary, limit, offset int, nextCommand ...string) {
	if m.lookup.State != ViewStateReady {
		return
	}
	m.pushLookupHistory()
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup.SelectedSection = 0
	m.lookup.ScrollOffset = 0
	m.lookup.Explorer = &LookupExplorerSnapshot{
		Kind:         LookupExplorerTransactions,
		Title:        strings.TrimSpace(title),
		ParentLabel:  string(m.lookup.Kind) + " " + strings.TrimSpace(m.lookup.Query),
		BackCommand:  strings.TrimSpace(backCommand),
		NextCommand:  firstOptionalString(nextCommand),
		ListLimit:    limit,
		ListOffset:   offset,
		Transactions: append([]backendclient.TransactionSummary(nil), transactions...),
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Opened transaction explorer for %s.", strings.TrimSpace(title)),
	}
}

// OpenLookupOperationExplorer switches the lookup pane into a dedicated operation list mode.
func (m *Model) OpenLookupOperationExplorer(title string, backCommand string, operations []backendclient.OperationSummary, limit, offset int, nextCommand ...string) {
	if m.lookup.State != ViewStateReady {
		return
	}
	m.pushLookupHistory()
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup.SelectedSection = 0
	m.lookup.ScrollOffset = 0
	m.lookup.Explorer = &LookupExplorerSnapshot{
		Kind:        LookupExplorerOperations,
		Title:       strings.TrimSpace(title),
		ParentLabel: string(m.lookup.Kind) + " " + strings.TrimSpace(m.lookup.Query),
		BackCommand: strings.TrimSpace(backCommand),
		NextCommand: firstOptionalString(nextCommand),
		ListLimit:   limit,
		ListOffset:  offset,
		Operations:  append([]backendclient.OperationSummary(nil), operations...),
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Opened operation explorer for %s.", strings.TrimSpace(title)),
	}
}

// OpenLookupHolderExplorer switches the lookup pane into a dedicated asset-holder list mode.
func (m *Model) OpenLookupHolderExplorer(title string, backCommand string, holders []backendclient.AssetHolderSummary, limit, offset int, nextCommand ...string) {
	if m.lookup.State != ViewStateReady {
		return
	}
	m.pushLookupHistory()
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup.SelectedSection = 0
	m.lookup.ScrollOffset = 0
	m.lookup.Explorer = &LookupExplorerSnapshot{
		Kind:        LookupExplorerHolders,
		Title:       strings.TrimSpace(title),
		ParentLabel: string(m.lookup.Kind) + " " + strings.TrimSpace(m.lookup.Query),
		BackCommand: strings.TrimSpace(backCommand),
		NextCommand: firstOptionalString(nextCommand),
		ListLimit:   limit,
		ListOffset:  offset,
		Holders:     append([]backendclient.AssetHolderSummary(nil), holders...),
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Opened holder explorer for %s.", strings.TrimSpace(title)),
	}
}

// UpdateLookupEvent stores a dedicated contract event detail view.
func (m *Model) UpdateLookupEvent(query string, snapshot backendclient.ContractEventLookupSnapshot, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(LookupEvent))
	decodeMode := m.lookup.DecodeMode
	if decodeMode == "" {
		decodeMode = ContractDecodeModeDecoded
	}
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:      strings.TrimSpace(query),
		Kind:       LookupEvent,
		BackendURL: strings.TrimSpace(resolvedSource.Label),
		Source:     resolvedSource,
		State:      ViewStateReady,
		Event:      &snapshot,
		DecodeMode: decodeMode,
	}
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded event %s from %s.", strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// UpdateLookupStorageEntry stores a dedicated contract storage entry detail view.
func (m *Model) UpdateLookupStorageEntry(query string, snapshot backendclient.ContractStorageLookupSnapshot, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, string(LookupStorage))
	decodeMode := m.lookup.DecodeMode
	if decodeMode == "" {
		decodeMode = ContractDecodeModeDecoded
	}
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:        strings.TrimSpace(query),
		Kind:         LookupStorage,
		BackendURL:   strings.TrimSpace(resolvedSource.Label),
		Source:       resolvedSource,
		State:        ViewStateReady,
		StorageEntry: &snapshot,
		DecodeMode:   decodeMode,
	}
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Loaded storage %s from %s.", strings.TrimSpace(query), valueOrFallbackSourceLabel(resolvedSource, "backend")),
	}
}

// OpenLookupEventExplorer switches the lookup pane into a dedicated contract-event list mode.
func (m *Model) OpenLookupEventExplorer(title string, backCommand string, events []backendclient.ContractEventSummary, limit, offset int, nextCommand ...string) {
	if m.lookup.State != ViewStateReady {
		return
	}
	m.pushLookupHistory()
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup.SelectedSection = 0
	m.lookup.ScrollOffset = 0
	m.lookup.Explorer = &LookupExplorerSnapshot{
		Kind:        LookupExplorerEvents,
		Title:       strings.TrimSpace(title),
		ParentLabel: string(m.lookup.Kind) + " " + strings.TrimSpace(m.lookup.Query),
		BackCommand: strings.TrimSpace(backCommand),
		NextCommand: firstOptionalString(nextCommand),
		ListLimit:   limit,
		ListOffset:  offset,
		Events:      append([]backendclient.ContractEventSummary(nil), events...),
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Opened event explorer for %s.", strings.TrimSpace(title)),
	}
}

// OpenLookupStorageExplorer switches the lookup pane into a dedicated contract-storage list mode.
func (m *Model) OpenLookupStorageExplorer(title string, backCommand string, entries []backendclient.ContractStorageSummary, limit, offset int, nextCommand ...string) {
	if m.lookup.State != ViewStateReady {
		return
	}
	m.pushLookupHistory()
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup.SelectedSection = 0
	m.lookup.ScrollOffset = 0
	m.lookup.Explorer = &LookupExplorerSnapshot{
		Kind:        LookupExplorerStorage,
		Title:       strings.TrimSpace(title),
		ParentLabel: string(m.lookup.Kind) + " " + strings.TrimSpace(m.lookup.Query),
		BackCommand: strings.TrimSpace(backCommand),
		NextCommand: firstOptionalString(nextCommand),
		ListLimit:   limit,
		ListOffset:  offset,
		Storage:     append([]backendclient.ContractStorageSummary(nil), entries...),
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Opened storage explorer for %s.", strings.TrimSpace(title)),
	}
}

// OpenLookupInvocationExplorer switches the lookup pane into a dedicated contract-invocation list mode.
func (m *Model) OpenLookupInvocationExplorer(title string, backCommand string, operations []backendclient.OperationSummary, limit, offset int, nextCommand ...string) {
	if m.lookup.State != ViewStateReady {
		return
	}
	m.pushLookupHistory()
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup.SelectedSection = 0
	m.lookup.ScrollOffset = 0
	m.lookup.Explorer = &LookupExplorerSnapshot{
		Kind:        LookupExplorerInvocations,
		Title:       strings.TrimSpace(title),
		ParentLabel: string(m.lookup.Kind) + " " + strings.TrimSpace(m.lookup.Query),
		BackCommand: strings.TrimSpace(backCommand),
		NextCommand: firstOptionalString(nextCommand),
		ListLimit:   limit,
		ListOffset:  offset,
		Operations:  append([]backendclient.OperationSummary(nil), operations...),
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Opened invocation explorer for %s.", strings.TrimSpace(title)),
	}
}

func firstOptionalString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

// OpenLookupResultExplorer switches the lookup pane into a generic entity list mode.
func (m *Model) OpenLookupResultExplorer(title string, parentLabel string, backCommand string, results []SearchResult, limit, offset int, source ...SourceMetadata) {
	profile, _ := m.config.Profile(m.config.DefaultProfile)
	resolvedSource := firstSourceMetadata(source, profile, "open")
	m.pushLookupHistory()
	m.pushHistory(ScreenLookup)
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup = LookupSnapshot{
		Query:      strings.TrimSpace(title),
		BackendURL: strings.TrimSpace(resolvedSource.Label),
		Source:     resolvedSource,
		State:      ViewStateReady,
		Explorer: &LookupExplorerSnapshot{
			Kind:        LookupExplorerResults,
			Title:       strings.TrimSpace(title),
			ParentLabel: strings.TrimSpace(parentLabel),
			BackCommand: strings.TrimSpace(backCommand),
			ListLimit:   limit,
			ListOffset:  offset,
			Results:     append([]SearchResult(nil), results...),
		},
	}
	resetLookupSelection(&m.lookup)
	if len(results) == 0 {
		m.lookup.State = ViewStateEmpty
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Opened %s explorer with %d result(s).", strings.TrimSpace(title), len(results)),
	}
}

// OpenLookupTimelineExplorer switches the lookup pane into a combined activity timeline.
func (m *Model) OpenLookupTimelineExplorer(title string, backCommand string, results []SearchResult, limit, offset int, nextCommand ...string) {
	if m.lookup.State != ViewStateReady {
		return
	}
	m.pushLookupHistory()
	m.current = ScreenLookup
	m.helpVisible = false
	m.lookup.SelectedSection = 0
	m.lookup.ScrollOffset = 0
	m.lookup.Explorer = &LookupExplorerSnapshot{
		Kind:        LookupExplorerTimeline,
		Title:       strings.TrimSpace(title),
		ParentLabel: string(m.lookup.Kind) + " " + strings.TrimSpace(m.lookup.Query),
		BackCommand: strings.TrimSpace(backCommand),
		NextCommand: firstOptionalString(nextCommand),
		ListLimit:   limit,
		ListOffset:  offset,
		Results:     append([]SearchResult(nil), results...),
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Opened timeline for %s.", strings.TrimSpace(title)),
	}
}

// CloseLookupExplorer returns the lookup pane to its detail view without refetching.
func (m *Model) CloseLookupExplorer() {
	if m.lookup.Explorer == nil {
		return
	}
	m.pushLookupHistory()
	label := strings.TrimSpace(m.lookup.Explorer.ParentLabel)
	m.lookup.Explorer = nil
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Returned to %s detail.", valueOrDefault(label, "lookup")),
	}
}

// ShouldApplyDefaultWatch reports whether profile watch presets should run on live entry.
func (m *Model) ShouldApplyDefaultWatch() bool {
	return !m.defaultWatchApplied && m.current == ScreenLiveFeed
}

// MarkDefaultWatchApplied records that a profile watch preset has been applied.
func (m *Model) MarkDefaultWatchApplied() {
	m.defaultWatchApplied = true
}

// SetScreen moves to a top-level screen.
func (m *Model) SetScreen(screen Screen) error {
	if _, ok := validScreens[screen]; !ok {
		return fmt.Errorf("unknown screen %q", screen)
	}
	m.pushHistory(screen)

	m.current = screen
	m.helpVisible = false
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Switched to %s.", screen),
	}

	return nil
}

// RestoreScreen rehydrates the last persisted screen when it is valid.
func (m *Model) RestoreScreen(screen Screen) error {
	if err := m.SetScreen(screen); err != nil {
		return err
	}

	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Restored screen %s from local cache.", screen),
	}
	m.cache.LastScreen = string(screen)

	return nil
}

// Back moves to the previous screen or previous lookup entity.
func (m *Model) Back() bool {
	if m.current == ScreenLookup && len(m.lookupHistory) == 0 && m.lookup.ReturnContext != nil {
		switch m.lookup.ReturnContext.Screen {
		case ScreenLiveFeed:
			return m.ReturnToLiveMonitoring()
		}
	}

	if m.current == ScreenLookup && len(m.lookupHistory) > 0 {
		prev := m.lookupHistory[len(m.lookupHistory)-1]
		m.lookupHistory = m.lookupHistory[:len(m.lookupHistory)-1]
		if m.lookup.State == ViewStateReady {
			m.lookupForward = append(m.lookupForward, m.lookup)
		}
		m.lookup = prev
		m.helpVisible = false
		m.status = Status{
			Level:   StatusInfo,
			Message: fmt.Sprintf("Returned to %s %s.", m.lookup.Kind, m.lookup.Query),
		}
		return true
	}

	if len(m.history) == 0 {
		m.status = Status{
			Level:   StatusWarn,
			Message: "No previous screen available.",
		}
		return false
	}

	last := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	if m.current != "" {
		m.forward = append(m.forward, m.current)
	}
	m.current = last
	m.helpVisible = false
	if last == ScreenLiveFeed {
		m.RestoreLiveFeedMonitoringContext()
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Returned to %s.", last),
	}
	return true
}

// Forward moves to the next screen when back navigation populated forward history.
func (m *Model) Forward() bool {
	if m.current == ScreenLookup && len(m.lookupForward) > 0 {
		next := m.lookupForward[len(m.lookupForward)-1]
		m.lookupForward = m.lookupForward[:len(m.lookupForward)-1]
		if m.lookup.State == ViewStateReady {
			m.lookupHistory = append(m.lookupHistory, m.lookup)
		}
		m.lookup = next
		m.helpVisible = false
		m.status = Status{
			Level:   StatusInfo,
			Message: fmt.Sprintf("Moved forward to %s %s.", m.lookup.Kind, m.lookup.Query),
		}
		return true
	}

	if len(m.forward) == 0 {
		m.status = Status{
			Level:   StatusWarn,
			Message: "No next screen available.",
		}
		return false
	}

	next := m.forward[len(m.forward)-1]
	m.forward = m.forward[:len(m.forward)-1]
	if m.current != "" {
		m.history = append(m.history, m.current)
	}
	m.current = next
	m.helpVisible = false
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Moved forward to %s.", next),
	}
	return true
}

func (m *Model) pushHistory(next Screen) {
	if m.current == "" || m.current == next {
		return
	}
	m.history = append(m.history, m.current)
	m.forward = nil
}

func (m *Model) pushLookupHistory() {
	if m.lookup.State == ViewStateReady {
		// Clone lookup snapshot to prevent reference sharing
		snapshot := m.lookup
		if m.lookup.Explorer != nil {
			explorer := *m.lookup.Explorer
			snapshot.Explorer = &explorer
		}

		if len(m.lookupHistory) > 0 {
			last := m.lookupHistory[len(m.lookupHistory)-1]
			if last.Kind == snapshot.Kind && last.Query == snapshot.Query && (last.Explorer == nil) == (snapshot.Explorer == nil) {
				if last.Explorer == nil && snapshot.Explorer == nil {
					return
				}
				if last.Explorer != nil && snapshot.Explorer != nil && last.Explorer.Kind == snapshot.Explorer.Kind {
					return
				}
			}
		}

		m.lookupHistory = append(m.lookupHistory, snapshot)
		if len(m.lookupHistory) > 50 {
			m.lookupHistory = m.lookupHistory[1:]
		}
		m.lookupForward = nil
	}
}

// UpdateCacheSnapshot replaces the renderer-facing cache summary.
func (m *Model) UpdateCacheSnapshot(snapshot CacheSnapshot) {
	m.cache = snapshot
}

// ToggleHelpOverlay shows or hides the keyboard help overlay.
func (m *Model) ToggleHelpOverlay() {
	m.helpVisible = !m.helpVisible
	state := "shown"
	if !m.helpVisible {
		state = "hidden"
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Keyboard help %s.", state),
	}
}

// OpenCommandPalette opens the command/search overlay.
func (m *Model) OpenCommandPalette(prompt, seed string) {
	m.helpVisible = false
	m.command = CommandPaletteSnapshot{
		Visible:       true,
		Prompt:        prompt,
		Input:         seed,
		Results:       inferSearchResults(seed),
		SelectedIndex: 0,
		SearchLimit:   SearchResultLimit(),
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: "Search overlay opened.",
	}
}

// SetCommandPaletteSearchLimit adjusts the backend search page size for the active palette.
func (m *Model) SetCommandPaletteSearchLimit(limit int) {
	if !m.command.Visible || limit <= 0 {
		return
	}
	m.command.SearchLimit = limit
}

// SetCommandPaletteResults replaces ranked palette rows and clamps the selection index.
func (m *Model) SetCommandPaletteResults(results []SearchResult) {
	if !m.command.Visible {
		return
	}
	m.command.Results = results
	if m.command.SelectedIndex >= len(m.command.Results) {
		m.command.SelectedIndex = clampIndex(len(m.command.Results)-1, len(m.command.Results))
	}
}

// CloseCommandPalette hides the command/search overlay.
func (m *Model) CloseCommandPalette() {
	if !m.command.Visible {
		return
	}
	m.command = CommandPaletteSnapshot{}
	m.status = Status{
		Level:   StatusInfo,
		Message: "Search overlay closed.",
	}
}

// SetCommandPaletteInput replaces the overlay input buffer.
func (m *Model) SetCommandPaletteInput(value string) {
	if !m.command.Visible {
		return
	}
	m.command.Input = value
	m.refreshCommandPaletteResults()
}

// AppendCommandPaletteInput adds text to the overlay input buffer.
func (m *Model) AppendCommandPaletteInput(value string) {
	if !m.command.Visible || value == "" {
		return
	}
	m.command.Input += value
	m.refreshCommandPaletteResults()
}

// TrimCommandPaletteInput removes the last rune from the overlay input buffer.
func (m *Model) TrimCommandPaletteInput() {
	if !m.command.Visible || m.command.Input == "" {
		return
	}
	runes := []rune(m.command.Input)
	if len(runes) == 0 {
		m.command.Input = ""
		m.refreshCommandPaletteResults()
		return
	}
	m.command.Input = string(runes[:len(runes)-1])
	m.refreshCommandPaletteResults()
}

// MoveCommandPaletteSelection adjusts the selected inferred search result.
func (m *Model) MoveCommandPaletteSelection(delta int) {
	if !m.command.Visible || len(m.command.Results) == 0 || delta == 0 {
		return
	}
	m.command.SelectedIndex = clampIndex(m.command.SelectedIndex+delta, len(m.command.Results))
}

// SelectedCommandPaletteResult returns the highlighted inferred search result.
func (m *Model) SelectedCommandPaletteResult() *SearchResult {
	if !m.command.Visible || len(m.command.Results) == 0 {
		return nil
	}
	if m.command.SelectedIndex < 0 || m.command.SelectedIndex >= len(m.command.Results) {
		return nil
	}
	result := m.command.Results[m.command.SelectedIndex]
	return &result
}

// MergeCommandPaletteBackendResults appends backend search results to current palette results.
func (m *Model) MergeCommandPaletteBackendResults(results []backendclient.SearchResult) {
	if !m.command.Visible {
		return
	}

	merged := m.command.Results
	if len(merged) == 0 {
		merged = inferSearchResults(m.command.Input)
	}
	seen := make(map[string]struct{}, len(merged)+len(results))
	for _, result := range merged {
		key := searchResultDedupKey(result)
		if key != "" {
			seen[key] = struct{}{}
		}
	}

	for _, result := range results {
		command := strings.TrimSpace(result.Command)
		if command == "" {
			continue
		}
		key := command
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		source := strings.TrimSpace(result.Source)
		if source == "" {
			source = "indexer"
		}
		merged = append(merged, SearchResult{
			Kind:        strings.TrimSpace(result.Kind),
			Title:       strings.TrimSpace(result.Title),
			Description: strings.TrimSpace(result.Description),
			Command:     command,
			Enabled:     true,
			Source:      source,
		})
	}

	m.command.Results = RankSearchResults(m.command.Input, merged)
	if m.command.SelectedIndex >= len(m.command.Results) {
		m.command.SelectedIndex = clampIndex(len(m.command.Results)-1, len(m.command.Results))
	}
}

// MergeCommandPaletteLocalResults appends locally persisted metadata results.
func (m *Model) MergeCommandPaletteLocalResults(results []SearchResult) {
	if !m.command.Visible {
		return
	}

	merged := m.command.Results
	if len(merged) == 0 {
		merged = inferSearchResults(m.command.Input)
	}
	seen := make(map[string]struct{}, len(merged)+len(results))
	for _, result := range merged {
		key := searchResultDedupKey(result)
		seen[key] = struct{}{}
	}

	for _, result := range results {
		key := searchResultDedupKey(result)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, result)
	}

	m.command.Results = RankSearchResults(m.command.Input, merged)
	if m.command.SelectedIndex >= len(m.command.Results) {
		m.command.SelectedIndex = clampIndex(len(m.command.Results)-1, len(m.command.Results))
	}
}

// SetInfoStatus surfaces a non-error runtime update without changing screens.
func (m *Model) SetInfoStatus(message string) {
	m.status = Status{
		Level:   StatusInfo,
		Message: message,
	}
}

// SetWarningStatus surfaces a recoverable warning without changing screens.
func (m *Model) SetWarningStatus(message string) {
	m.status = Status{
		Level:   StatusWarn,
		Message: message,
	}
}

// SetErrorStatus surfaces a runtime failure without changing screens.
func (m *Model) SetErrorStatus(message string) {
	m.status = Status{
		Level:   StatusError,
		Message: message,
	}
}

func (m *Model) refreshCommandPaletteResults() {
	m.command.Results = inferSearchResults(m.command.Input)
	if m.command.SelectedIndex >= len(m.command.Results) {
		m.command.SelectedIndex = clampIndex(len(m.command.Results)-1, len(m.command.Results))
	}
}

// FocusNext rotates focus through the standard desktop shell areas.
func (m *Model) FocusNext() {
	m.FocusNextAvailable(FocusMain, FocusSidebar, FocusStatus)
}

// FocusNextAvailable rotates focus through the supplied available areas.
func (m *Model) FocusNextAvailable(areas ...FocusArea) {
	if len(areas) == 0 {
		areas = []FocusArea{FocusMain}
	}
	for index, area := range areas {
		if area != m.focus {
			continue
		}
		m.focus = areas[(index+1)%len(areas)]
		m.status = Status{
			Level:   StatusInfo,
			Message: fmt.Sprintf("Focus moved to %s.", m.focus),
		}
		return
	}
	m.focus = areas[0]
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Focus moved to %s.", m.focus),
	}
}

// MoveSelection adjusts the active row selection within the current screen.
func (m *Model) MoveSelection(delta int) {
	if m.current != ScreenLiveFeed || len(m.liveFeed.RecentTransactions) == 0 || delta == 0 {
		return
	}

	m.selection.LiveFeedIndex = clampIndex(m.selection.LiveFeedIndex+delta, len(m.liveFeed.RecentTransactions))
	selected := m.liveFeed.RecentTransactions[m.selection.LiveFeedIndex]
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Selected live transaction %d: %s.", m.selection.LiveFeedIndex+1, selected.Hash),
	}
}

// SelectedLiveTransaction returns the currently highlighted live transaction.
func (m *Model) SelectedLiveTransaction() *backendclient.TransactionSummary {
	if m.current != ScreenLiveFeed || len(m.liveFeed.RecentTransactions) == 0 {
		return nil
	}
	if m.selection.LiveFeedIndex < 0 || m.selection.LiveFeedIndex >= len(m.liveFeed.RecentTransactions) {
		return nil
	}

	tx := m.liveFeed.RecentTransactions[m.selection.LiveFeedIndex]
	return &tx
}

// HandleCommand applies one command to the app state and reports whether the
// caller should continue the interactive loop.
func (m *Model) HandleCommand(input string) bool {
	command := strings.TrimSpace(strings.ToLower(input))
	if command == "" {
		m.status = Status{
			Level:   StatusWarn,
			Message: "Empty command. Use help to inspect available actions.",
		}
		return true
	}

	fields := strings.Fields(command)
	if len(fields) > 0 && (fields[0] == "live" || fields[0] == "feed") {
		if len(fields) == 1 {
			_ = m.SetScreen(ScreenLiveFeed)
			return true
		}
		switch fields[1] {
		case "pause", "paused":
			m.SetLiveFeedPaused(true)
		case "resume", "start":
			m.SetLiveFeedPaused(false)
		case "filter":
			if len(fields) < 3 {
				m.status = Status{Level: StatusError, Message: "Usage: live filter all|soroban|classic|clear|account|contract|asset|operation <value>."}
				return true
			}
			if err := m.SetLiveFeedFilterCommand(fields[1:]); err != nil {
				m.status = Status{Level: StatusError, Message: err.Error()}
			}
		case "all":
			_ = m.SetLiveFeedFilter(LiveFeedFilterAll)
		case "soroban":
			_ = m.SetLiveFeedFilter(LiveFeedFilterSoroban)
		case "classic":
			_ = m.SetLiveFeedFilter(LiveFeedFilterClassic)
		case "return", "monitor", "monitoring":
			if !m.ReturnToLiveMonitoring() {
				m.status = Status{Level: StatusWarn, Message: "No live monitoring context to restore."}
			}
		default:
			m.status = Status{Level: StatusError, Message: "Usage: live pause|resume|return|filter all|soroban|classic."}
		}
		return true
	}

	switch command {
	case "home":
		_ = m.SetScreen(ScreenHome)
	case "live", "feed":
		_ = m.SetScreen(ScreenLiveFeed)
	case "lookup", "search":
		_ = m.SetScreen(ScreenLookup)
	case "settings", "config":
		_ = m.SetScreen(ScreenSettings)
	case "status":
		profile, _ := m.config.Profile(m.config.DefaultProfile)
		m.status = Status{
			Level:   StatusInfo,
			Message: fmt.Sprintf("Profile %q targets %s at %s.", profile.Name, profile.Network, profile.RPCEndpoint),
		}
	case "back":
		_ = m.Back()
	case "forward", "next":
		_ = m.Forward()
	case "help":
		m.ToggleHelpOverlay()
	case "quit", "exit":
		m.status = Status{
			Level:   StatusInfo,
			Message: "Shutting down tui.",
		}
		return false
	default:
		m.status = Status{
			Level:   StatusError,
			Message: fmt.Sprintf("Unknown command %q.", command),
		}
	}

	return true
}

func clampIndex(value, count int) int {
	if count <= 0 {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value >= count {
		return count - 1
	}
	return value
}

func profileHasDataBackend(profile config.Profile) bool {
	return strings.TrimSpace(profile.IndexerURL) != "" || strings.TrimSpace(profile.RPCEndpoint) != ""
}

func activeBackendLabel(profile config.Profile) string {
	mode := profile.NormalizedBackendMode()
	switch mode {
	case config.BackendModeRPC:
		return strings.TrimSpace(profile.RPCEndpoint)
	case config.BackendModeHybrid:
		if strings.TrimSpace(profile.IndexerURL) != "" {
			return strings.TrimSpace(profile.IndexerURL)
		}
		return strings.TrimSpace(profile.RPCEndpoint)
	default:
		if strings.TrimSpace(profile.IndexerURL) != "" {
			return strings.TrimSpace(profile.IndexerURL)
		}
		return strings.TrimSpace(profile.RPCEndpoint)
	}
}

// DefaultSourceMetadata declares the configured source policy for an operation.
func DefaultSourceMetadata(profile config.Profile, operation string) SourceMetadata {
	preferred := profile.PreferredSource()
	fallback := profile.FallbackSource()
	policy := sourcePolicy(preferred, fallback)

	label := strings.TrimSpace(profile.RPCEndpoint)
	if preferred == "indexer" && strings.TrimSpace(profile.IndexerURL) != "" {
		label = strings.TrimSpace(profile.IndexerURL)
	}

	return SourceMetadata{
		Mode:      profile.NormalizedBackendMode(),
		Operation: strings.TrimSpace(operation),
		Policy:    policy,
		Preferred: preferred,
		Actual:    preferred,
		Label:     label,
	}
}

func sourcePolicy(preferred string, fallback string) string {
	preferred = strings.TrimSpace(preferred)
	fallback = strings.TrimSpace(fallback)
	if preferred == "" {
		return "single-source"
	}
	if fallback == "" {
		return fmt.Sprintf("single-source: %s only; no field merge", preferred)
	}
	return fmt.Sprintf("single-source: %s -> %s; no field merge", preferred, fallback)
}

func valueOrFallbackSourceLabel(source SourceMetadata, fallback string) string {
	if strings.TrimSpace(source.Label) != "" {
		return strings.TrimSpace(source.Label)
	}
	return strings.TrimSpace(fallback)
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(fallback)
}

func firstSourceMetadata(values []SourceMetadata, profile config.Profile, operation string) SourceMetadata {
	if len(values) > 0 {
		return values[0]
	}
	return DefaultSourceMetadata(profile, operation)
}

func inferSearchResults(input string) []SearchResult {
	query := strings.TrimSpace(input)
	if query == "" {
		return nil
	}
	if looksLikeExplicitCommand(query) {
		return inferWorkspaceCompletions(query)
	}

	candidates := searchinfer.FromQuery(query)
	if len(candidates) == 0 {
		return nil
	}
	results := make([]SearchResult, 0, len(candidates))
	for _, candidate := range candidates {
		results = append(results, SearchResult{
			Kind:        candidate.Kind,
			Title:       candidate.Title,
			Description: candidate.Description,
			Command:     candidate.Command,
			Enabled:     true,
			Source:      "local",
		})
	}
	return results
}

func inferWorkspaceCompletions(query string) []SearchResult {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	if len(fields) == 0 || len(fields) > 2 {
		return nil
	}
	second := ""
	if len(fields) == 2 {
		second = fields[1]
	}
	switch fields[0] {
	case "bookmark":
		return filterSubcommandCompletions([]SearchResult{
			{Kind: "command", Title: "bookmark add [title]", Description: "bookmark current entity", Command: "bookmark add ", Enabled: true, Source: "local"},
			{Kind: "command", Title: "bookmark remove", Description: "remove bookmarks for current entity", Command: "bookmark remove", Enabled: true, Source: "local"},
			{Kind: "command", Title: "bookmark note <text>", Description: "annotate bookmark for current entity", Command: "bookmark note ", Enabled: true, Source: "local"},
		}, second)
	case "note":
		return filterSubcommandCompletions([]SearchResult{
			{Kind: "command", Title: "note add [title] [| body]", Description: "add a note for current entity", Command: "note add ", Enabled: true, Source: "local"},
			{Kind: "command", Title: "note remove [filter]", Description: "remove notes (optionally by title keyword)", Command: "note remove", Enabled: true, Source: "local"},
			{Kind: "command", Title: "note body [filter |] <text>", Description: "update note body (optional title filter before |)", Command: "note body ", Enabled: true, Source: "local"},
		}, second)
	case "label":
		return filterSubcommandCompletions([]SearchResult{
			{Kind: "command", Title: "label add <name>", Description: "apply a label to current entity", Command: "label add ", Enabled: true, Source: "local"},
			{Kind: "command", Title: "label remove <name>", Description: "detach a label from current entity", Command: "label remove ", Enabled: true, Source: "local"},
			{Kind: "command", Title: "label delete <name>", Description: "delete label definition entirely", Command: "label delete ", Enabled: true, Source: "local"},
			{Kind: "command", Title: "label color <name> <color>", Description: "set label color", Command: "label color ", Enabled: true, Source: "local"},
		}, second)
	case "open":
		return filterSubcommandCompletions([]SearchResult{
			{Kind: "command", Title: "open recent", Description: "browse recently visited entities", Command: "open recent", Enabled: true, Source: "local"},
			{Kind: "command", Title: "open bookmarks", Description: "browse saved bookmarks", Command: "open bookmarks", Enabled: true, Source: "local"},
			{Kind: "command", Title: "open notes", Description: "browse saved notes", Command: "open notes", Enabled: true, Source: "local"},
			{Kind: "command", Title: "open labels", Description: "browse labels and their entities", Command: "open labels", Enabled: true, Source: "local"},
			{Kind: "command", Title: "open views", Description: "browse saved investigation views", Command: "open views", Enabled: true, Source: "local"},
			{Kind: "command", Title: "open cache", Description: "reload cached payload for current entity", Command: "open cache", Enabled: true, Source: "local"},
		}, second)
	case "view":
		return filterSubcommandCompletions([]SearchResult{
			{Kind: "command", Title: "view save <name>", Description: "save current screen and filters", Command: "view save ", Enabled: true, Source: "local"},
			{Kind: "command", Title: "view open <name>", Description: "restore a saved view", Command: "view open ", Enabled: true, Source: "local"},
			{Kind: "command", Title: "view delete <name>", Description: "delete a saved view", Command: "view delete ", Enabled: true, Source: "local"},
		}, second)
	}
	return nil
}

func filterSubcommandCompletions(completions []SearchResult, prefix string) []SearchResult {
	if prefix == "" {
		return completions
	}
	var filtered []SearchResult
	for _, c := range completions {
		parts := strings.Fields(strings.TrimSpace(c.Command))
		if len(parts) >= 2 && strings.HasPrefix(parts[1], prefix) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// InferSearchResults exposes the local inference used by the command palette.
func InferSearchResults(input string) []SearchResult {
	return inferSearchResults(input)
}

// RankSearchResults orders command palette rows by how directly they match the query.
func RankSearchResults(query string, results []SearchResult) []SearchResult {
	ranked := append([]SearchResult(nil), results...)
	query = strings.ToLower(strings.TrimSpace(query))
	sort.SliceStable(ranked, func(i, j int) bool {
		left := searchResultScore(query, ranked[i])
		right := searchResultScore(query, ranked[j])
		if left == right {
			return i < j
		}
		return left > right
	})
	return ranked
}

func searchResultScore(query string, result SearchResult) int {
	score := 0
	if result.Enabled {
		score += 20
	} else {
		score -= 100
	}
	switch strings.ToLower(strings.TrimSpace(result.Source)) {
	case "indexer":
		score += 10
	case "horizon":
		score += 8
	case "local":
		score += 6
	}
	if query == "" {
		return score
	}

	kind := strings.ToLower(strings.TrimSpace(result.Kind))
	title := strings.ToLower(strings.TrimSpace(result.Title))
	description := strings.ToLower(strings.TrimSpace(result.Description))
	command := strings.ToLower(strings.TrimSpace(result.Command))
	target := commandTarget(command)

	if target == query {
		score += 140
	} else if strings.HasPrefix(target, query) {
		score += 90
	} else if strings.Contains(target, query) {
		score += 45
	}
	if isPartialHexQuery(query) && strings.Contains(command, query) {
		score += 60
	}
	if assetCode, assetIssuer, ok := parseAssetQuery(query); ok {
		if strings.Contains(command, assetCode+":"+assetIssuer) {
			score += 120
		} else if strings.Contains(command, assetCode+":") {
			score += 70
		}
	}
	if title == query || strings.TrimPrefix(title, kind+" ") == query {
		score += 100
	} else if strings.HasPrefix(title, query) || strings.HasPrefix(strings.TrimPrefix(title, kind+" "), query) {
		score += 70
	} else if strings.Contains(title, query) {
		score += 35
	}
	if kind == query {
		score += 25
	}
	if strings.Contains(description, query) {
		score += 15
	}
	return score
}

func isPartialHexQuery(query string) bool {
	if len(query) < 8 || len(query) >= 64 {
		return false
	}
	for _, r := range query {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func parseAssetQuery(query string) (code string, issuer string, ok bool) {
	code, issuer, found := strings.Cut(query, ":")
	if !found {
		return "", "", false
	}
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || issuer == "" {
		return "", "", false
	}
	return strings.ToLower(code), strings.ToLower(issuer), true
}

func commandTarget(command string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(command)))
	if len(fields) < 3 || fields[0] != "lookup" {
		return ""
	}
	return strings.Join(fields[2:], " ")
}

func searchResultDedupKey(result SearchResult) string {
	if command := strings.TrimSpace(result.Command); command != "" {
		return command
	}
	return strings.TrimSpace(result.Kind) + ":" + strings.TrimSpace(result.Title) + ":" + strings.TrimSpace(result.Description)
}

func looksLikeExplicitCommand(value string) bool {
	first, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(value)), " ")
	switch first {
	case "home", "live", "feed", "lookup", "search", "settings", "config", "status", "back", "forward", "next", "help", "quit", "exit", "open",
		"bookmark", "note", "label", "view", "watch":
		return true
	default:
		return false
	}
}

func isTransactionHash(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func isAccountAddress(value string) bool {
	_, err := strkey.Decode(strkey.VersionByteAccountID, value)
	return err == nil
}

func isContractAddress(value string) bool {
	_, err := strkey.Decode(strkey.VersionByteContract, value)
	return err == nil
}

func isLedgerSequence(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	sequence, err := strconv.ParseUint(value, 10, 32)
	return err == nil && sequence > 0
}

func truncateSearchValue(value string) string {
	if len(value) <= 16 {
		return value
	}
	return value[:8] + "..." + value[len(value)-6:]
}

// ParseScreen converts a user value into a typed screen constant.
func ParseScreen(value string) (Screen, error) {
	screen := Screen(strings.TrimSpace(strings.ToLower(value)))
	if _, ok := validScreens[screen]; !ok {
		return "", fmt.Errorf("unknown screen %q", value)
	}

	return screen, nil
}

// BuildCacheSnapshot converts local persistence state into a renderer-friendly shape.
func BuildCacheSnapshot(
	cfg config.Config,
	store *cache.Store,
	cachePath string,
	schemaVersion int,
	profiles []cache.Profile,
	lastScreen string,
	status string,
) CacheSnapshot {
	profile, _ := cfg.Profile(cfg.DefaultProfile)
	snapshot := CacheSnapshot{
		Enabled:      cfg.Cache.Driver != "",
		Available:    store != nil,
		Path:         cachePath,
		Schema:       schemaVersion,
		Profiles:     len(profiles),
		LastScreen:   lastScreen,
		Status:       status,
		DefaultID:    cfg.DefaultProfile,
		DefaultLabel: profile.Name,
	}

	if snapshot.DefaultLabel == "" {
		snapshot.DefaultLabel = cfg.DefaultProfile
	}

	return snapshot
}
