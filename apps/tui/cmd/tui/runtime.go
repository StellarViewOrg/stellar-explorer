package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/clipboard"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/ui"
)

type liveFeedTickMsg struct{}

var (
	readClipboard  = clipboard.Read
	writeClipboard = clipboard.Write
)

type interactiveRuntime struct {
	tty   *os.File
	out   io.Writer
	store *cache.Store
}

func runInteractive(cfg config.Config, model *app.Model, store *cache.Store, stdin *os.File, stdout io.Writer, stderr io.Writer) int {
	programModel := newTeaRuntimeModel(cfg, model, store)
	program := tea.NewProgram(
		programModel,
		tea.WithAltScreen(),
		tea.WithInput(stdin),
		tea.WithOutput(stdout),
	)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(stderr, "run tui: %v\n", err)
		return 1
	}
	if err := persistSessionState(store, model.Snapshot().Current); err != nil {
		fmt.Fprintf(stderr, "persist session: %v\n", err)
		return 1
	}

	return 0
}

func (r interactiveRuntime) render(model *app.Model) error {
	if err := persistSessionState(r.store, model.Snapshot().Current); err != nil {
		return err
	}
	_, err := fmt.Fprint(r.out, ui.Render(model.Snapshot()))
	return err
}

type teaRuntimeModel struct {
	cfg        config.Config
	model      *app.Model
	store      *cache.Store
	runtime    interactiveRuntime
	dashboard  ui.DashboardModel
	liveStream *liveStreamListener
}

func newTeaRuntimeModel(cfg config.Config, model *app.Model, store *cache.Store) teaRuntimeModel {
	listener := newLiveStreamListener(cfg)
	return teaRuntimeModel{
		cfg:        cfg,
		model:      model,
		store:      store,
		runtime:    interactiveRuntime{store: store},
		dashboard:  ui.NewDashboardModel(model.Snapshot(), 104, 34),
		liveStream: listener,
	}
}

func (m teaRuntimeModel) Init() tea.Cmd {
	cmds := []tea.Cmd{liveFeedTick()}
	if m.liveStream != nil {
		cmds = append(cmds, m.liveStream.wait())
	}
	return tea.Batch(cmds...)
}

func (m teaRuntimeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.dashboard = m.dashboard.WithSnapshot(m.model.Snapshot())
		if updated, _ := m.dashboard.Update(msg); updated != nil {
			if dashboard, ok := updated.(ui.DashboardModel); ok {
				m.dashboard = dashboard
			}
		}
		return m, nil
	case tea.KeyMsg:
		m.dashboard = m.dashboard.WithSnapshot(m.model.Snapshot())
		keepRunning, err := m.runtime.handleTeaKey(context.Background(), m.cfg, m.model, m.store, &m.dashboard, msg)
		if err != nil {
			m.model.SetErrorStatus(fmt.Sprintf("Interactive action failed: %v", err))
		}
		if err := persistSessionState(m.store, m.model.Snapshot().Current); err != nil {
			m.model.SetErrorStatus(fmt.Sprintf("Persist session failed: %v", err))
		}
		if !keepRunning {
			return m, tea.Quit
		}
		return m, nil
	case liveFeedTickMsg:
		refreshCurrentScreen(context.Background(), m.cfg, m.model, m.store)
		return m, liveFeedTick()
	case liveStreamUpdateMsg:
		if m.model != nil && m.model.Snapshot().Current == app.ScreenLiveFeed && !m.model.LiveFeedPaused() {
			applyLiveStreamUpdate(m.model, msg.Update)
			if m.store != nil {
				profile, ok := m.cfg.Profile(m.cfg.DefaultProfile)
				if ok {
					_ = persistLiveFeedScrollback(context.Background(), m.store, profile.Name, m.model.Snapshot().LiveFeed.Scrollback)
				}
			}
		}
		if m.liveStream != nil {
			return m, m.liveStream.wait()
		}
		return m, nil
	case liveStreamErrorMsg:
		handleLiveStreamError(m.model, msg.Err)
		if m.liveStream != nil {
			return m, m.liveStream.wait()
		}
		return m, nil
	default:
		return m, nil
	}
}

func liveFeedTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return liveFeedTickMsg{}
	})
}

