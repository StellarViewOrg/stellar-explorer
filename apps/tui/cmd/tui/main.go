package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/networkbackend"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/ui"
)

const stateKeySession = "session"

type sessionState struct {
	LastScreen string `json:"last_screen"`
}

var openCacheStore = func(ctx context.Context, driverName, dsn string) (*cache.Store, error) {
	return cache.OpenSQLite(ctx, driverName, dsn)
}

type lookupBackend interface {
	Label() string
	Search(ctx context.Context, query string, limit int) (backendclient.SearchResponse, error)
	LiveFeedSummary(ctx context.Context) (backendclient.LiveFeedSummaryResponse, error)
	Ledger(ctx context.Context, sequence uint32) (backendclient.LedgerLookupResponse, error)
	Transaction(ctx context.Context, hash string) (backendclient.TransactionLookupResponse, error)
	Account(ctx context.Context, id string) (backendclient.AccountLookupResponse, error)
	Asset(ctx context.Context, code string, issuer string) (backendclient.AssetLookupResponse, error)
	Contract(ctx context.Context, id string) (backendclient.ContractLookupResponse, error)
}

type explorerListBackend interface {
	Ledgers(ctx context.Context, limit int, before uint32) ([]backendclient.LedgerSummary, error)
	LedgerTransactions(ctx context.Context, sequence uint32, limit int, offset int) ([]backendclient.TransactionSummary, error)
	Accounts(ctx context.Context, limit int) ([]backendclient.AccountDetail, error)
	AccountOperations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error)
	AccountTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error)
	Assets(ctx context.Context, limit int) ([]backendclient.AssetDetail, error)
	AssetHolders(ctx context.Context, code string, issuer string, limit int, offset int) ([]backendclient.AssetHolderSummary, error)
	AssetTimeline(ctx context.Context, code string, issuer string, limit int, offset int, category string) ([]backendclient.TimelineItem, error)
	Contracts(ctx context.Context, limit int) ([]backendclient.ContractDetail, error)
	ContractEvents(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractEventSummary, error)
	ContractStorage(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractStorageSummary, error)
	ContractInvocations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error)
	ContractTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error)
}

type sourceMetadataProvider interface {
	SourceMetadata() app.SourceMetadata
}

var openLookupBackend = func(profile config.Profile) (lookupBackend, error) {
	mode := strings.ToLower(strings.TrimSpace(profile.BackendMode))
	switch mode {
	case config.BackendModeRPC:
		return networkbackend.New(profile)
	case config.BackendModeHybrid:
		network, networkErr := networkbackend.New(profile)
		if strings.TrimSpace(profile.IndexerURL) == "" {
			return network, networkErr
		}
		client, clientErr := backendclient.New(profile.IndexerURL)
		if clientErr != nil {
			return network, networkErr
		}
		if networkErr != nil {
			return client, nil
		}
		return newHybridLookupBackend(client, network), nil
	default:
		if strings.TrimSpace(profile.IndexerURL) != "" {
			return backendclient.New(profile.IndexerURL)
		}
		return networkbackend.New(profile)
	}
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin *os.File, stdout, stderr *os.File) int {
	flags := flag.NewFlagSet("tui", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var configPath string
	var labelsPath string
	var screenValue string
	var command string
	var noInteractive bool
	var showVersion bool

	flags.StringVar(&configPath, "config", "", "path to stellar-tui config file")
	flags.StringVar(&labelsPath, "labels", "", "path to labels.toml")
	flags.StringVar(&screenValue, "screen", "", "initial screen: home|live-feed|lookup|settings")
	flags.StringVar(&command, "command", "", "single command to apply before rendering")
	flags.BoolVar(&noInteractive, "no-interactive", false, "render once and exit")
	flags.BoolVar(&showVersion, "version", false, "print version and exit")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if showVersion {
		fmt.Fprintf(stdout, "stellar-tui %s\n", config.Version)
		return 0
	}

	cfg, resolvedPath, createdConfig, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}
	if createdConfig {
		fmt.Fprintf(stderr, "created config at %s\n", resolvedPath)
	}

	startup, err := initializeStartupState(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "initialize cache: %v\n", err)
		return 1
	}
	if startup.store != nil {
		defer func() {
			_ = startup.store.Close()
		}()
		if err := importLabelsFromFile(context.Background(), startup.store, cfg, labelsPath); err != nil {
			fmt.Fprintf(stderr, "import labels: %v\n", err)
			return 1
		}
	}

	model := app.NewModel(cfg, resolvedPath, startup.snapshot)
	if err := restoreLiveFeedScrollback(context.Background(), startup.store, cfg, model); err != nil {
		fmt.Fprintf(stderr, "restore live feed: %v\n", err)
		return 1
	}
	if startup.restoredScreen != "" {
		if screen, err := app.ParseScreen(startup.restoredScreen); err == nil {
			_ = model.RestoreScreen(screen)
		}
	}

	if screenValue != "" {
		screen, err := app.ParseScreen(screenValue)
		if err != nil {
			fmt.Fprintf(stderr, "parse screen: %v\n", err)
			return 1
		}
		if err := model.SetScreen(screen); err != nil {
			fmt.Fprintf(stderr, "set screen: %v\n", err)
			return 1
		}
	}

	refreshCurrentScreen(context.Background(), cfg, model, startup.store)
	if err := persistSessionState(startup.store, model.Snapshot().Current); err != nil {
		fmt.Fprintf(stderr, "persist session: %v\n", err)
		return 1
	}

	if command != "" {
		keepRunning, err := executeCommand(context.Background(), cfg, model, command, startup.store)
		if err != nil {
			fmt.Fprintf(stderr, "command failed: %v\n", err)
			return 1
		}
		if !keepRunning {
			if err := persistSessionState(startup.store, model.Snapshot().Current); err != nil {
				fmt.Fprintf(stderr, "persist session: %v\n", err)
				return 1
			}
			fmt.Fprintln(stdout, ui.Render(model.Snapshot()))
			return 0
		}
	}
	refreshCurrentScreen(context.Background(), cfg, model, startup.store)

	if noInteractive || !isInteractive(stdin) {
		if err := persistSessionState(startup.store, model.Snapshot().Current); err != nil {
			fmt.Fprintf(stderr, "persist session: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, ui.Render(model.Snapshot()))
		return 0
	}

	return runInteractive(cfg, model, startup.store, stdin, stdout, stderr)
}

func initializeLookupBackend(cfg config.Config) (lookupBackend, error) {
	profile, ok := cfg.Profile(cfg.DefaultProfile)
	if !ok {
		return nil, fmt.Errorf("default profile %q not found", cfg.DefaultProfile)
	}
	if strings.TrimSpace(profile.IndexerURL) == "" && strings.TrimSpace(profile.RPCEndpoint) == "" {
		return nil, nil
	}

	return openLookupBackend(profile)
}

func refreshCurrentScreen(ctx context.Context, cfg config.Config, model *app.Model, stores ...*cache.Store) {
	if model == nil {
		return
	}

	store := firstStore(stores)
	switch model.Snapshot().Current {
	case app.ScreenLookup:
		refreshActiveLookup(ctx, cfg, model, store)
	case app.ScreenLiveFeed:
		if model.LiveFeedPaused() {
			return
		}

		profile, ok := cfg.Profile(cfg.DefaultProfile)
		if !ok {
			return
		}
		backend, err := initializeLookupBackend(cfg)
		if err != nil {
			return
		}

		mergeTransactions := model.Snapshot().LiveFeed.SourceMode != app.LiveFeedSourceStream
		if mergeTransactions {
			_ = model.RefreshLiveFeed(ctx, backend, sourceMetadataFor(profile, "live-feed", backend))
		} else {
			_ = model.RefreshLiveFeedMetadata(ctx, backend, sourceMetadataFor(profile, "live-feed", backend))
		}
		if store != nil {
			_ = persistLiveFeedScrollback(ctx, store, profile.Name, model.Snapshot().LiveFeed.Scrollback)
		}
	}
}

func executeCommand(ctx context.Context, cfg config.Config, model *app.Model, input string, stores ...*cache.Store) (bool, error) {
	if isLookupCommand(input) {
		backend, err := initializeLookupBackend(cfg)
		if err != nil {
			return false, err
		}
		if backend == nil {
			profile, _ := cfg.Profile(cfg.DefaultProfile)
			model.SetLookupError("", "", errors.New("indexer_url is not configured for the current profile"), app.DefaultSourceMetadata(profile, "lookup"))
			return true, nil
		}
		return executeLookupCommand(ctx, model, backend, input, firstStore(stores))
	}

	if isWorkspaceCommand(input) {
		return executeWorkspaceCommand(ctx, model, firstStore(stores), input)
	}

	if isViewCommand(input) {
		return executeViewCommand(ctx, cfg, model, firstStore(stores), input)
	}

	if isWatchCommand(input) {
		return executeWatchCommand(ctx, cfg, model, firstStore(stores), input)
	}

	if isSearchMoreCommand(input) {
		return executeSearchMoreCommand(ctx, cfg, model, firstStore(stores), input)
	}

	if isOpenCommand(input) {
		fields := strings.Fields(strings.TrimSpace(input))
		if len(fields) >= 2 && strings.EqualFold(fields[1], "cache") {
			return executeOpenCacheCommand(ctx, model, firstStore(stores), fields[2:])
		}
		backend, err := initializeLookupBackend(cfg)
		if err != nil {
			return false, err
		}
		if backend == nil {
			profile, _ := cfg.Profile(cfg.DefaultProfile)
			model.SetLookupError("", "", errors.New("indexer_url is not configured for the current profile"), app.DefaultSourceMetadata(profile, "open"))
			return true, nil
		}
		return executeOpenCommand(ctx, model, backend, input)
	}

	keepRunning := model.HandleCommand(input)
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(input)))
	if keepRunning && len(fields) > 0 && (fields[0] == "live" || fields[0] == "feed") {
		profile, ok := cfg.Profile(cfg.DefaultProfile)
		if ok {
			applyProfileWatchOnLiveScreen(ctx, model, firstStore(stores), profile)
		}
	}
	return keepRunning, nil
}