func (m teaRuntimeModel) View() string {
	m.dashboard = m.dashboard.WithSnapshot(m.model.Snapshot())
	return m.dashboard.View()
}

func (r interactiveRuntime) handleInput(ctx context.Context, cfg config.Config, model *app.Model, input []byte) (bool, error) {
	if model.Snapshot().Command.Visible {
		return r.handleCommandPaletteInput(ctx, cfg, model, input)
	}

	key := normalizeInteractiveKey(input)
	switch key {
	case keyUnknown:
		return true, nil
	case keyQuit:
		return false, nil
	case keyHome:
		return true, executeInteractiveCommand(model, "home")
	case keyLiveFeed:
		return true, executeInteractiveCommand(model, "live")
	case keyLookup:
		return true, executeInteractiveCommand(model, "lookup")
	case keySettings:
		return true, executeInteractiveCommand(model, "settings")
	case keyBack:
		return true, executeInteractiveCommand(model, "back")
	case keyForward:
		return true, executeInteractiveCommand(model, "forward")
	case keyHelp:
		model.ToggleHelpOverlay()
		return true, nil
	case keyRefresh:
		refreshCurrentScreen(ctx, cfg, model, r.store)
		return true, nil
	case keySearch:
		model.OpenCommandPalette("Search / Command", "")
		return true, nil
	case keyFocusNext:
		model.FocusNext()
		return true, nil
	case keyMoveDown:
		model.MoveSelection(1)
		return true, nil
	case keyMoveUp:
		model.MoveSelection(-1)
		return true, nil
	case keyEnter:
		return true, activateSelection(ctx, cfg, model, nil, r.store)
	default:
		return true, nil
	}
}

func (r interactiveRuntime) handleCommandPaletteInput(ctx context.Context, cfg config.Config, model *app.Model, input []byte) (bool, error) {
	key := normalizeInteractiveKey(input)
	switch key {
	case keyQuit:
		return false, nil
	case keyEscape:
		model.CloseCommandPalette()
		return true, nil
	case keyEnter:
		command := strings.TrimSpace(model.Snapshot().Command.Input)
		if result := model.SelectedCommandPaletteResult(); result != nil {
			if !result.Enabled {
				model.CloseCommandPalette()
				model.SetWarningStatus(fmt.Sprintf("%s lookup is not implemented yet.", result.Title))
				return true, nil
			}
			command = result.Command
		}
		model.CloseCommandPalette()
		if command == "" {
			return true, nil
		}
		keepRunning, err := executeCommand(ctx, cfg, model, command, r.store)
		if err != nil {
			return false, err
		}
		refreshCurrentScreen(ctx, cfg, model, r.store)
		return keepRunning, nil
	case keyBackspace:
		model.TrimCommandPaletteInput()
		refreshCommandPaletteSearch(ctx, cfg, model, r.store)
		return true, nil
	case keyMoveDown:
		model.MoveCommandPaletteSelection(1)
		return true, nil
	case keyMoveUp:
		model.MoveCommandPaletteSelection(-1)
		return true, nil
	default:
		model.AppendCommandPaletteInput(printableInput(input))
		refreshCommandPaletteSearch(ctx, cfg, model, r.store)
		return true, nil
	}
}

func (r interactiveRuntime) handleTeaKey(ctx context.Context, cfg config.Config, model *app.Model, store *cache.Store, dashboard *ui.DashboardModel, msg tea.KeyMsg) (bool, error) {
	updated, cmd := dashboard.Update(msg)
	if updatedDashboard, ok := updated.(ui.DashboardModel); ok {
		*dashboard = updatedDashboard
		syncLookupSelection(model, dashboard)
	}
	if cmd == nil {
		return true, nil
	}

	action, ok := cmd().(ui.ActionMsg)
	if !ok {
		return true, nil
	}
	return r.applyAction(ctx, cfg, model, store, dashboard, action)
}

func syncLookupSelection(model *app.Model, dashboard *ui.DashboardModel) {
	if dashboard == nil || model.Snapshot().Current != app.ScreenLookup {
		return
	}
	section, offset := dashboard.LookupSelection()
	model.SetLookupSelection(section, offset)
}

func (r interactiveRuntime) handleTeaCommandPaletteKey(ctx context.Context, cfg config.Config, model *app.Model, key string) (bool, error) {
	switch key {
	case "ctrl+c":
		return false, nil
	case "esc":
		model.CloseCommandPalette()
		return true, nil
	case "enter":
		command := strings.TrimSpace(model.Snapshot().Command.Input)
		if result := model.SelectedCommandPaletteResult(); result != nil {
			if !result.Enabled {
				model.CloseCommandPalette()
				model.SetWarningStatus(fmt.Sprintf("%s lookup is not implemented yet.", result.Title))
				return true, nil
			}
			command = result.Command
		}
		model.CloseCommandPalette()
		if command == "" {
			return true, nil
		}
		keepRunning, err := executeCommand(ctx, cfg, model, command, r.store)
		if err != nil {
			return false, err
		}
		refreshCurrentScreen(ctx, cfg, model, r.store)
		return keepRunning, nil
	case "backspace":
		model.TrimCommandPaletteInput()
		refreshCommandPaletteSearch(ctx, cfg, model, r.store)
		return true, nil
	case "down", "ctrl+n":
		model.MoveCommandPaletteSelection(1)
		return true, nil
	case "up", "ctrl+p":
		model.MoveCommandPaletteSelection(-1)
		return true, nil
	default:
		if isPrintableTeaKey(key) {
			model.AppendCommandPaletteInput(key)
			refreshCommandPaletteSearch(ctx, cfg, model, r.store)
		}
		return true, nil
	}
}