func executeLookupCommand(ctx context.Context, model *app.Model, backend lookupBackend, input string, store *cache.Store) (bool, error) {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) == 0 {
		return false, nil
	}

	if len(fields) < 3 {
		model.SetLookupError("", "", errors.New("usage: lookup ledger|tx|account|asset|contract <id>"), sourceMetadataFor(model.Snapshot().Profile, "lookup", backend))
		return true, nil
	}

	kind := normalizeLookupKind(fields[1])
	query := strings.Join(fields[2:], " ")

	opts := lookupCacheOptions{}

	switch kind {
	case app.LookupLedger:
		sequence, err := strconvParseUint32(query)
		if err != nil {
			model.SetLookupError(kind, query, errors.New("ledger sequence must be a positive integer"), sourceMetadataFor(model.Snapshot().Profile, string(kind), backend))
			return true, nil
		}
		return performLedgerLookup(ctx, model, store, backend, query, sequence, opts)
	case app.LookupTransaction:
		return performTransactionLookup(ctx, model, store, backend, query, opts)
	case app.LookupOperation:
		txHash, orderText, ok := strings.Cut(strings.TrimSpace(query), ":")
		txHash = strings.TrimSpace(txHash)
		orderText = strings.TrimSpace(orderText)
		if !ok || txHash == "" || orderText == "" {
			model.SetLookupError(kind, query, errors.New("operation lookup requires TXHASH:INDEX"), sourceMetadataFor(model.Snapshot().Profile, string(kind), backend))
			return true, nil
		}
		index, err := strconv.Atoi(orderText)
		if err != nil || index < 1 {
			model.SetLookupError(kind, query, errors.New("operation index must be a positive integer"), sourceMetadataFor(model.Snapshot().Profile, string(kind), backend))
			return true, nil
		}
		if err := openOperationDetail(ctx, model, backend, index, txHash); err != nil {
			model.SetLookupError(kind, query, err, sourceMetadataFor(model.Snapshot().Profile, string(kind), backend))
			return true, nil
		}
		return true, nil
	case app.LookupAccount:
		return performAccountLookup(ctx, model, store, backend, query, opts)
	case app.LookupAsset:
		code, issuer, ok := strings.Cut(strings.TrimSpace(query), ":")
		code = strings.TrimSpace(code)
		issuer = strings.TrimSpace(issuer)
		if !ok || code == "" || issuer == "" {
			model.SetLookupError(kind, query, errors.New("asset lookup requires CODE:ISSUER"), sourceMetadataFor(model.Snapshot().Profile, string(kind), backend))
			return true, nil
		}
		target := code + ":" + issuer
		return performAssetLookup(ctx, model, store, backend, query, target, code, issuer, opts)
	case app.LookupContract:
		return performContractLookup(ctx, model, store, backend, query, opts)
	default:
		model.SetLookupError(app.LookupKind(fields[1]), query, errors.New("unknown lookup type"), sourceMetadataFor(model.Snapshot().Profile, string(kind), backend))
		return true, nil
	}
}

func firstStore(stores []*cache.Store) *cache.Store {
	if len(stores) == 0 {
		return nil
	}
	return stores[0]
}

func afterLookupLoaded(ctx context.Context, model *app.Model, store *cache.Store, kind string, target string, payload any) {
	if model == nil || store == nil {
		return
	}
	profile := model.Snapshot().Profile
	if err := persistVisitedEntity(ctx, store, profile.Name, kind, target, payload, model.Snapshot().Lookup.Source.Label); err != nil {
		model.SetWarningStatus(err.Error())
		return
	}
	metadata, err := loadLookupMetadata(ctx, store, profile.Name, kind, target)
	if err != nil {
		model.SetWarningStatus(err.Error())
		return
	}
	model.SetLookupMetadata(metadata)
}

func persistVisitedEntity(ctx context.Context, store *cache.Store, profileID string, kind string, target string, payload any, sourceLabel string) error {
	if store == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cache entity payload: %w", err)
	}
	entity := cache.EntityCache{
		ProfileID:   profileID,
		Kind:        strings.TrimSpace(kind),
		Target:      strings.TrimSpace(target),
		Title:       entityTitle(kind, target),
		Summary:     entitySummary(kind, payload),
		Payload:     string(data),
		SourceLabel: sourceLabel,
	}
	if err := store.UpsertEntityCache(ctx, entity); err != nil {
		return fmt.Errorf("save visited entity: %w", err)
	}
	return nil
}

func loadLookupMetadata(ctx context.Context, store *cache.Store, profileID string, kind string, target string) (app.LookupMetadataSnapshot, error) {
	var metadata app.LookupMetadataSnapshot
	if store == nil {
		return metadata, nil
	}

	labels, err := store.ListLabels(ctx)
	if err != nil {
		return metadata, fmt.Errorf("load labels: %w", err)
	}
	labelByID := make(map[string]cache.Label, len(labels))
	for _, label := range labels {
		if label.ProfileID == profileID {
			labelByID[label.ID] = label
		}
	}
	labelTargets, err := store.ListLabelTargets(ctx)
	if err != nil {
		return metadata, fmt.Errorf("load label targets: %w", err)
	}
	for _, labelTarget := range labelTargets {
		if labelTarget.ProfileID != profileID || !lookupTargetMatches(labelTarget.Kind, labelTarget.Target, kind, target) {
			continue
		}
		if label, ok := labelByID[labelTarget.LabelID]; ok {
			metadata.Labels = append(metadata.Labels, app.LookupLabelSnapshot{Name: label.Name, Color: label.Color})
		}
	}

	bookmarks, err := store.ListBookmarks(ctx)
	if err != nil {
		return metadata, fmt.Errorf("load bookmarks: %w", err)
	}
	for _, bookmark := range bookmarks {
		if bookmark.ProfileID == profileID && lookupTargetMatches(bookmark.Kind, bookmark.Target, kind, target) {
			metadata.Bookmarks = append(metadata.Bookmarks, app.LookupBookmarkSnapshot{Title: bookmark.Title, Notes: bookmark.Notes})
		}
	}

	notes, err := store.ListNotes(ctx)
	if err != nil {
		return metadata, fmt.Errorf("load notes: %w", err)
	}
	for _, note := range notes {
		if note.ProfileID == profileID && noteTargetMatches(note.Target, kind, target) {
			metadata.Notes = append(metadata.Notes, app.LookupNoteSnapshot{Title: note.Title, Body: note.Body})
		}
	}

	entities, err := store.ListEntityCache(ctx, profileID, 100)
	if err != nil {
		return metadata, fmt.Errorf("load entity cache: %w", err)
	}
	for _, entity := range entities {
		if lookupTargetMatches(entity.Kind, entity.Target, kind, target) {
			metadata.Cached = &app.LookupCacheSnapshot{
				UpdatedAt:   entity.UpdatedAt,
				SourceLabel: entity.SourceLabel,
				Summary:     entity.Summary,
			}
			break
		}
	}

	return metadata, nil
}

func lookupTargetMatches(candidateKind string, candidateTarget string, kind string, target string) bool {
	return strings.EqualFold(strings.TrimSpace(candidateKind), strings.TrimSpace(kind)) &&
		strings.EqualFold(strings.TrimSpace(candidateTarget), strings.TrimSpace(target))
}

func noteTargetMatches(candidate string, kind string, target string) bool {
	candidate = strings.TrimSpace(candidate)
	return strings.EqualFold(candidate, strings.TrimSpace(target)) ||
		strings.EqualFold(candidate, strings.TrimSpace(kind)+":"+strings.TrimSpace(target))
}

func entityTitle(kind string, target string) string {
	return strings.TrimSpace(kind) + " " + strings.TrimSpace(target)
}

func entitySummary(kind string, payload any) string {
	switch value := payload.(type) {
	case backendclient.LedgerLookupResponse:
		if value.Ledger != nil {
			return fmt.Sprintf("%d tx / %d ops", value.Ledger.TransactionCount, value.Ledger.OperationCount)
		}
	case backendclient.TransactionLookupResponse:
		if value.Transaction != nil {
			return fmt.Sprintf("ledger %d / %d ops", value.Transaction.LedgerSequence, value.Transaction.OperationCount)
		}
	case backendclient.AccountLookupResponse:
		if value.Account != nil {
			return "balance " + value.Account.Balance
		}
	case backendclient.AssetLookupResponse:
		if value.Asset != nil {
			return fmt.Sprintf("%d accounts / supply %s", value.Asset.NumAccounts, value.Asset.TotalSupply)
		}
	case backendclient.ContractLookupResponse:
		if value.Contract != nil {
			return fmt.Sprintf("%d invocations / %d events", value.Contract.InvocationCount, value.Contract.EventCount)
		}
	}
	return strings.TrimSpace(kind)
}