func (r interactiveRuntime) applyAction(ctx context.Context, cfg config.Config, model *app.Model, store *cache.Store, dashboard *ui.DashboardModel, action ui.ActionMsg) (bool, error) {
	switch action.Kind {
	case ui.ActionNone:
		return true, nil
	case ui.ActionQuit:
		return false, nil
	case ui.ActionHome:
		return true, executeInteractiveCommand(model, "home")
	case ui.ActionLiveFeed:
		if err := executeInteractiveCommand(model, "live"); err != nil {
			return true, err
		}
		profile, ok := cfg.Profile(cfg.DefaultProfile)
		if ok {
			applyProfileWatchOnLiveScreen(ctx, model, store, profile)
		}
		return true, nil
	case ui.ActionLookup:
		return true, executeInteractiveCommand(model, "lookup")
	case ui.ActionSettings:
		return true, executeInteractiveCommand(model, "settings")
	case ui.ActionBack:
		syncLookupSelection(model, dashboard)
		return true, executeInteractiveCommand(model, "back")
	case ui.ActionForward:
		syncLookupSelection(model, dashboard)
		return true, executeInteractiveCommand(model, "forward")
	case ui.ActionToggleHelp:
		model.ToggleHelpOverlay()
		return true, nil
	case ui.ActionRefresh:
		refreshCurrentScreen(ctx, cfg, model, store)
		return true, nil
	case ui.ActionToggleLivePause:
		model.SetLiveFeedPaused(!model.LiveFeedPaused())
		return true, nil
	case ui.ActionCycleLiveFilter:
		model.CycleLiveFeedFilter()
		return true, nil
	case ui.ActionShiftLiveReplay:
		model.ShiftLiveFeedReplay(action.Delta)
		return true, nil
	case ui.ActionOpenCommandPalette:
		model.OpenCommandPalette("Search / Command", "")
		dashboard.OpenCommandPalette("Search / Command", "")
		return true, nil
	case ui.ActionFocusNext:
		model.FocusNextAvailable(dashboard.FocusOrder()...)
		return true, nil
	case ui.ActionMoveSelection:
		if message := selectionStatusMessage(model.Snapshot(), dashboard); strings.TrimSpace(message) != "" {
			model.SetInfoStatus(message)
		}
		return true, nil
	case ui.ActionActivateSelection:
		syncLookupSelection(model, dashboard)
		return true, activateSelection(ctx, cfg, model, dashboard, store)
	case ui.ActionCloseCommand:
		model.CloseCommandPalette()
		dashboard.CloseCommandPalette()
		return true, nil
	case ui.ActionSubmitCommand:
		return r.submitCommandPalette(ctx, cfg, model, dashboard)
	case ui.ActionCommandBackspace:
		results, err := loadCommandPaletteResults(ctx, cfg, store, dashboard.CommandInput(), model.Snapshot().Profile.Name, commandPaletteSearchLimit(model))
		if err != nil {
			model.SetWarningStatus(err.Error())
		} else {
			dashboard.SetCommandPaletteResults(results)
		}
		return true, nil
	case ui.ActionMoveCommandResult:
		return true, nil
	case ui.ActionAppendCommandInput:
		results, err := loadCommandPaletteResults(ctx, cfg, store, dashboard.CommandInput(), model.Snapshot().Profile.Name, commandPaletteSearchLimit(model))
		if err != nil {
			model.SetWarningStatus(err.Error())
		} else {
			dashboard.SetCommandPaletteResults(results)
		}
		return true, nil
	case ui.ActionPasteClipboard:
		text, err := readClipboard(ctx)
		if err != nil {
			model.SetWarningStatus(fmt.Sprintf("Clipboard paste unavailable: %v", err))
			return true, nil
		}
		text = normalizeClipboardText(text)
		if text == "" {
			model.SetWarningStatus("Clipboard is empty.")
			return true, nil
		}
		dashboard.AppendCommandPaletteInput(text)
		results, err := loadCommandPaletteResults(ctx, cfg, store, dashboard.CommandInput(), model.Snapshot().Profile.Name, commandPaletteSearchLimit(model))
		if err != nil {
			model.SetWarningStatus(err.Error())
		} else {
			dashboard.SetCommandPaletteResults(results)
			model.SetInfoStatus(fmt.Sprintf("Pasted %d chars from clipboard.", len([]rune(text))))
		}
		return true, nil
	case ui.ActionCopySelection:
		value := clipboardSelection(model.Snapshot(), dashboard)
		if strings.TrimSpace(value) == "" {
			model.SetWarningStatus("Nothing copyable is selected on the current screen.")
			return true, nil
		}
		if err := writeClipboard(ctx, value); err != nil {
			model.SetWarningStatus(fmt.Sprintf("Clipboard copy unavailable: %v", err))
			return true, nil
		}
		model.SetInfoStatus(fmt.Sprintf("Copied to clipboard: %s", truncateStatusValue(value)))
		return true, nil
	case ui.ActionMarkBookmark:
		snapshot := model.Snapshot()
		seed := "bookmark add "
		if snapshot.Current == app.ScreenLookup && snapshot.Lookup.State == app.ViewStateReady {
			seed = "bookmark add " + string(snapshot.Lookup.Kind) + " " + strings.TrimSpace(snapshot.Lookup.Query)
		}
		model.OpenCommandPalette("Bookmark Entity", seed)
		dashboard.OpenCommandPalette("Bookmark Entity", seed)
		return true, nil
	case ui.ActionQuickBookmark:
		_, err := executeWorkspaceCommand(ctx, model, store, "bookmark add")
		return true, err
	case ui.ActionRemoveBookmark:
		_, err := executeWorkspaceCommand(ctx, model, store, "bookmark remove")
		return true, err
	case ui.ActionOpenNotePalette:
		model.OpenCommandPalette("Add Note", strings.TrimSpace(action.Text))
		dashboard.OpenCommandPalette("Add Note", strings.TrimSpace(action.Text))
		return true, nil
	case ui.ActionOpenLabelPalette:
		model.OpenCommandPalette("Add Label", strings.TrimSpace(action.Text))
		dashboard.OpenCommandPalette("Add Label", strings.TrimSpace(action.Text))
		return true, nil
	case ui.ActionCycleContractTab:
		if err := model.CycleContractWorkspaceTab(action.Delta); err != nil {
			model.SetWarningStatus(err.Error())
		}
		return true, nil
	case ui.ActionSetContractDecodeRaw:
		if err := model.SetContractDecodeMode(app.ContractDecodeModeRaw); err != nil {
			model.SetWarningStatus(err.Error())
		}
		return true, nil
	case ui.ActionSetContractDecodeDecoded:
		if err := model.SetContractDecodeMode(app.ContractDecodeModeDecoded); err != nil {
			model.SetWarningStatus(err.Error())
		}
		return true, nil
	case ui.ActionToggleLookupExpand:
		if err := model.ToggleLookupExpandAll(); err != nil {
			model.SetWarningStatus(err.Error())
		}
		return true, nil
	case ui.ActionToggleLookupVisual:
		if err := model.ToggleLookupVisualMode(); err != nil {
			model.SetWarningStatus(err.Error())
		}
		return true, nil
	default:
		return true, nil
	}
}

func (r interactiveRuntime) submitCommandPalette(ctx context.Context, cfg config.Config, model *app.Model, dashboard *ui.DashboardModel) (bool, error) {
	syncLookupSelection(model, dashboard)
	command := strings.TrimSpace(dashboard.CommandInput())
	if result := dashboard.SelectedCommandPaletteResult(); result != nil {
		if !result.Enabled {
			model.CloseCommandPalette()
			dashboard.CloseCommandPalette()
			model.SetWarningStatus(fmt.Sprintf("%s lookup is not implemented yet.", result.Title))
			return true, nil
		}
		command = result.Command
	}
	model.CloseCommandPalette()
	dashboard.CloseCommandPalette()
	if command == "" {
		return true, nil
	}
	keepRunning, err := executeCommand(ctx, cfg, model, command, r.store)
	if err != nil {
		return false, err
	}
	refreshCurrentScreen(ctx, cfg, model, r.store)
	return keepRunning, nil
}

func executeInteractiveCommand(model *app.Model, command string) error {
	if keepRunning := model.HandleCommand(command); !keepRunning {
		return nil
	}
	return nil
}

func refreshCommandPaletteSearch(ctx context.Context, cfg config.Config, model *app.Model, store *cache.Store) {
	snapshot := model.Snapshot()
	query := strings.TrimSpace(snapshot.Command.Input)
	if !snapshot.Command.Visible || query == "" || len(query) < 2 {
		return
	}

	limit := commandPaletteSearchLimit(model)
	if results, err := localMetadataSearchResults(ctx, store, query, snapshot.Profile.Name, limit); err == nil {
		model.MergeCommandPaletteLocalResults(results)
	} else {
		model.SetWarningStatus(fmt.Sprintf("Local metadata search unavailable: %v", err))
	}

	backend, err := initializeLookupBackend(cfg)
	if err != nil || backend == nil {
		return
	}

	response, err := backend.Search(ctx, query, limit)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Search backend unavailable: %v", err))
		return
	}
	model.MergeCommandPaletteBackendResults(response.Results)
	if len(response.Results) >= limit {
		merged := append(model.Snapshot().Command.Results, app.SearchMoreResult(query, limit))
		model.SetCommandPaletteResults(app.RankSearchResults(query, merged))
	}
}

func commandPaletteSearchLimit(model *app.Model) int {
	if model == nil {
		return app.SearchResultLimit()
	}
	limit := model.Snapshot().Command.SearchLimit
	if limit <= 0 {
		return app.SearchResultLimit()
	}
	return limit
}

func loadCommandPaletteResults(ctx context.Context, cfg config.Config, store *cache.Store, query string, profileName string, limit int) ([]app.SearchResult, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 {
		limit = app.SearchResultLimit()
	}
	results := app.InferSearchResults(query)
	if query == "" || len(query) < 2 {
		return results, nil
	}

	if localResults, err := localMetadataSearchResults(ctx, store, query, profileName, limit); err == nil {
		results = mergeLocalSearchResults(results, localResults)
	} else {
		return results, fmt.Errorf("local metadata search unavailable: %w", err)
	}

	backend, err := initializeLookupBackend(cfg)
	if err != nil || backend == nil {
		return app.RankSearchResults(query, results), nil
	}

	response, err := backend.Search(ctx, query, limit)
	if err != nil {
		return results, fmt.Errorf("search backend unavailable: %w", err)
	}
	results = app.RankSearchResults(query, mergeBackendSearchResults(results, response.Results))
	if len(response.Results) >= limit {
		results = append(results, app.SearchMoreResult(query, limit))
	}
	return results, nil
}

func mergeBackendSearchResults(base []app.SearchResult, results []backendclient.SearchResult) []app.SearchResult {
	merged := append([]app.SearchResult(nil), base...)
	seen := make(map[string]struct{}, len(merged)+len(results))
	for _, result := range merged {
		seen[searchResultDedupKey(result)] = struct{}{}
	}
	for _, result := range results {
		command := strings.TrimSpace(result.Command)
		if command == "" {
			continue
		}
		if _, ok := seen[command]; ok {
			continue
		}
		seen[command] = struct{}{}
		source := strings.TrimSpace(result.Source)
		if source == "" {
			source = "indexer"
		}
		merged = append(merged, app.SearchResult{
			Kind:        strings.TrimSpace(result.Kind),
			Title:       strings.TrimSpace(result.Title),
			Description: strings.TrimSpace(result.Description),
			Command:     command,
			Enabled:     true,
			Source:      source,
		})
	}
	return app.RankSearchResults("", merged)
}