func isLookupCommand(input string) bool {
	fields := strings.Fields(strings.TrimSpace(input))
	return len(fields) > 0 && strings.EqualFold(fields[0], "lookup")
}

func isOpenCommand(input string) bool {
	fields := strings.Fields(strings.TrimSpace(input))
	return len(fields) > 0 && strings.EqualFold(fields[0], "open")
}

func executeOpenCommand(ctx context.Context, model *app.Model, backend lookupBackend, input string) (bool, error) {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) < 2 {
		model.SetLookupError("", "", errors.New("usage: open tx <n> | open account"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
		return true, nil
	}

	switch strings.ToLower(fields[1]) {
	case "decode":
		if len(fields) < 3 {
			model.SetLookupError("", "", errors.New("usage: open decode raw|decoded"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		mode, err := parseContractDecodeMode(fields[2])
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if err := model.SetContractDecodeMode(mode); err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		return true, nil
	case "ledgers":
		listBackend, ok := backend.(explorerListBackend)
		if !ok {
			model.SetLookupError("", "ledgers", errors.New("ledger list is unavailable for the current backend"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		limit, before, err := parseLedgerListArgs(fields[2:])
		if err != nil {
			model.SetLookupError("", "ledgers", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		ledgers, err := listBackend.Ledgers(ctx, limit, before)
		if err != nil {
			model.SetLookupError("", "ledgers", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupResultExplorer(ledgerExplorerTitle(before), "", "", ledgerResults(ledgers, limit), limit, int(before), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
		return true, nil
	case "accounts":
		listBackend, ok := backend.(explorerListBackend)
		if !ok {
			model.SetLookupError("", "accounts", errors.New("account list is unavailable for the current backend"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		accounts, err := listBackend.Accounts(ctx, 20)
		if err != nil {
			model.SetLookupError("", "accounts", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupResultExplorer("Recent Accounts", "", "", accountResults(accounts), 20, 0, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
		return true, nil
	case "assets":
		listBackend, ok := backend.(explorerListBackend)
		if !ok {
			model.SetLookupError("", "assets", errors.New("asset list is unavailable for the current backend"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		assets, err := listBackend.Assets(ctx, 20)
		if err != nil {
			model.SetLookupError("", "assets", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupResultExplorer("Recent Assets", "", "", assetResults(assets), 20, 0, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
		return true, nil
	case "contracts":
		listBackend, ok := backend.(explorerListBackend)
		if !ok {
			model.SetLookupError("", "contracts", errors.New("contract list is unavailable for the current backend"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		contracts, err := listBackend.Contracts(ctx, 20)
		if err != nil {
			model.SetLookupError("", "contracts", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupResultExplorer("Recent Contracts", "", "", contractResults(contracts), 20, 0, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
		return true, nil
	case "tx", "transaction":
		if len(fields) < 3 {
			model.SetLookupError("", "", errors.New("usage: open tx <n>"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		index, err := strconv.Atoi(fields[2])
		if err != nil || index < 1 {
			model.SetLookupError("", "", errors.New("transaction index must be a positive integer"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}

		snapshot := model.Snapshot()
		transactions, scopeLabel := openTransactionScope(snapshot)
		if len(transactions) == 0 {
			model.SetLookupError("", "", errors.New("open tx <n> is only available from the live feed or a lookup with related transactions"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if index > len(transactions) {
			model.SetLookupError("", "", fmt.Errorf("%s only has %d transaction(s)", scopeLabel, len(transactions)), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}

		target := transactions[index-1].Hash
		response, err := backend.Transaction(ctx, target)
		if err != nil {
			model.SetLookupError(app.LookupTransaction, target, err, sourceMetadataFor(model.Snapshot().Profile, string(app.LookupTransaction), backend))
			return true, nil
		}
		model.UpdateLookupTransaction(target, response, sourceMetadataFor(model.Snapshot().Profile, string(app.LookupTransaction), backend))
		return true, nil
	case "txs", "transactions":
		snapshot := model.Snapshot()
		limit, offset, err := parseListPageArgs(fields[2:], 10)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		transactions, title, backCommand, nextCommand, err := openTransactionExplorerScope(ctx, backend, snapshot, limit, offset)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if len(transactions) == 0 {
			model.SetLookupError("", "", errors.New("open txs is only available from ledger, account, asset, or contract lookup results"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupTransactionExplorer(title, backCommand, transactions, limit, offset, nextCommand)
		return true, nil
	case "ops", "operations":
		snapshot := model.Snapshot()
		limit, offset, err := parseListPageArgs(fields[2:], 10)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		operations, title, backCommand, nextCommand, err := openOperationExplorerScope(ctx, backend, snapshot, limit, offset)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if len(operations) == 0 {
			model.SetLookupError("", "", errors.New("open ops is only available from transaction or account lookup results with operations"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupOperationExplorer(title, backCommand, operations, limit, offset, nextCommand)
		return true, nil
	case "holders":
		snapshot := model.Snapshot()
		limit, offset, err := parseListPageArgs(fields[2:], 10)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		holders, title, backCommand, nextCommand, err := openHolderExplorerScope(ctx, backend, snapshot, limit, offset)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if len(holders) == 0 {
			model.SetLookupError("", "", errors.New("open holders is only available from asset lookup results with holders"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupHolderExplorer(title, backCommand, holders, limit, offset, nextCommand)
		return true, nil
	case "events":
		snapshot := model.Snapshot()
		limit, offset, err := parseListPageArgs(fields[2:], 10)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		events, title, backCommand, nextCommand, err := openEventExplorerScope(ctx, backend, snapshot, limit, offset)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if len(events) == 0 {
			model.SetLookupError("", "", errors.New("open events is only available from contract lookup results with events"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupEventExplorer(title, backCommand, events, limit, offset, nextCommand)
		return true, nil
	case "event":
		if len(fields) < 3 {
			model.SetLookupError("", "", errors.New("usage: open event <n>"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		index, err := strconv.Atoi(fields[2])
		if err != nil || index < 1 {
			model.SetLookupError("", "", errors.New("event index must be a positive integer"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if err := openEventDetail(ctx, model, backend, index); err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		return true, nil
	case "storage":
		if len(fields) >= 3 {
			if index, err := strconv.Atoi(fields[2]); err == nil && index >= 1 {
				if err := openStorageEntryDetail(ctx, model, backend, index); err != nil {
					model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
					return true, nil
				}
				return true, nil
			}
		}
		snapshot := model.Snapshot()
		limit, offset, err := parseListPageArgs(fields[2:], 10)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		entries, title, backCommand, nextCommand, err := openStorageExplorerScope(ctx, backend, snapshot, limit, offset)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if len(entries) == 0 {
			model.SetLookupError("", "", errors.New("open storage is only available from contract lookup results with storage entries"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupStorageExplorer(title, backCommand, entries, limit, offset, nextCommand)
		return true, nil
	case "invocation":
		if len(fields) < 3 {
			model.SetLookupError("", "", errors.New("usage: open invocation <n>"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		index, err := strconv.Atoi(fields[2])
		if err != nil || index < 1 {
			model.SetLookupError("", "", errors.New("invocation index must be a positive integer"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if err := openInvocationDetail(ctx, model, backend, index); err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		return true, nil
	case "invocations":
		snapshot := model.Snapshot()
		limit, offset, err := parseListPageArgs(fields[2:], 10)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		operations, title, backCommand, nextCommand, err := openInvocationExplorerScope(ctx, backend, snapshot, limit, offset)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if len(operations) == 0 {
			model.SetLookupError("", "", errors.New("open invocations is only available from contract lookup results with invocations"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupInvocationExplorer(title, backCommand, operations, limit, offset, nextCommand)
		return true, nil
	case "timeline":
		snapshot := model.Snapshot()
		limit, offset, category, err := parseTimelineArgs(fields[2:], 20)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		items, title, backCommand, nextCommand, err := openTimelineExplorerScope(ctx, backend, snapshot, limit, offset, category)
		if err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if len(items) == 0 {
			model.SetLookupError("", "", errors.New("open timeline is only available from account, asset, or contract lookup results with activity"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.OpenLookupTimelineExplorer(title, backCommand, items, limit, offset, nextCommand)
		return true, nil
	case "op", "operation":
		if len(fields) < 3 {
			model.SetLookupError("", "", errors.New("usage: open op <n>"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		index, err := strconv.Atoi(fields[2])
		if err != nil || index < 1 {
			model.SetLookupError("", "", errors.New("operation index must be a positive integer"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		if err := openOperationDetail(ctx, model, backend, index, ""); err != nil {
			model.SetLookupError("", "", err, sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		return true, nil
	case "detail":
		snapshot := model.Snapshot()
		if snapshot.Current != app.ScreenLookup || snapshot.Lookup.Explorer == nil {
			model.SetLookupError("", "", errors.New("open detail is only available from an explorer list"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}
		model.CloseLookupExplorer()
		return true, nil
	case "account":
		snapshot := model.Snapshot()
		if snapshot.Lookup.Transaction == nil || snapshot.Lookup.Transaction.Transaction == nil {
			model.SetLookupError("", "", errors.New("open account is only available from a transaction lookup"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}

		target := strings.TrimSpace(snapshot.Lookup.Transaction.Transaction.Account)
		if target == "" {
			model.SetLookupError("", "", errors.New("current transaction has no source account"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
			return true, nil
		}

		response, err := backend.Account(ctx, target)
		if err != nil {
			model.SetLookupError(app.LookupAccount, target, err, sourceMetadataFor(model.Snapshot().Profile, string(app.LookupAccount), backend))
			return true, nil
		}
		model.UpdateLookupAccount(target, response, sourceMetadataFor(model.Snapshot().Profile, string(app.LookupAccount), backend))
		return true, nil
	default:
		model.SetLookupError("", "", errors.New("usage: open ledgers|accounts|assets|contracts|txs|ops|op <n>|holders|events|timeline|decode|detail|account|tx <n>"), sourceMetadataFor(model.Snapshot().Profile, "open", backend))
		return true, nil
	}
}

func parseContractDecodeMode(value string) (app.ContractDecodeMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "decoded", "decode":
		return app.ContractDecodeModeDecoded, nil
	case "raw":
		return app.ContractDecodeModeRaw, nil
	default:
		return "", errors.New("decode mode must be raw or decoded")
	}
}

func parseLedgerListArgs(args []string) (int, uint32, error) {
	limit := 20
	var before uint32

	for index := 0; index < len(args); index++ {
		key := strings.ToLower(strings.TrimSpace(args[index]))
		switch key {
		case "limit", "--limit":
			index++
			if index >= len(args) {
				return 0, 0, errors.New("open ledgers limit requires a value")
			}
			parsed, err := strconv.Atoi(args[index])
			if err != nil || parsed <= 0 {
				return 0, 0, errors.New("open ledgers limit must be a positive integer")
			}
			if parsed > 50 {
				parsed = 50
			}
			limit = parsed
		case "before", "--before":
			index++
			if index >= len(args) {
				return 0, 0, errors.New("open ledgers before requires a sequence")
			}
			parsed, err := strconv.ParseUint(args[index], 10, 32)
			if err != nil || parsed == 0 {
				return 0, 0, errors.New("open ledgers before must be a positive ledger sequence")
			}
			before = uint32(parsed)
		default:
			return 0, 0, fmt.Errorf("unknown open ledgers option %q", args[index])
		}
	}

	return limit, before, nil
}

func ledgerExplorerTitle(before uint32) string {
	if before == 0 {
		return "Recent Ledgers"
	}
	return fmt.Sprintf("Ledgers Before %d", before)
}

func ledgerResults(ledgers []backendclient.LedgerSummary, limit int) []app.SearchResult {
	results := make([]app.SearchResult, 0, len(ledgers)+1)
	for _, ledger := range ledgers {
		sequence := strconv.FormatUint(uint64(ledger.Sequence), 10)
		results = append(results, app.SearchResult{
			Kind:        "ledger",
			Title:       "Ledger " + sequence,
			Description: fmt.Sprintf("%d tx / %d ops", ledger.TransactionCount, ledger.OperationCount),
			Command:     "lookup ledger " + sequence,
			Enabled:     true,
			Source:      "indexer",
		})
	}
	if len(ledgers) >= limit && len(ledgers) > 0 {
		last := ledgers[len(ledgers)-1]
		results = append(results, app.SearchResult{
			Kind:        "page",
			Title:       "Next Page",
			Description: fmt.Sprintf("before ledger %d", last.Sequence),
			Command:     fmt.Sprintf("open ledgers limit %d before %d", limit, last.Sequence),
			Enabled:     true,
			Source:      "local",
		})
	}
	return results
}

func accountResults(accounts []backendclient.AccountDetail) []app.SearchResult {
	results := make([]app.SearchResult, 0, len(accounts))
	for _, account := range accounts {
		results = append(results, app.SearchResult{
			Kind:        "account",
			Title:       "Account " + truncateCommandLabel(account.ID, 16),
			Description: fmt.Sprintf("%s balance %s", truncateCommandLabel(account.ID, 18), account.Balance),
			Command:     "lookup account " + account.ID,
			Enabled:     true,
			Source:      "indexer",
		})
	}
	return results
}

func assetResults(assets []backendclient.AssetDetail) []app.SearchResult {
	results := make([]app.SearchResult, 0, len(assets))
	for _, asset := range assets {
		id := strings.TrimSpace(asset.AssetCode + ":" + asset.AssetIssuer)
		results = append(results, app.SearchResult{
			Kind:        "asset",
			Title:       "Asset " + asset.AssetCode,
			Description: fmt.Sprintf("%s accounts %d supply %s", truncateCommandLabel(id, 24), asset.NumAccounts, asset.TotalSupply),
			Command:     "lookup asset " + id,
			Enabled:     true,
			Source:      "indexer",
		})
	}
	return results
}

func contractResults(contracts []backendclient.ContractDetail) []app.SearchResult {
	results := make([]app.SearchResult, 0, len(contracts))
	for _, contract := range contracts {
		results = append(results, app.SearchResult{
			Kind:        "contract",
			Title:       "Contract " + truncateCommandLabel(contract.ContractID, 16),
			Description: fmt.Sprintf("%d invocations / %d events", contract.InvocationCount, contract.EventCount),
			Command:     "lookup contract " + contract.ContractID,
			Enabled:     true,
			Source:      "indexer",
		})
	}
	return results
}

func truncateCommandLabel(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func openTransactionScope(snapshot app.Snapshot) ([]backendclient.TransactionSummary, string) {
	if snapshot.Current == app.ScreenLookup && snapshot.Lookup.Explorer != nil && snapshot.Lookup.Explorer.Kind == app.LookupExplorerTransactions {
		label := strings.TrimSpace(snapshot.Lookup.Explorer.Title)
		if label == "" {
			label = "current explorer"
		}
		return snapshot.Lookup.Explorer.Transactions, label
	}
	switch snapshot.Current {
	case app.ScreenLiveFeed:
		return snapshot.LiveFeed.RecentTransactions, "live feed"
	case app.ScreenLookup:
		switch snapshot.Lookup.Kind {
		case app.LookupLedger:
			if snapshot.Lookup.Ledger != nil {
				return snapshot.Lookup.Ledger.Transactions, "current ledger lookup"
			}
		case app.LookupAccount:
			if snapshot.Lookup.Account != nil {
				return snapshot.Lookup.Account.RecentTransactions, "current account lookup"
			}
		case app.LookupAsset:
			if snapshot.Lookup.Asset != nil {
				return snapshot.Lookup.Asset.RecentTransactions, "current asset lookup"
			}
		case app.LookupContract:
			if snapshot.Lookup.Contract != nil {
				return snapshot.Lookup.Contract.RecentTransactions, "current contract lookup"
			}
		}
	}
	return nil, ""
}

func parseListPageArgs(args []string, defaultLimit int) (int, int, error) {
	limit := defaultLimit
	offset := 0
	for index := 0; index < len(args); index++ {
		key := strings.ToLower(strings.TrimSpace(args[index]))
		switch key {
		case "limit", "--limit":
			index++
			if index >= len(args) {
				return 0, 0, errors.New("list limit requires a value")
			}
			parsed, err := strconv.Atoi(args[index])
			if err != nil || parsed <= 0 {
				return 0, 0, errors.New("list limit must be a positive integer")
			}
			if parsed > 50 {
				parsed = 50
			}
			limit = parsed
		case "offset", "--offset":
			index++
			if index >= len(args) {
				return 0, 0, errors.New("list offset requires a value")
			}
			parsed, err := strconv.Atoi(args[index])
			if err != nil || parsed < 0 {
				return 0, 0, errors.New("list offset must be a non-negative integer")
			}
			offset = parsed
		default:
			return 0, 0, fmt.Errorf("unknown list option %q", args[index])
		}
	}
	return limit, offset, nil
}

func parseTimelineArgs(args []string, defaultLimit int) (int, int, string, error) {
	limit := defaultLimit
	offset := 0
	category := ""
	for index := 0; index < len(args); index++ {
		key := strings.ToLower(strings.TrimSpace(args[index]))
		switch key {
		case "limit", "--limit":
			index++
			if index >= len(args) {
				return 0, 0, "", errors.New("timeline limit requires a value")
			}
			parsed, err := strconv.Atoi(args[index])
			if err != nil || parsed <= 0 {
				return 0, 0, "", errors.New("timeline limit must be a positive integer")
			}
			if parsed > 50 {
				parsed = 50
			}
			limit = parsed
		case "offset", "--offset":
			index++
			if index >= len(args) {
				return 0, 0, "", errors.New("timeline offset requires a value")
			}
			parsed, err := strconv.Atoi(args[index])
			if err != nil || parsed < 0 {
				return 0, 0, "", errors.New("timeline offset must be a non-negative integer")
			}
			offset = parsed
		case "type", "--type", "kind", "--kind", "category", "--category":
			index++
			if index >= len(args) {
				return 0, 0, "", errors.New("timeline type requires a value")
			}
			normalized, ok := normalizeTimelineCategory(args[index])
			if !ok {
				return 0, 0, "", fmt.Errorf("unknown timeline type %q", args[index])
			}
			category = normalized
		default:
			normalized, ok := normalizeTimelineCategory(key)
			if !ok {
				return 0, 0, "", fmt.Errorf("unknown timeline option %q", args[index])
			}
			category = normalized
		}
	}
	return limit, offset, category, nil
}

func normalizeTimelineCategory(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "all", "activity":
		return "", true
	case "tx", "transaction", "transactions":
		return "tx", true
	case "op", "ops", "operation", "operations":
		return "op", true
	case "holder", "holders":
		return "holder", true
	case "event", "events":
		return "event", true
	default:
		return "", false
	}
}

func nextPageCommand(base string, limit int, offset int, count int) string {
	if count < limit {
		return ""
	}
	return fmt.Sprintf("%s limit %d offset %d", base, limit, offset+count)
}

func timelineNextPageCommand(limit int, offset int, count int, category string) string {
	base := "open timeline"
	if strings.TrimSpace(category) != "" {
		base += " type " + strings.TrimSpace(category)
	}
	return nextPageCommand(base, limit, offset, count)
}

func openTransactionExplorerScope(ctx context.Context, backend lookupBackend, snapshot app.Snapshot, limit int, offset int) ([]backendclient.TransactionSummary, string, string, string, error) {
	if snapshot.Current != app.ScreenLookup {
		return nil, "", "", "", nil
	}
	switch snapshot.Lookup.Kind {
	case app.LookupLedger:
		if snapshot.Lookup.Ledger == nil || snapshot.Lookup.Ledger.Ledger == nil {
			return nil, "", "", "", nil
		}
		sequence := snapshot.Lookup.Ledger.Ledger.Sequence
		if listBackend, ok := backend.(explorerListBackend); ok {
			transactions, err := listBackend.LedgerTransactions(ctx, sequence, limit, offset)
			if err != nil {
				return nil, "", "", "", err
			}
			return transactions, "Ledger Transactions", "open detail", nextPageCommand("open txs", limit, offset, len(transactions)), nil
		}
		return pagedSlice(snapshot.Lookup.Ledger.Transactions, limit, offset), "Ledger Transactions", "open detail", "", nil
	case app.LookupAccount:
		if snapshot.Lookup.Account == nil || len(snapshot.Lookup.Account.RecentTransactions) == 0 {
			return nil, "", "", "", nil
		}
		return pagedSlice(snapshot.Lookup.Account.RecentTransactions, limit, offset), "Account Transactions", "open detail", "", nil
	case app.LookupAsset:
		if snapshot.Lookup.Asset == nil || len(snapshot.Lookup.Asset.RecentTransactions) == 0 {
			return nil, "", "", "", nil
		}
		return pagedSlice(snapshot.Lookup.Asset.RecentTransactions, limit, offset), "Asset Transactions", "open detail", "", nil
	case app.LookupContract:
		if snapshot.Lookup.Contract == nil || len(snapshot.Lookup.Contract.RecentTransactions) == 0 {
			return nil, "", "", "", nil
		}
		return pagedSlice(snapshot.Lookup.Contract.RecentTransactions, limit, offset), "Contract Transactions", "open detail", "", nil
	default:
		return nil, "", "", "", nil
	}
}

func openOperationExplorerScope(ctx context.Context, backend lookupBackend, snapshot app.Snapshot, limit int, offset int) ([]backendclient.OperationSummary, string, string, string, error) {
	if snapshot.Current != app.ScreenLookup {
		return nil, "", "", "", nil
	}
	switch snapshot.Lookup.Kind {
	case app.LookupTransaction:
		if snapshot.Lookup.Transaction == nil || len(snapshot.Lookup.Transaction.Operations) == 0 {
			return nil, "", "", "", nil
		}
		operations := pagedSlice(snapshot.Lookup.Transaction.Operations, limit, offset)
		return operations, "Transaction Operations", "open detail", nextPageCommand("open ops", limit, offset, len(operations)), nil
	case app.LookupAccount:
		if snapshot.Lookup.Account == nil || snapshot.Lookup.Account.Account == nil {
			return nil, "", "", "", nil
		}
		accountID := snapshot.Lookup.Account.Account.ID
		if listBackend, ok := backend.(explorerListBackend); ok {
			operations, err := listBackend.AccountOperations(ctx, accountID, limit, offset)
			if err != nil {
				return nil, "", "", "", err
			}
			return operations, "Account Operations", "open detail", nextPageCommand("open ops", limit, offset, len(operations)), nil
		}
		operations := pagedSlice(snapshot.Lookup.Account.RecentOperations, limit, offset)
		return operations, "Account Operations", "open detail", nextPageCommand("open ops", limit, offset, len(operations)), nil
	default:
		return nil, "", "", "", nil
	}
}

func openHolderExplorerScope(ctx context.Context, backend lookupBackend, snapshot app.Snapshot, limit int, offset int) ([]backendclient.AssetHolderSummary, string, string, string, error) {
	if snapshot.Current != app.ScreenLookup || snapshot.Lookup.Kind != app.LookupAsset {
		return nil, "", "", "", nil
	}
	if snapshot.Lookup.Asset == nil || snapshot.Lookup.Asset.Asset == nil {
		return nil, "", "", "", nil
	}
	asset := snapshot.Lookup.Asset.Asset
	if listBackend, ok := backend.(explorerListBackend); ok {
		holders, err := listBackend.AssetHolders(ctx, asset.AssetCode, asset.AssetIssuer, limit, offset)
		if err != nil {
			return nil, "", "", "", err
		}
		return holders, "Asset Holders", "open detail", nextPageCommand("open holders", limit, offset, len(holders)), nil
	}
	holders := pagedSlice(snapshot.Lookup.Asset.TopHolders, limit, offset)
	return holders, "Asset Holders", "open detail", nextPageCommand("open holders", limit, offset, len(holders)), nil
}

func openStorageExplorerScope(ctx context.Context, backend lookupBackend, snapshot app.Snapshot, limit int, offset int) ([]backendclient.ContractStorageSummary, string, string, string, error) {
	if snapshot.Current != app.ScreenLookup || snapshot.Lookup.Kind != app.LookupContract {
		return nil, "", "", "", nil
	}
	if snapshot.Lookup.Contract == nil || snapshot.Lookup.Contract.Contract == nil {
		return nil, "", "", "", nil
	}
	contractID := snapshot.Lookup.Contract.Contract.ContractID
	if listBackend, ok := backend.(explorerListBackend); ok {
		entries, err := listBackend.ContractStorage(ctx, contractID, limit, offset)
		if err != nil {
			return nil, "", "", "", err
		}
		return entries, "Contract Storage", "open detail", nextPageCommand("open storage", limit, offset, len(entries)), nil
	}
	entries := pagedSlice(snapshot.Lookup.Contract.Storage, limit, offset)
	return entries, "Contract Storage", "open detail", nextPageCommand("open storage", limit, offset, len(entries)), nil
}

func openInvocationExplorerScope(ctx context.Context, backend lookupBackend, snapshot app.Snapshot, limit int, offset int) ([]backendclient.OperationSummary, string, string, string, error) {
	if snapshot.Current != app.ScreenLookup || snapshot.Lookup.Kind != app.LookupContract {
		return nil, "", "", "", nil
	}
	if snapshot.Lookup.Contract == nil || snapshot.Lookup.Contract.Contract == nil {
		return nil, "", "", "", nil
	}
	contractID := snapshot.Lookup.Contract.Contract.ContractID
	if listBackend, ok := backend.(explorerListBackend); ok {
		operations, err := listBackend.ContractInvocations(ctx, contractID, limit, offset)
		if err != nil {
			return nil, "", "", "", err
		}
		return operations, "Contract Invocations", "open detail", nextPageCommand("open invocations", limit, offset, len(operations)), nil
	}
	return nil, "Contract Invocations", "open detail", "", errors.New("contract invocations require an indexed backend")
}

func openEventExplorerScope(ctx context.Context, backend lookupBackend, snapshot app.Snapshot, limit int, offset int) ([]backendclient.ContractEventSummary, string, string, string, error) {
	if snapshot.Current != app.ScreenLookup || snapshot.Lookup.Kind != app.LookupContract {
		return nil, "", "", "", nil
	}
	if snapshot.Lookup.Contract == nil || snapshot.Lookup.Contract.Contract == nil {
		return nil, "", "", "", nil
	}
	contractID := snapshot.Lookup.Contract.Contract.ContractID
	if listBackend, ok := backend.(explorerListBackend); ok {
		events, err := listBackend.ContractEvents(ctx, contractID, limit, offset)
		if err != nil {
			return nil, "", "", "", err
		}
		return events, "Contract Events", "open detail", nextPageCommand("open events", limit, offset, len(events)), nil
	}
	events := pagedSlice(snapshot.Lookup.Contract.RecentEvents, limit, offset)
	return events, "Contract Events", "open detail", nextPageCommand("open events", limit, offset, len(events)), nil
}

type timelineItem struct {
	result app.SearchResult
	when   time.Time
	order  int
}

func openTimelineExplorerScope(ctx context.Context, backend lookupBackend, snapshot app.Snapshot, limit int, offset int, category string) ([]app.SearchResult, string, string, string, error) {
	if snapshot.Current != app.ScreenLookup {
		return nil, "", "", "", nil
	}

	items := []timelineItem{}
	switch snapshot.Lookup.Kind {
	case app.LookupAccount:
		if snapshot.Lookup.Account == nil || snapshot.Lookup.Account.Account == nil {
			return nil, "", "", "", nil
		}
		accountID := snapshot.Lookup.Account.Account.ID
		if listBackend, ok := backend.(explorerListBackend); ok {
			timeline, err := listBackend.AccountTimeline(ctx, accountID, limit, offset, category)
			if err == nil {
				return timelineSearchResults(timeline), timelineTitle("Account Timeline", category), "open detail", timelineNextPageCommand(limit, offset, len(timeline), category), nil
			}
			if len(snapshot.Lookup.Account.RecentTransactions) == 0 && len(snapshot.Lookup.Account.RecentOperations) == 0 {
				return nil, "", "", "", err
			}
		}
		if timelineCategoryMatches(category, "tx") {
			for index, tx := range snapshot.Lookup.Account.RecentTransactions {
				items = append(items, transactionTimelineItem(tx, index))
			}
		}
		if timelineCategoryMatches(category, "op") {
			for index, op := range snapshot.Lookup.Account.RecentOperations {
				items = append(items, operationTimelineItem(op, index))
			}
		}
		results := pagedSlice(sortedTimelineResults(items), limit, offset)
		return results, timelineTitle("Account Timeline", category), "open detail", timelineNextPageCommand(limit, offset, len(results), category), nil
	case app.LookupAsset:
		if snapshot.Lookup.Asset == nil || snapshot.Lookup.Asset.Asset == nil {
			return nil, "", "", "", nil
		}
		code := snapshot.Lookup.Asset.Asset.AssetCode
		issuer := snapshot.Lookup.Asset.Asset.AssetIssuer
		if listBackend, ok := backend.(explorerListBackend); ok {
			timeline, err := listBackend.AssetTimeline(ctx, code, issuer, limit, offset, category)
			if err == nil {
				return timelineSearchResults(timeline), timelineTitle("Asset Timeline", category), "open detail", timelineNextPageCommand(limit, offset, len(timeline), category), nil
			}
			if len(snapshot.Lookup.Asset.RecentTransactions) == 0 && len(snapshot.Lookup.Asset.TopHolders) == 0 {
				return nil, "", "", "", err
			}
		}
		if timelineCategoryMatches(category, "tx") {
			for index, tx := range snapshot.Lookup.Asset.RecentTransactions {
				items = append(items, transactionTimelineItem(tx, index))
			}
		}
		if timelineCategoryMatches(category, "holder") {
			for index, holder := range snapshot.Lookup.Asset.TopHolders {
				items = append(items, holderTimelineItem(holder, index))
			}
		}
		results := pagedSlice(sortedTimelineResults(items), limit, offset)
		return results, timelineTitle("Asset Timeline", category), "open detail", timelineNextPageCommand(limit, offset, len(results), category), nil
	case app.LookupContract:
		if snapshot.Lookup.Contract == nil || snapshot.Lookup.Contract.Contract == nil {
			return nil, "", "", "", nil
		}
		contractID := snapshot.Lookup.Contract.Contract.ContractID
		if listBackend, ok := backend.(explorerListBackend); ok {
			timeline, err := listBackend.ContractTimeline(ctx, contractID, limit, offset, category)
			if err == nil {
				return timelineSearchResults(timeline), timelineTitle("Contract Timeline", category), "open detail", timelineNextPageCommand(limit, offset, len(timeline), category), nil
			}
			if len(snapshot.Lookup.Contract.RecentTransactions) == 0 && len(snapshot.Lookup.Contract.RecentEvents) == 0 {
				return nil, "", "", "", err
			}
		}
		if timelineCategoryMatches(category, "tx") {
			for index, tx := range snapshot.Lookup.Contract.RecentTransactions {
				items = append(items, transactionTimelineItem(tx, index))
			}
		}
		if timelineCategoryMatches(category, "event") {
			for index, event := range snapshot.Lookup.Contract.RecentEvents {
				items = append(items, eventTimelineItem(event, index))
			}
		}
		results := pagedSlice(sortedTimelineResults(items), limit, offset)
		return results, timelineTitle("Contract Timeline", category), "open detail", timelineNextPageCommand(limit, offset, len(results), category), nil
	default:
		return nil, "", "", "", nil
	}
}

func timelineCategoryMatches(active string, candidate string) bool {
	active = strings.TrimSpace(active)
	return active == "" || active == candidate
}

func timelineTitle(base string, category string) string {
	switch strings.TrimSpace(category) {
	case "tx":
		return base + ": Transactions"
	case "op":
		return base + ": Operations"
	case "holder":
		return base + ": Holders"
	case "event":
		return base + ": Events"
	default:
		return base
	}
}

func timelineSearchResults(items []backendclient.TimelineItem) []app.SearchResult {
	results := make([]app.SearchResult, 0, len(items))
	for _, item := range items {
		results = append(results, app.SearchResult{
			Kind:        strings.TrimSpace(item.Kind),
			Title:       strings.TrimSpace(item.Title),
			Description: strings.TrimSpace(item.Description + "  " + formatTimelineTime(item.OccurredAt)),
			Command:     strings.TrimSpace(item.Command),
			Enabled:     true,
			Source:      "indexer",
		})
	}
	return results
}

func sortedTimelineResults(items []timelineItem) []app.SearchResult {
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].when.Equal(items[j].when) {
			if items[i].when.IsZero() {
				return false
			}
			if items[j].when.IsZero() {
				return true
			}
			return items[i].when.After(items[j].when)
		}
		return items[i].order < items[j].order
	})

	results := make([]app.SearchResult, 0, len(items))
	for _, item := range items {
		results = append(results, item.result)
	}
	return results
}

func transactionTimelineItem(tx backendclient.TransactionSummary, order int) timelineItem {
	return timelineItem{
		when:  tx.CreatedAt,
		order: order,
		result: app.SearchResult{
			Kind:        "tx",
			Title:       "Transaction " + truncateCommandLabel(tx.Hash, 16),
			Description: fmt.Sprintf("ledger %d  ops %d  %s", tx.LedgerSequence, tx.OperationCount, formatTimelineTime(tx.CreatedAt)),
			Command:     "lookup tx " + tx.Hash,
			Enabled:     true,
			Source:      "timeline",
		},
	}
}

func operationTimelineItem(op backendclient.OperationSummary, order int) timelineItem {
	return timelineItem{
		when:  op.CreatedAt,
		order: order,
		result: app.SearchResult{
			Kind:        "op",
			Title:       "Operation " + valueOrDefaultString(op.TypeName, "unknown"),
			Description: fmt.Sprintf("tx %s  %s", truncateCommandLabel(op.TransactionHash, 16), formatTimelineTime(op.CreatedAt)),
			Command:     timelineOperationCommand(op),
			Enabled:     true,
			Source:      "timeline",
		},
	}
}

func holderTimelineItem(holder backendclient.AssetHolderSummary, order int) timelineItem {
	return timelineItem{
		when:  holder.UpdatedAt,
		order: order,
		result: app.SearchResult{
			Kind:        "holder",
			Title:       "Holder " + truncateCommandLabel(holder.AccountID, 16),
			Description: fmt.Sprintf("balance %s  %s", holder.Balance, formatTimelineTime(holder.UpdatedAt)),
			Command:     "lookup account " + holder.AccountID,
			Enabled:     true,
			Source:      "timeline",
		},
	}
}

func eventTimelineItem(event backendclient.ContractEventSummary, order int) timelineItem {
	return timelineItem{
		when:  event.CreatedAt,
		order: order,
		result: app.SearchResult{
			Kind:        "event",
			Title:       fmt.Sprintf("Event type %d", event.Type),
			Description: fmt.Sprintf("ledger %d  tx %s  %s", event.LedgerSequence, truncateCommandLabel(event.TransactionHash, 16), formatTimelineTime(event.CreatedAt)),
			Command:     "lookup tx " + event.TransactionHash,
			Enabled:     true,
			Source:      "timeline",
		},
	}
}

func timelineOperationCommand(op backendclient.OperationSummary) string {
	if op.ContractID != nil && strings.TrimSpace(*op.ContractID) != "" {
		return "lookup contract " + strings.TrimSpace(*op.ContractID)
	}
	if op.Destination != nil && strings.TrimSpace(*op.Destination) != "" {
		return "lookup account " + strings.TrimSpace(*op.Destination)
	}
	if op.AssetCode != nil && op.AssetIssuer != nil && strings.TrimSpace(*op.AssetCode) != "" && strings.TrimSpace(*op.AssetIssuer) != "" {
		return fmt.Sprintf("lookup asset %s:%s", strings.TrimSpace(*op.AssetCode), strings.TrimSpace(*op.AssetIssuer))
	}
	if op.SourceAccount != nil && strings.TrimSpace(*op.SourceAccount) != "" {
		return "lookup account " + strings.TrimSpace(*op.SourceAccount)
	}
	if strings.TrimSpace(op.TransactionHash) != "" {
		return "lookup tx " + strings.TrimSpace(op.TransactionHash)
	}
	return ""
}

func formatTimelineTime(value time.Time) string {
	if value.IsZero() {
		return "unknown time"
	}
	return value.UTC().Format("2006-01-02 15:04")
}

func valueOrDefaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func pagedSlice[T any](values []T, limit int, offset int) []T {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= len(values) {
		return nil
	}
	end := offset + limit
	if end > len(values) {
		end = len(values)
	}
	return append([]T(nil), values[offset:end]...)
}

func sourceMetadataFor(profile config.Profile, operation string, backend lookupBackend) app.SourceMetadata {
	base := app.DefaultSourceMetadata(profile, operation)
	if provider, ok := backend.(sourceMetadataProvider); ok {
		meta := provider.SourceMetadata()
		if meta.Mode == "" {
			meta.Mode = base.Mode
		}
		if meta.Operation == "" {
			meta.Operation = base.Operation
		}
		if meta.Policy == "" {
			meta.Policy = base.Policy
		}
		if meta.Preferred == "" {
			meta.Preferred = base.Preferred
		}
		if meta.Actual == "" {
			meta.Actual = base.Actual
		}
		if meta.Label == "" {
			meta.Label = base.Label
		}
		return meta
	}
	base.Label = strings.TrimSpace(backend.Label())
	return base
}

func openEventDetail(ctx context.Context, model *app.Model, backend lookupBackend, index int) error {
	event, contractID, err := resolveEventSelection(model.Snapshot(), index-1)
	if err != nil {
		return err
	}
	source := sourceMetadataFor(model.Snapshot().Profile, string(app.LookupEvent), backend)
	query := fmt.Sprintf("%s:event-%d", contractID, index)
	model.UpdateLookupEvent(query, backendclient.ContractEventLookupSnapshot{
		ParentContractID: contractID,
		Event:            event,
	}, source)
	return nil
}

func openStorageEntryDetail(ctx context.Context, model *app.Model, backend lookupBackend, index int) error {
	entry, contractID, err := resolveStorageEntrySelection(ctx, model.Snapshot(), backend, index-1)
	if err != nil {
		return err
	}
	source := sourceMetadataFor(model.Snapshot().Profile, string(app.LookupStorage), backend)
	query := fmt.Sprintf("%s:storage-%d", contractID, index)
	model.UpdateLookupStorageEntry(query, backendclient.ContractStorageLookupSnapshot{
		ParentContractID: contractID,
		Entry:            entry,
	}, source)
	return nil
}

func openInvocationDetail(ctx context.Context, model *app.Model, backend lookupBackend, index int) error {
	op, parentHash, err := resolveInvocationSelection(ctx, model.Snapshot(), backend, index-1)
	if err != nil {
		return err
	}
	parentHash = strings.TrimSpace(parentHash)
	if parentHash == "" {
		parentHash = strings.TrimSpace(op.TransactionHash)
	}
	if parentHash == "" {
		return errors.New("invocation is missing parent transaction context")
	}
	source := sourceMetadataFor(model.Snapshot().Profile, string(app.LookupOperation), backend)
	query := fmt.Sprintf("%s:%d", parentHash, index)
	model.UpdateLookupOperation(query, backendclient.OperationLookupSnapshot{
		ParentTransactionHash: parentHash,
		Operation:             op,
	}, source)
	return nil
}

func resolveEventSelection(snapshot app.Snapshot, index int) (backendclient.ContractEventSummary, string, error) {
	if index < 0 {
		return backendclient.ContractEventSummary{}, "", errors.New("event index out of range")
	}
	contractID := contractIDFromSnapshot(snapshot)
	if contractID == "" {
		return backendclient.ContractEventSummary{}, "", errors.New("open event <n> is only available from a contract or event list")
	}
	if snapshot.Lookup.Kind == app.LookupContract && snapshot.Lookup.Contract != nil {
		events := snapshot.Lookup.Contract.RecentEvents
		if index >= len(events) {
			return backendclient.ContractEventSummary{}, "", fmt.Errorf("contract only has %d event(s)", len(events))
		}
		return events[index], contractID, nil
	}
	if snapshot.Lookup.Explorer != nil && snapshot.Lookup.Explorer.Kind == app.LookupExplorerEvents {
		events := snapshot.Lookup.Explorer.Events
		if index >= len(events) {
			return backendclient.ContractEventSummary{}, "", fmt.Errorf("event list only has %d row(s)", len(events))
		}
		return events[index], contractID, nil
	}
	return backendclient.ContractEventSummary{}, "", errors.New("open event <n> is only available from a contract or event list")
}

func resolveStorageEntrySelection(ctx context.Context, snapshot app.Snapshot, backend lookupBackend, index int) (backendclient.ContractStorageSummary, string, error) {
	if index < 0 {
		return backendclient.ContractStorageSummary{}, "", errors.New("storage index out of range")
	}
	contractID := contractIDFromSnapshot(snapshot)
	if contractID == "" {
		return backendclient.ContractStorageSummary{}, "", errors.New("open storage <n> is only available from a contract or storage list")
	}
	if snapshot.Lookup.Kind == app.LookupContract && snapshot.Lookup.Contract != nil {
		entries := snapshot.Lookup.Contract.Storage
		if index >= len(entries) {
			return backendclient.ContractStorageSummary{}, "", fmt.Errorf("contract only has %d storage row(s)", len(entries))
		}
		return entries[index], contractID, nil
	}
	if snapshot.Lookup.Explorer != nil && snapshot.Lookup.Explorer.Kind == app.LookupExplorerStorage {
		entries := snapshot.Lookup.Explorer.Storage
		if index >= len(entries) {
			return backendclient.ContractStorageSummary{}, "", fmt.Errorf("storage list only has %d row(s)", len(entries))
		}
		return entries[index], contractID, nil
	}
	return backendclient.ContractStorageSummary{}, "", errors.New("open storage <n> is only available from a contract or storage list")
}

func resolveInvocationSelection(ctx context.Context, snapshot app.Snapshot, backend lookupBackend, index int) (backendclient.OperationSummary, string, error) {
	if index < 0 {
		return backendclient.OperationSummary{}, "", errors.New("invocation index out of range")
	}
	if snapshot.Lookup.Explorer != nil && snapshot.Lookup.Explorer.Kind == app.LookupExplorerInvocations {
		operations := snapshot.Lookup.Explorer.Operations
		if index >= len(operations) {
			return backendclient.OperationSummary{}, "", fmt.Errorf("invocation list only has %d row(s)", len(operations))
		}
		op := operations[index]
		return op, strings.TrimSpace(op.TransactionHash), nil
	}
	contractID := contractIDFromSnapshot(snapshot)
	if contractID == "" {
		return backendclient.OperationSummary{}, "", errors.New("open invocation <n> is only available from a contract invocation list")
	}
	if listBackend, ok := backend.(explorerListBackend); ok {
		limit := 10
		offset := 0
		if snapshot.Lookup.Explorer != nil {
			limit = snapshot.Lookup.Explorer.ListLimit
			offset = snapshot.Lookup.Explorer.ListOffset
		}
		operations, err := listBackend.ContractInvocations(ctx, contractID, limit, offset)
		if err != nil {
			return backendclient.OperationSummary{}, "", err
		}
		if index >= len(operations) {
			return backendclient.OperationSummary{}, "", fmt.Errorf("invocation list only has %d row(s)", len(operations))
		}
		op := operations[index]
		return op, strings.TrimSpace(op.TransactionHash), nil
	}
	return backendclient.OperationSummary{}, "", errors.New("open invocation <n> requires an indexed backend")
}

func contractIDFromSnapshot(snapshot app.Snapshot) string {
	if snapshot.Lookup.Kind == app.LookupContract && snapshot.Lookup.Contract != nil && snapshot.Lookup.Contract.Contract != nil {
		return strings.TrimSpace(snapshot.Lookup.Contract.Contract.ContractID)
	}
	if snapshot.Lookup.Explorer != nil {
		parent := strings.TrimSpace(snapshot.Lookup.Explorer.ParentLabel)
		if strings.HasPrefix(parent, "contract ") {
			return strings.TrimSpace(strings.TrimPrefix(parent, "contract "))
		}
	}
	return ""
}

func openOperationDetail(ctx context.Context, model *app.Model, backend lookupBackend, index int, explicitTxHash string) error {
	op, parentHash, err := resolveOperationSelection(model.Snapshot(), index-1)
	if err != nil {
		if !errors.Is(err, errOperationContextMissing) {
			return err
		}
		explicitTxHash = strings.TrimSpace(explicitTxHash)
		if explicitTxHash == "" {
			return errors.New("open op <n> is only available from a transaction or operation list")
		}
		response, fetchErr := backend.Transaction(ctx, explicitTxHash)
		if fetchErr != nil {
			return fetchErr
		}
		op, err = resolveOperationInTransaction(response.Operations, index)
		if err != nil {
			return err
		}
		parentHash = explicitTxHash
		if response.Transaction != nil && strings.TrimSpace(response.Transaction.Hash) != "" {
			parentHash = strings.TrimSpace(response.Transaction.Hash)
		}
	}

	parentHash = strings.TrimSpace(parentHash)
	if parentHash == "" {
		parentHash = strings.TrimSpace(op.TransactionHash)
	}
	if parentHash == "" {
		return errors.New("operation is missing parent transaction context")
	}

	source := sourceMetadataFor(model.Snapshot().Profile, string(app.LookupOperation), backend)
	query := fmt.Sprintf("%s:%d", parentHash, index)
	model.UpdateLookupOperation(query, backendclient.OperationLookupSnapshot{
		ParentTransactionHash: parentHash,
		Operation:             op,
	}, source)
	return nil
}

func resolveOperationInTransaction(operations []backendclient.OperationSummary, order int) (backendclient.OperationSummary, error) {
	if order < 1 {
		return backendclient.OperationSummary{}, errors.New("operation index must be a positive integer")
	}
	for _, op := range operations {
		if int(op.ApplicationOrder) == order {
			return op, nil
		}
	}
	if order > len(operations) {
		return backendclient.OperationSummary{}, fmt.Errorf("transaction only has %d operation(s)", len(operations))
	}
	return operations[order-1], nil
}

var errOperationContextMissing = errors.New("operation context missing")

func resolveOperationSelection(snapshot app.Snapshot, index int) (backendclient.OperationSummary, string, error) {
	if index < 0 {
		return backendclient.OperationSummary{}, "", errors.New("operation index out of range")
	}

	lookup := snapshot.Lookup
	if lookup.Kind == app.LookupTransaction && lookup.Transaction != nil {
		operations := lookup.Transaction.Operations
		if index >= len(operations) {
			return backendclient.OperationSummary{}, "", fmt.Errorf("transaction only has %d operation(s)", len(operations))
		}
		parentHash := ""
		if lookup.Transaction.Transaction != nil {
			parentHash = strings.TrimSpace(lookup.Transaction.Transaction.Hash)
		}
		return operations[index], parentHash, nil
	}

	if lookup.Explorer != nil && lookup.Explorer.Kind == app.LookupExplorerOperations {
		operations := lookup.Explorer.Operations
		if index >= len(operations) {
			return backendclient.OperationSummary{}, "", fmt.Errorf("operation list only has %d row(s)", len(operations))
		}
		return operations[index], strings.TrimSpace(operations[index].TransactionHash), nil
	}

	if lookup.Explorer != nil && lookup.Explorer.Kind == app.LookupExplorerInvocations {
		operations := lookup.Explorer.Operations
		if index >= len(operations) {
			return backendclient.OperationSummary{}, "", fmt.Errorf("invocation list only has %d row(s)", len(operations))
		}
		return operations[index], strings.TrimSpace(operations[index].TransactionHash), nil
	}

	return backendclient.OperationSummary{}, "", errOperationContextMissing
}

func normalizeLookupKind(value string) app.LookupKind {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "ledger":
		return app.LookupLedger
	case "tx", "transaction":
		return app.LookupTransaction
	case "op", "operation":
		return app.LookupOperation
	case "account", "acct":
		return app.LookupAccount
	case "asset":
		return app.LookupAsset
	case "contract":
		return app.LookupContract
	default:
		return app.LookupKind(strings.ToLower(strings.TrimSpace(value)))
	}
}

func isInteractive(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

// normalizeArgs makes test-driven execution easier if later we need to inject
// command aliases or shell-friendly forms.
func normalizeArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(args))
	for _, arg := range args {
		normalized = append(normalized, strings.TrimSpace(arg))
	}

	return normalized
}

type startupState struct {
	store          *cache.Store
	snapshot       app.CacheSnapshot
	restoredScreen string
}

func initializeStartupState(cfg config.Config) (startupState, error) {
	if cfg.Cache.Driver == "" {
		return startupState{
			snapshot: app.BuildCacheSnapshot(cfg, nil, "", 0, nil, "", "Cache disabled in config."),
		}, nil
	}

	cachePath, err := config.ResolveCachePath(cfg.Cache.Path)
	if err != nil {
		return startupState{}, err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return startupState{}, fmt.Errorf("create cache directory: %w", err)
	}

	store, err := openCacheStore(context.Background(), cfg.Cache.Driver, cachePath)
	if err != nil {
		return startupState{
			snapshot: app.BuildCacheSnapshot(cfg, nil, cachePath, 0, nil, "", fmt.Sprintf("Cache unavailable: %v", err)),
		}, nil
	}

	profiles, err := seedProfiles(context.Background(), store, cfg)
	if err != nil {
		_ = store.Close()
		return startupState{}, err
	}

	schemaVersion, err := store.SchemaVersion(context.Background())
	if err != nil {
		_ = store.Close()
		return startupState{}, err
	}

	lastScreen, err := loadSessionScreen(context.Background(), store)
	if err != nil {
		_ = store.Close()
		return startupState{}, err
	}

	return startupState{
		store:          store,
		restoredScreen: lastScreen,
		snapshot: app.BuildCacheSnapshot(
			cfg,
			store,
			cachePath,
			schemaVersion,
			profiles,
			lastScreen,
			fmt.Sprintf("Cache ready at %s with %d profile(s).", cachePath, len(profiles)),
		),
	}, nil
}

func seedProfiles(ctx context.Context, store *cache.Store, cfg config.Config) ([]cache.Profile, error) {
	if store == nil {
		return nil, nil
	}

	for _, profile := range cfg.Profiles {
		record := cache.Profile{
			ID:         profile.Name,
			Name:       profile.Name,
			Network:    profile.Network,
			RPCURL:     profile.RPCEndpoint,
			IndexerURL: profile.IndexerURL,
		}
		if err := store.UpsertProfile(ctx, record); err != nil {
			return nil, fmt.Errorf("seed profile %q: %w", profile.Name, err)
		}
	}

	profiles, err := store.ListProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cached profiles: %w", err)
	}

	return profiles, nil
}

func loadSessionScreen(ctx context.Context, store *cache.Store) (string, error) {
	if store == nil {
		return "", nil
	}

	var session sessionState
	if err := store.GetStateJSON(ctx, stateKeySession, &session); err != nil {
		if errors.Is(err, cache.ErrNoRows()) {
			return "", nil
		}
		return "", fmt.Errorf("load session state: %w", err)
	}

	return session.LastScreen, nil
}

func persistSessionState(store *cache.Store, current app.Screen) error {
	if store == nil {
		return nil
	}

	session := sessionState{LastScreen: string(current)}
	if err := store.SetStateJSON(context.Background(), stateKeySession, session); err != nil {
		return fmt.Errorf("save session state: %w", err)
	}

	return nil
}

func restoreLiveFeedScrollback(ctx context.Context, store *cache.Store, cfg config.Config, model *app.Model) error {
	if store == nil || model == nil {
		return nil
	}
	profile, ok := cfg.Profile(cfg.DefaultProfile)
	if !ok {
		return nil
	}
	rows, err := store.ListLiveTransactions(ctx, profile.Name, 100)
	if err != nil {
		return fmt.Errorf("load cached live transactions: %w", err)
	}
	model.RestoreLiveFeedScrollback(cacheLiveTransactionsToBackend(rows))
	return nil
}

func persistLiveFeedScrollback(ctx context.Context, store *cache.Store, profileID string, transactions []backendclient.TransactionSummary) error {
	if store == nil || strings.TrimSpace(profileID) == "" || len(transactions) == 0 {
		return nil
	}
	if len(transactions) > 100 {
		transactions = transactions[:100]
	}
	if err := store.UpsertLiveTransactions(ctx, backendTransactionsToCache(profileID, transactions)); err != nil {
		return fmt.Errorf("save cached live transactions: %w", err)
	}
	return nil
}

func backendTransactionsToCache(profileID string, transactions []backendclient.TransactionSummary) []cache.LiveTransaction {
	rows := make([]cache.LiveTransaction, 0, len(transactions))
	for _, tx := range transactions {
		if strings.TrimSpace(tx.Hash) == "" {
			continue
		}
		rows = append(rows, cache.LiveTransaction{
			ProfileID:        profileID,
			Hash:             tx.Hash,
			LedgerSequence:   tx.LedgerSequence,
			ApplicationOrder: tx.ApplicationOrder,
			Account:          tx.Account,
			OperationCount:   tx.OperationCount,
			Status:           tx.Status,
			IsSoroban:        tx.IsSoroban,
			CreatedAt:        tx.CreatedAt,
		})
	}
	return rows
}

func cacheLiveTransactionsToBackend(transactions []cache.LiveTransaction) []backendclient.TransactionSummary {
	rows := make([]backendclient.TransactionSummary, 0, len(transactions))
	for _, tx := range transactions {
		rows = append(rows, backendclient.TransactionSummary{
			Hash:             tx.Hash,
			LedgerSequence:   tx.LedgerSequence,
			ApplicationOrder: tx.ApplicationOrder,
			Account:          tx.Account,
			OperationCount:   tx.OperationCount,
			Status:           tx.Status,
			IsSoroban:        tx.IsSoroban,
			CreatedAt:        tx.CreatedAt,
		})
	}
	return rows
}