func normalizeClipboardText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Fields(value)
	return strings.TrimSpace(strings.Join(lines, " "))
}

func clipboardSelection(snapshot app.Snapshot, dashboard *ui.DashboardModel) string {
	switch snapshot.Current {
	case app.ScreenLiveFeed:
		if dashboard != nil {
			if value := dashboard.LiveFeedClipboardValue(); strings.TrimSpace(value) != "" {
				return value
			}
		}
		index := snapshot.Selection.LiveFeedIndex
		if index >= 0 && index < len(snapshot.LiveFeed.RecentTransactions) {
			return strings.TrimSpace(snapshot.LiveFeed.RecentTransactions[index].Hash)
		}
	case app.ScreenLookup:
		if dashboard != nil {
			if value := dashboard.LookupClipboardValue(); strings.TrimSpace(value) != "" {
				return value
			}
		}
		switch snapshot.Lookup.Kind {
		case app.LookupLedger:
			if snapshot.Lookup.Ledger != nil && snapshot.Lookup.Ledger.Ledger != nil {
				return fmt.Sprintf("%d", snapshot.Lookup.Ledger.Ledger.Sequence)
			}
		case app.LookupTransaction:
			if snapshot.Lookup.Transaction != nil && snapshot.Lookup.Transaction.Transaction != nil {
				return strings.TrimSpace(snapshot.Lookup.Transaction.Transaction.Hash)
			}
		case app.LookupOperation:
			if snapshot.Lookup.Operation != nil {
				return strings.TrimSpace(snapshot.Lookup.Operation.ParentTransactionHash)
			}
		case app.LookupAccount:
			if snapshot.Lookup.Account != nil && snapshot.Lookup.Account.Account != nil {
				return strings.TrimSpace(snapshot.Lookup.Account.Account.ID)
			}
		case app.LookupAsset:
			if snapshot.Lookup.Asset != nil && snapshot.Lookup.Asset.Asset != nil {
				asset := snapshot.Lookup.Asset.Asset
				return strings.TrimSpace(asset.AssetCode + ":" + asset.AssetIssuer)
			}
		case app.LookupContract:
			if snapshot.Lookup.Contract != nil && snapshot.Lookup.Contract.Contract != nil {
				return strings.TrimSpace(snapshot.Lookup.Contract.Contract.ContractID)
			}
		case app.LookupEvent:
			if snapshot.Lookup.Event != nil {
				return strings.TrimSpace(snapshot.Lookup.Event.ParentContractID)
			}
		case app.LookupStorage:
			if snapshot.Lookup.StorageEntry != nil {
				return strings.TrimSpace(snapshot.Lookup.StorageEntry.ParentContractID)
			}
		}
		return strings.TrimSpace(snapshot.Lookup.Query)
	}
	return ""
}

func selectionStatusMessage(snapshot app.Snapshot, dashboard *ui.DashboardModel) string {
	if dashboard == nil {
		return ""
	}
	switch snapshot.Current {
	case app.ScreenLiveFeed:
		selected := dashboard.SelectedLiveTransaction()
		if selected == nil {
			return ""
		}
		return fmt.Sprintf("Selected live transaction %d: %s.", selected.Index+1, selected.Hash)
	case app.ScreenLookup:
		return strings.TrimSpace(dashboard.LookupActionHint())
	default:
		return ""
	}
}

func truncateStatusValue(value string) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= 48 {
		return value
	}
	runes := []rune(value)
	return string(runes[:48]) + "..."
}

func mergeLocalSearchResults(base []app.SearchResult, results []app.SearchResult) []app.SearchResult {
	merged := append([]app.SearchResult(nil), base...)
	seen := make(map[string]struct{}, len(merged)+len(results))
	for _, result := range merged {
		seen[searchResultDedupKey(result)] = struct{}{}
	}
	for _, result := range results {
		key := searchResultDedupKey(result)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, result)
	}
	return app.RankSearchResults("", merged)
}

func searchResultDedupKey(result app.SearchResult) string {
	if command := strings.TrimSpace(result.Command); command != "" {
		return command
	}
	return strings.TrimSpace(result.Kind) + ":" + strings.TrimSpace(result.Title) + ":" + strings.TrimSpace(result.Description)
}

func activateSelection(ctx context.Context, cfg config.Config, model *app.Model, dashboard *ui.DashboardModel, store *cache.Store) error {
	snapshot := model.Snapshot()

	switch snapshot.Current {
	case app.ScreenLiveFeed:
		if dashboard != nil {
			selectedIndex, scrollOffset := dashboard.LiveFeedSelection()
			model.CaptureLiveFeedMonitoringContext(selectedIndex, scrollOffset)
		}
		var selected *app.TransactionSummarySelection
		if dashboard != nil {
			selected = dashboard.SelectedLiveTransaction()
		}
		if selected == nil {
			current := model.SelectedLiveTransaction()
			if current != nil {
				selected = &app.TransactionSummarySelection{
					Hash:       current.Hash,
					Ledger:     current.LedgerSequence,
					Account:    current.Account,
					Ops:        current.OperationCount,
					HasAccount: strings.TrimSpace(current.Account) != "",
					IsSoroban:  current.IsSoroban,
					CreatedAt:  current.CreatedAt,
					StatusCode: current.Status,
				}
			}
		}
		if selected == nil {
			return nil
		}
		backend, err := initializeLookupBackend(cfg)
		if err != nil {
			return err
		}
		if backend == nil {
			model.SetLookupError("", "", fmt.Errorf("lookup backend is unavailable"), app.DefaultSourceMetadata(model.Snapshot().Profile, "transaction"))
			return nil
		}

		response, err := backend.Transaction(ctx, selected.Hash)
		if err != nil {
			model.SetLookupError(app.LookupTransaction, selected.Hash, err, sourceMetadataFor(model.Snapshot().Profile, string(app.LookupTransaction), backend))
			return nil
		}
		model.UpdateLookupTransaction(selected.Hash, response, sourceMetadataFor(model.Snapshot().Profile, string(app.LookupTransaction), backend))
		model.SetLookupReturnContext(app.ScreenLiveFeed, "live feed")
		afterLookupLoaded(ctx, model, store, string(app.LookupTransaction), selected.Hash, response)
	case app.ScreenLookup:
		command := ""
		if dashboard != nil {
			command = strings.TrimSpace(dashboard.LookupActionCommand())
		}
		if command == "" {
			return nil
		}
		keepRunning, err := executeCommand(ctx, cfg, model, command, store)
		if err != nil {
			return err
		}
		if !keepRunning {
			return nil
		}
	}

	return nil
}

type interactiveKey int

const (
	keyUnknown interactiveKey = iota
	keyQuit
	keyHome
	keyLiveFeed
	keyLookup
	keySettings
	keyBack
	keyForward
	keyHelp
	keyRefresh
	keySearch
	keyFocusNext
	keyMoveDown
	keyMoveUp
	keyEnter
	keyBackspace
	keyEscape
)

func normalizeInteractiveKey(input []byte) interactiveKey {
	switch string(input) {
	case "q", "\x03":
		return keyQuit
	case "h":
		return keyHome
	case "l":
		return keyLiveFeed
	case "u":
		return keyLookup
	case "s":
		return keySettings
	case "b":
		return keyBack
	case "f":
		return keyForward
	case "?":
		return keyHelp
	case "r":
		return keyRefresh
	case "/":
		return keySearch
	case "\t":
		return keyFocusNext
	case "j", "\x1b[B":
		return keyMoveDown
	case "k", "\x1b[A":
		return keyMoveUp
	case "\r", "\n":
		return keyEnter
	case "\x7f", "\b":
		return keyBackspace
	case "\x1b":
		return keyEscape
	default:
		return keyUnknown
	}
}

func printableInput(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	value := string(input)
	for _, r := range value {
		if r < 32 || r == 127 {
			return ""
		}
	}
	return value
}

func isPrintableTeaKey(key string) bool {
	if key == "space" {
		return true
	}
	if len(key) != 1 {
		return false
	}
	for _, r := range key {
		return r >= 32 && r != 127
	}
	return false
}
