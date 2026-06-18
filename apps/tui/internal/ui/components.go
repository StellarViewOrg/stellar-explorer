package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
)

// DashboardModel is the root Bubble Tea view model for the terminal shell.
type DashboardModel struct {
	snapshot app.Snapshot
	width    int
	height   int
	sidebar  SidebarModel
	home     HomeModel
	liveFeed LiveFeedModel
	lookup   LookupModel
	settings SettingsModel
	command  CommandPaletteModel
}

func NewDashboardModel(snapshot app.Snapshot, width, height int) DashboardModel {
	width = clamp(width, minWidth, 220)
	height = clamp(height, minHeight, 80)
	return DashboardModel{
		snapshot: snapshot,
		width:    width,
		height:   height,
		sidebar:  NewSidebarModel(snapshot, clamp(width/3, 27, 34), height),
		home:     NewHomeModel(snapshot, width, height),
		liveFeed: NewLiveFeedModel(snapshot, width, height),
		lookup:   NewLookupModel(snapshot, width, height),
		settings: NewSettingsModel(snapshot, width, height),
		command:  NewCommandPaletteModel(snapshot, width),
	}
}

func (m DashboardModel) WithSnapshot(snapshot app.Snapshot) DashboardModel {
	m.snapshot = snapshot
	m.sidebar = m.sidebar.WithSnapshot(snapshot)
	m.home = m.home.WithSnapshot(snapshot)
	m.liveFeed = m.liveFeed.WithSnapshot(snapshot)
	m.lookup = m.lookup.WithSnapshot(snapshot)
	m.settings = m.settings.WithSnapshot(snapshot)
	m.command = m.command.WithSnapshot(snapshot)
	return m
}

func (m *DashboardModel) OpenCommandPalette(prompt, seed string) {
	m.command = m.command.Open(prompt, seed)
}

func (m *DashboardModel) CloseCommandPalette() {
	m.command = m.command.Close()
}

func (m *DashboardModel) CommandInput() string {
	return m.command.Input()
}

func (m *DashboardModel) SelectedCommandPaletteResult() *app.SearchResult {
	return m.command.SelectedResult()
}

func (m *DashboardModel) SetCommandPaletteResults(results []app.SearchResult) {
	m.command = m.command.SetResults(results)
}

func (m *DashboardModel) AppendCommandPaletteInput(text string) {
	m.command = m.command.AppendInput(text)
}

func (m DashboardModel) LiveFeedClipboardValue() string {
	return m.liveFeed.ClipboardValue()
}

func (m DashboardModel) SelectedLiveTransaction() *app.TransactionSummarySelection {
	return m.liveFeed.SelectedTransaction()
}

func (m DashboardModel) LiveFeedSelection() (selected int, offset int) {
	return m.liveFeed.selected, m.liveFeed.offset
}

func (m DashboardModel) LookupClipboardValue() string {
	return m.lookup.ClipboardValue()
}

func (m DashboardModel) LookupActionCommand() string {
	return m.lookup.ActionCommand()
}

func (m DashboardModel) LookupActionHint() string {
	return m.lookup.ActionHint()
}

func (m DashboardModel) LookupSelection() (int, int) {
	return m.lookup.selectedSection, m.lookup.offset
}

func (m DashboardModel) FocusOrder() []app.FocusArea {
	if m.width < 92 {
		return []app.FocusArea{app.FocusMain, app.FocusStatus}
	}
	return []app.FocusArea{app.FocusMain, app.FocusSidebar, app.FocusStatus}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = clamp(msg.Width, minWidth, 220)
		m.height = clamp(msg.Height, minHeight, 80)
		m.sidebar.width = clamp(m.width/3, 27, 34)
		m.sidebar.height = m.height
		m.home.width = m.width
		m.home.height = m.height
		m.liveFeed.width = m.width
		m.liveFeed.height = m.height
		m.lookup.width = m.width
		m.lookup.height = m.height
		m.settings.width = m.width
		m.settings.height = m.height
		return m, nil
	case tea.KeyMsg:
		if m.command.visible {
			model, cmd := m.command.Update(msg)
			if command, ok := model.(CommandPaletteModel); ok {
				m.command = command
			}
			return m, cmd
		}
		updated, cmd, handled := m.componentKeyAction(msg)
		if handled {
			m = updated
			return m, cmd
		}
		if cmd := globalKeyAction(msg.String()); cmd != nil {
			return m, cmd
		}
		return m, nil
	}
	return m, nil
}

func (m DashboardModel) View() string {
	const (
		headerHeight = 3
		footerHeight = 1
		statusHeight = 4
	)
	contentHeight := clamp(m.height-headerHeight-statusHeight-footerHeight, 8, 64)

	parts := []string{
		NewHeaderModel(m.snapshot, m.width).View(),
		m.renderMain(contentHeight),
		NewStatusModel(m.snapshot, m.width, statusHeight).View(),
		NewFooterModel(m.snapshot, m.width).View(),
	}

	if m.snapshot.HelpVisible {
		parts = append(parts, NewHelpOverlayModel(m.snapshot, m.width).View())
	}
	if m.command.visible {
		command := m.command
		command.width = m.width
		parts = append(parts, command.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m DashboardModel) renderMain(height int) string {
	if m.width < 92 {
		sidebarHeight := clamp(height/3, 8, 12)
		bodyHeight := max(8, height-sidebarHeight-1)
		body := m.renderBody(m.width, bodyHeight)
		sidebar := m.sidebar.WithSnapshot(m.snapshot)
		sidebar.width = m.width
		sidebar.height = sidebarHeight
		return lipgloss.JoinVertical(lipgloss.Left, body, " ", sidebar.View())
	}

	sidebarWidth := clamp(m.width/3, 27, 34)
	bodyWidth := m.width - sidebarWidth - 1
	sidebar := m.sidebar.WithSnapshot(m.snapshot)
	sidebar.width = sidebarWidth
	sidebar.height = height
	body := m.renderBody(bodyWidth, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar.View(), " ", body)
}

func (m DashboardModel) renderBody(width, height int) string {
	switch m.snapshot.Current {
	case app.ScreenLiveFeed:
		liveFeed := m.liveFeed
		liveFeed = liveFeed.WithSnapshot(m.snapshot)
		liveFeed.width = width
		liveFeed.height = height
		return liveFeed.View()
	case app.ScreenLookup:
		lookup := m.lookup
		lookup = lookup.WithSnapshot(m.snapshot)
		lookup.width = width
		lookup.height = height
		return lookup.View()
	case app.ScreenSettings:
		settings := m.settings.WithSnapshot(m.snapshot)
		settings.width = width
		settings.height = height
		return settings.View()
	default:
		home := m.home.WithSnapshot(m.snapshot)
		home.width = width
		home.height = height
		return home.View()
	}
}

type HeaderModel struct {
	snapshot app.Snapshot
	width    int
}

func NewHeaderModel(snapshot app.Snapshot, width int) HeaderModel {
	return HeaderModel{snapshot: snapshot, width: width}
}

func (m HeaderModel) Init() tea.Cmd { return nil }

func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
	}
	return m, nil
}

func (m HeaderModel) View() string {
	return renderHeader(m.snapshot, m.width)
}

type MainModel struct {
	snapshot app.Snapshot
	width    int
	height   int
}

func NewMainModel(snapshot app.Snapshot, width, height int) MainModel {
	return MainModel{snapshot: snapshot, width: width, height: height}
}

func (m MainModel) Init() tea.Cmd { return nil }

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
	}
	return m, nil
}

func (m MainModel) View() string {
	return renderMain(m.snapshot, m.width, m.height)
}

type SidebarModel struct {
	snapshot app.Snapshot
	width    int
	height   int
	selected app.Screen
}

func NewSidebarModel(snapshot app.Snapshot, width, height int) SidebarModel {
	return SidebarModel{snapshot: snapshot, width: width, height: height, selected: snapshot.Current}
}

func (m SidebarModel) Init() tea.Cmd { return nil }

func (m SidebarModel) WithSnapshot(snapshot app.Snapshot) SidebarModel {
	m.snapshot = snapshot
	if !isSidebarScreen(m.selected) {
		m.selected = snapshot.Current
	}
	return m
}

func (m SidebarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			m.selected = nextSidebarScreen(m.selected, 1)
			return m, nil
		case "up", "k":
			m.selected = nextSidebarScreen(m.selected, -1)
			return m, nil
		case "enter":
			return m, sidebarAction(m.selected)
		}
	}
	return m, nil
}

func (m SidebarModel) View() string {
	selected := m.selected
	if !isSidebarScreen(selected) {
		selected = m.snapshot.Current
	}
	return renderSidebarPanel(m.snapshot, m.width, m.height, selected, m.snapshot.Focus == app.FocusSidebar)
}

type BodyModel struct {
	snapshot app.Snapshot
	width    int
	height   int
}

func NewBodyModel(snapshot app.Snapshot, width, height int) BodyModel {
	return BodyModel{snapshot: snapshot, width: width, height: height}
}

func (m BodyModel) Init() tea.Cmd { return nil }

func (m BodyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
	}
	return m, nil
}

func (m BodyModel) View() string {
	switch m.snapshot.Current {
	case app.ScreenLiveFeed:
		return NewLiveFeedModel(m.snapshot, m.width, m.height).View()
	case app.ScreenLookup:
		return NewLookupModel(m.snapshot, m.width, m.height).View()
	case app.ScreenSettings:
		return NewSettingsModel(m.snapshot, m.width, m.height).View()
	default:
		return NewHomeModel(m.snapshot, m.width, m.height).View()
	}
}

type LiveFeedModel struct {
	snapshot  app.Snapshot
	width     int
	height    int
	offset    int
	selected  int
	copyField liveFeedCopyField
}

func NewLiveFeedModel(snapshot app.Snapshot, width, height int) LiveFeedModel {
	return LiveFeedModel{
		snapshot:  snapshot,
		width:     width,
		height:    height,
		selected:  clampLiveSelection(snapshot.Selection.LiveFeedIndex, len(snapshot.LiveFeed.RecentTransactions)),
		copyField: liveFeedCopyHash,
	}.WithSnapshot(snapshot)
}

func (m LiveFeedModel) Init() tea.Cmd { return nil }

func (m LiveFeedModel) WithSnapshot(snapshot app.Snapshot) LiveFeedModel {
	m.snapshot = snapshot
	m.selected = clampLiveSelection(m.selected, len(snapshot.LiveFeed.RecentTransactions))
	if len(snapshot.LiveFeed.RecentTransactions) == 0 {
		m.selected = 0
		m.offset = 0
		return m
	}
	if snapshot.Selection.LiveFeedIndex > 0 {
		m.selected = clampLiveSelection(snapshot.Selection.LiveFeedIndex, len(snapshot.LiveFeed.RecentTransactions))
	}
	if snapshot.Selection.LiveFeedScrollOffset > 0 {
		m.offset = snapshot.Selection.LiveFeedScrollOffset
	}
	m.offset = clampLiveFeedWindowOffset(m.offset, m.selected, m.height-2, len(snapshot.LiveFeed.RecentTransactions))
	return m
}

func (m LiveFeedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			m.moveSelection(1)
			return m, actionDelta(ActionMoveSelection, 1)
		case "up", "k":
			m.moveSelection(-1)
			return m, actionDelta(ActionMoveSelection, -1)
		case "pgdown", "ctrl+f":
			m.pageSelection(1)
			return m, actionDelta(ActionMoveSelection, 1)
		case "pgup", "ctrl+b":
			m.pageSelection(-1)
			return m, actionDelta(ActionMoveSelection, -1)
		case "home", "g":
			m.jumpSelection(true)
			return m, actionDelta(ActionMoveSelection, -1)
		case "end", "G":
			m.jumpSelection(false)
			return m, actionDelta(ActionMoveSelection, 1)
		case "left":
			m.copyField = m.copyField.prev()
			return m, nil
		case "right":
			m.copyField = m.copyField.next()
			return m, nil
		case "enter":
			return m, action(ActionActivateSelection)
		case "r":
			return m, action(ActionRefresh)
		case "p":
			return m, action(ActionToggleLivePause)
		case "t":
			return m, action(ActionCycleLiveFilter)
		case "[", "{":
			return m, actionDelta(ActionShiftLiveReplay, 1)
		case "]", "}":
			return m, actionDelta(ActionShiftLiveReplay, -1)
		}
	}
	return m, nil
}

func (m LiveFeedModel) View() string {
	return styledPanel("Live Feed", m.width, m.height, liveFeedView(m.snapshot, panelContentWidth(m.width), m.height-2, m.offset, m.selected, m.copyField), m.snapshot.Focus == app.FocusMain)
}

func (m *LiveFeedModel) moveSelection(delta int) {
	count := len(m.snapshot.LiveFeed.RecentTransactions)
	if count == 0 || delta == 0 {
		return
	}
	m.selected = clampLiveSelection(m.selected+delta, count)
	m.offset = clampLiveFeedWindowOffset(m.offset, m.selected, m.height-2, count)
}

func (m *LiveFeedModel) pageSelection(delta int) {
	count := len(m.snapshot.LiveFeed.RecentTransactions)
	if count == 0 || delta == 0 {
		return
	}
	step := max(1, liveFeedVisibleRows(m.height-2)-1)
	m.selected = clampLiveSelection(m.selected+(step*delta), count)
	m.offset = clampLiveFeedWindowOffset(m.offset, m.selected, m.height-2, count)
}

func (m *LiveFeedModel) jumpSelection(toStart bool) {
	count := len(m.snapshot.LiveFeed.RecentTransactions)
	if count == 0 {
		return
	}
	if toStart {
		m.selected = 0
	} else {
		m.selected = count - 1
	}
	m.offset = clampLiveFeedWindowOffset(m.offset, m.selected, m.height-2, count)
}

func (m LiveFeedModel) ClipboardValue() string {
	index := m.selected
	if index < 0 || index >= len(m.snapshot.LiveFeed.RecentTransactions) {
		return ""
	}
	tx := m.snapshot.LiveFeed.RecentTransactions[index]
	switch m.copyField {
	case liveFeedCopyLedger:
		return fmt.Sprintf("%d", tx.LedgerSequence)
	case liveFeedCopyAccount:
		return tx.Account
	default:
		return tx.Hash
	}
}

func (m LiveFeedModel) SelectedTransaction() *app.TransactionSummarySelection {
	index := m.selected
	if index < 0 || index >= len(m.snapshot.LiveFeed.RecentTransactions) {
		return nil
	}
	tx := m.snapshot.LiveFeed.RecentTransactions[index]
	return &app.TransactionSummarySelection{
		Index:      index,
		Hash:       tx.Hash,
		Ledger:     tx.LedgerSequence,
		Account:    tx.Account,
		Ops:        tx.OperationCount,
		HasAccount: tx.Account != "",
		IsSoroban:  tx.IsSoroban,
		CreatedAt:  tx.CreatedAt,
		StatusCode: tx.Status,
	}
}

type LookupModel struct {
	snapshot        app.Snapshot
	width           int
	height          int
	offset          int
	selectedSection int
}

func NewLookupModel(snapshot app.Snapshot, width, height int) LookupModel {
	return LookupModel{snapshot: snapshot, width: width, height: height}.WithSnapshot(snapshot)
}

func (m LookupModel) Init() tea.Cmd { return nil }

func (m LookupModel) WithSnapshot(snapshot app.Snapshot) LookupModel {
	m.snapshot = snapshot
	sections := lookupSections(snapshot.Lookup)
	if len(sections) == 0 {
		m.offset = 0
		m.selectedSection = 0
		return m
	}
	m.selectedSection = snapshot.Lookup.SelectedSection
	m.offset = snapshot.Lookup.ScrollOffset
	m.selectedSection = nextSelectableLookupSection(sections, clamp(m.selectedSection, 0, len(sections)-1), 0, m.snapshot.Lookup.VisualMode)
	m.offset = clamp(m.offset, 0, max(0, len(sections)-lookupVisibleRows(m.height-2, snapshot.Lookup)))
	m.ensureSelectionVisible(len(sections))
	return m
}

func (m LookupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			m.moveSelection(1)
			return m, nil
		case "up", "k":
			m.moveSelection(-1)
			return m, nil
		case "pgdown", "ctrl+f":
			m.pageSelection(1)
			return m, nil
		case "pgup", "ctrl+b":
			m.pageSelection(-1)
			return m, nil
		case "home", "g":
			m.jumpSelection(true)
			return m, nil
		case "end", "G":
			m.jumpSelection(false)
			return m, nil
		case "enter":
			return m, action(ActionActivateSelection)
		case "tab":
			if app.LookupSupportsContractWorkspaceTabs(m.snapshot.Lookup) {
				return m, actionDelta(ActionCycleContractTab, 1)
			}
			return m, nil
		case "shift+tab":
			if app.LookupSupportsContractWorkspaceTabs(m.snapshot.Lookup) {
				return m, actionDelta(ActionCycleContractTab, -1)
			}
			return m, nil
		case "c":
			if app.LookupSupportsDecodeShortcuts(m.snapshot.Lookup) {
				return m, action(ActionSetContractDecodeRaw)
			}
			return m, nil
		case "d":
			if app.LookupSupportsDecodeShortcuts(m.snapshot.Lookup) {
				return m, action(ActionSetContractDecodeDecoded)
			}
			return m, nil
		case "e":
			if m.snapshot.Lookup.State == app.ViewStateReady && m.snapshot.Lookup.Explorer == nil {
				return m, action(ActionToggleLookupExpand)
			}
			return m, nil
		case "v":
			if app.LookupSupportsVisualMode(m.snapshot.Lookup) {
				return m, action(ActionToggleLookupVisual)
			}
			return m, nil
		case "esc":
			if m.snapshot.Lookup.VisualMode {
				return m, action(ActionToggleLookupVisual)
			}
			return m, nil
		case "m":
			return m, action(ActionMarkBookmark)
		case ",":
			return m, action(ActionQuickBookmark)
		case "-":
			return m, action(ActionRemoveBookmark)
		case ";":
			return m, actionText(ActionOpenNotePalette, "note add ")
		case ".":
			return m, actionText(ActionOpenLabelPalette, "label add ")
		}
	}
	return m, nil
}

func (m LookupModel) View() string {
	return styledPanel("Lookup", m.width, m.height, lookupView(m.snapshot, panelContentWidth(m.width), m.height-2, m.offset, m.selectedSection), m.snapshot.Focus == app.FocusMain)
}

func (m LookupModel) ClipboardValue() string {
	return selectedLookupClipboardValue(m.snapshot, m.selectedSection)
}

func (m LookupModel) ActionCommand() string {
	return selectedLookupActionCommand(m.snapshot, m.selectedSection)
}

func (m LookupModel) ActionHint() string {
	return selectedLookupActionHint(m.snapshot, m.selectedSection)
}

func (m *LookupModel) moveSelection(delta int) {
	sections := lookupSections(m.snapshot.Lookup)
	if len(sections) == 0 || delta == 0 {
		return
	}
	m.selectedSection = nextSelectableLookupSection(sections, m.selectedSection, delta, m.snapshot.Lookup.VisualMode)
	m.ensureSelectionVisible(len(sections))
}

func (m *LookupModel) pageSelection(delta int) {
	sections := lookupSections(m.snapshot.Lookup)
	if len(sections) == 0 || delta == 0 {
		return
	}
	visible := max(1, lookupVisibleRows(m.height-2, m.snapshot.Lookup)-1)
	step := visible * delta
	target := clamp(m.selectedSection+step, 0, len(sections)-1)
	if delta > 0 {
		m.selectedSection = nextSelectableLookupSection(sections, target, 1, m.snapshot.Lookup.VisualMode)
	} else {
		m.selectedSection = nextSelectableLookupSection(sections, target, -1, m.snapshot.Lookup.VisualMode)
	}
	m.ensureSelectionVisible(len(sections))
}

func (m *LookupModel) jumpSelection(toStart bool) {
	sections := lookupSections(m.snapshot.Lookup)
	if len(sections) == 0 {
		return
	}
	if toStart {
		m.selectedSection = firstSelectableLookupSection(sections, m.snapshot.Lookup.VisualMode)
	} else {
		m.selectedSection = lastSelectableLookupSection(sections, m.snapshot.Lookup.VisualMode)
	}
	m.ensureSelectionVisible(len(sections))
}

func (m *LookupModel) ensureSelectionVisible(sectionCount int) {
	if sectionCount == 0 {
		m.offset = 0
		return
	}
	visible := lookupVisibleRows(m.height-2, m.snapshot.Lookup)
	if m.selectedSection < m.offset {
		m.offset = m.selectedSection
	}
	if m.selectedSection >= m.offset+visible {
		m.offset = m.selectedSection - visible + 1
	}
	m.offset = clamp(m.offset, 0, max(0, sectionCount-visible))
}

type HomeModel struct {
	snapshot app.Snapshot
	width    int
	height   int
	offset   int
}

func NewHomeModel(snapshot app.Snapshot, width, height int) HomeModel {
	return HomeModel{snapshot: snapshot, width: width, height: height}
}

func (m HomeModel) Init() tea.Cmd { return nil }

func (m HomeModel) WithSnapshot(snapshot app.Snapshot) HomeModel {
	m.snapshot = snapshot
	m.offset = clampScrollOffset(m.offset, len(homeView(snapshot, panelContentWidth(m.width))), m.height-2)
	return m
}

func (m HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			m.offset++
		case "up", "k":
			m.offset--
		case "pgdown", "ctrl+f":
			m.offset += max(1, m.height-4)
		case "pgup", "ctrl+b":
			m.offset -= max(1, m.height-4)
		case "home", "g":
			m.offset = 0
		case "end", "G":
			m.offset = len(homeView(m.snapshot, panelContentWidth(m.width)))
		}
		m.offset = clampScrollOffset(m.offset, len(homeView(m.snapshot, panelContentWidth(m.width))), m.height-2)
	}
	return m, nil
}

func (m HomeModel) View() string {
	return styledPanel("Home", m.width, m.height, windowLines(homeView(m.snapshot, panelContentWidth(m.width)), m.offset, m.height-2), m.snapshot.Focus == app.FocusMain)
}

type SettingsModel struct {
	snapshot app.Snapshot
	width    int
	height   int
	offset   int
}

func NewSettingsModel(snapshot app.Snapshot, width, height int) SettingsModel {
	return SettingsModel{snapshot: snapshot, width: width, height: height}
}

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m SettingsModel) WithSnapshot(snapshot app.Snapshot) SettingsModel {
	m.snapshot = snapshot
	m.offset = clampScrollOffset(m.offset, len(settingsView(snapshot, panelContentWidth(m.width))), m.height-2)
	return m
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			m.offset++
		case "up", "k":
			m.offset--
		case "pgdown", "ctrl+f":
			m.offset += max(1, m.height-4)
		case "pgup", "ctrl+b":
			m.offset -= max(1, m.height-4)
		case "home", "g":
			m.offset = 0
		case "end", "G":
			m.offset = len(settingsView(m.snapshot, panelContentWidth(m.width)))
		}
		m.offset = clampScrollOffset(m.offset, len(settingsView(m.snapshot, panelContentWidth(m.width))), m.height-2)
	}
	return m, nil
}

func (m SettingsModel) View() string {
	return styledPanel("Settings", m.width, m.height, windowLines(settingsView(m.snapshot, panelContentWidth(m.width)), m.offset, m.height-2), m.snapshot.Focus == app.FocusMain)
}

type StatusModel struct {
	snapshot app.Snapshot
	width    int
	height   int
}

func NewStatusModel(snapshot app.Snapshot, width, height int) StatusModel {
	return StatusModel{snapshot: snapshot, width: width, height: height}
}

func (m StatusModel) Init() tea.Cmd { return nil }

func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
	}
	return m, nil
}

func (m StatusModel) View() string {
	return renderStatus(m.snapshot, m.width, m.height)
}

type FooterModel struct {
	snapshot app.Snapshot
	width    int
}

func NewFooterModel(snapshot app.Snapshot, width int) FooterModel {
	return FooterModel{snapshot: snapshot, width: width}
}

func (m FooterModel) Init() tea.Cmd { return nil }

func (m FooterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
	}
	return m, nil
}

func (m FooterModel) View() string {
	return renderFooter(m.snapshot, m.width)
}

type HelpOverlayModel struct {
	snapshot app.Snapshot
	width    int
}

func NewHelpOverlayModel(snapshot app.Snapshot, width int) HelpOverlayModel {
	return HelpOverlayModel{snapshot: snapshot, width: width}
}

func (m HelpOverlayModel) Init() tea.Cmd { return nil }

func (m HelpOverlayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
	}
	return m, nil
}

func (m HelpOverlayModel) View() string {
	return renderHelpOverlay(m.snapshot, m.width)
}

type CommandPaletteModel struct {
	snapshot      app.Snapshot
	width         int
	visible       bool
	prompt        string
	input         string
	results       []app.SearchResult
	selectedIndex int
}

func NewCommandPaletteModel(snapshot app.Snapshot, width int) CommandPaletteModel {
	return CommandPaletteModel{
		snapshot:      snapshot,
		width:         width,
		visible:       snapshot.Command.Visible,
		prompt:        snapshot.Command.Prompt,
		input:         snapshot.Command.Input,
		results:       append([]app.SearchResult(nil), snapshot.Command.Results...),
		selectedIndex: snapshot.Command.SelectedIndex,
	}
}

func (m CommandPaletteModel) Init() tea.Cmd { return nil }

func (m CommandPaletteModel) WithSnapshot(snapshot app.Snapshot) CommandPaletteModel {
	m.snapshot = snapshot
	if !m.visible && snapshot.Command.Visible {
		return NewCommandPaletteModel(snapshot, m.width)
	}
	return m
}

func (m CommandPaletteModel) Open(prompt, seed string) CommandPaletteModel {
	m.visible = true
	m.prompt = prompt
	m.input = seed
	m.results = app.InferSearchResults(seed)
	m.selectedIndex = 0
	return m
}

func (m CommandPaletteModel) Close() CommandPaletteModel {
	m.visible = false
	m.prompt = ""
	m.input = ""
	m.results = nil
	m.selectedIndex = 0
	return m
}

func (m CommandPaletteModel) Input() string {
	return m.input
}

func (m CommandPaletteModel) SelectedResult() *app.SearchResult {
	if len(m.results) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.results) {
		return nil
	}
	result := m.results[m.selectedIndex]
	return &result
}

func (m CommandPaletteModel) SetResults(results []app.SearchResult) CommandPaletteModel {
	m.results = append([]app.SearchResult(nil), results...)
	if m.selectedIndex >= len(m.results) {
		m.selectedIndex = clamp(len(m.results)-1, 0, max(0, len(m.results)-1))
	}
	return m
}

func (m CommandPaletteModel) AppendInput(text string) CommandPaletteModel {
	if text == "" {
		return m
	}
	m.input += text
	m.results = app.InferSearchResults(m.input)
	if len(m.results) > 0 {
		m.selectedIndex = clamp(m.selectedIndex, 0, len(m.results)-1)
	} else {
		m.selectedIndex = 0
	}
	return m
}

func (m CommandPaletteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, action(ActionQuit)
		case "esc":
			return m, action(ActionCloseCommand)
		case "enter":
			return m, action(ActionSubmitCommand)
		case "backspace":
			runes := []rune(m.input)
			if len(runes) > 0 {
				m.input = string(runes[:len(runes)-1])
			} else {
				m.input = ""
			}
			m.results = app.InferSearchResults(m.input)
			if m.selectedIndex >= len(m.results) {
				m.selectedIndex = clamp(len(m.results)-1, 0, max(0, len(m.results)-1))
			}
			return m, action(ActionCommandBackspace)
		case "down", "ctrl+n":
			if len(m.results) > 0 {
				m.selectedIndex = clamp(m.selectedIndex+1, 0, len(m.results)-1)
			}
			return m, actionDelta(ActionMoveCommandResult, 1)
		case "up", "ctrl+p":
			if len(m.results) > 0 {
				m.selectedIndex = clamp(m.selectedIndex-1, 0, len(m.results)-1)
			}
			return m, actionDelta(ActionMoveCommandResult, -1)
		case "ctrl+v":
			return m, action(ActionPasteClipboard)
		default:
			if text := printableTeaKey(msg.String()); text != "" {
				m = m.AppendInput(text)
				return m, actionText(ActionAppendCommandInput, text)
			}
		}
	}
	return m, nil
}

func (m CommandPaletteModel) View() string {
	return renderCommandOverlayState(m.width, m.prompt, m.input, m.results, m.selectedIndex)
}

func (m DashboardModel) componentKeyAction(msg tea.KeyMsg) (DashboardModel, tea.Cmd, bool) {
	if m.snapshot.Focus == app.FocusSidebar {
		model, cmd := m.sidebar.Update(msg)
		if sidebar, ok := model.(SidebarModel); ok {
			m.sidebar = sidebar
		}
		switch msg.String() {
		case "down", "j", "up", "k", "enter":
			return m, cmd, true
		default:
			return m, cmd, cmd != nil
		}
	}
	if m.snapshot.Focus != app.FocusMain {
		return m, nil, false
	}

	switch m.snapshot.Current {
	case app.ScreenHome:
		model, cmd := m.home.Update(msg)
		if home, ok := model.(HomeModel); ok {
			m.home = home
		}
		switch msg.String() {
		case "down", "j", "up", "k", "pgdown", "ctrl+f", "pgup", "ctrl+b", "home", "g", "end", "G":
			return m, cmd, true
		default:
			return m, cmd, cmd != nil
		}
	case app.ScreenLiveFeed:
		model, cmd := m.liveFeed.Update(msg)
		if liveFeed, ok := model.(LiveFeedModel); ok {
			m.liveFeed = liveFeed
		}
		switch msg.String() {
		case "down", "j", "up", "k", "pgdown", "ctrl+f", "pgup", "ctrl+b", "home", "g", "end", "G", "enter", "left", "right", "[", "{", "]", "}":
			return m, cmd, true
		default:
			return m, cmd, cmd != nil
		}
	case app.ScreenLookup:
		model, cmd := m.lookup.Update(msg)
		if lookup, ok := model.(LookupModel); ok {
			m.lookup = lookup
		}
		switch msg.String() {
		case "down", "j", "up", "k", "pgdown", "ctrl+f", "pgup", "ctrl+b", "home", "g", "end", "G", "enter", "m", ",", "-", ";", ".":
			return m, cmd, true
		case "tab", "shift+tab", "c", "d", "e", "v", "esc":
			return m, cmd, cmd != nil
		default:
			return m, cmd, cmd != nil
		}
	case app.ScreenSettings:
		model, cmd := m.settings.Update(msg)
		if settings, ok := model.(SettingsModel); ok {
			m.settings = settings
		}
		switch msg.String() {
		case "down", "j", "up", "k", "pgdown", "ctrl+f", "pgup", "ctrl+b", "home", "g", "end", "G":
			return m, cmd, true
		default:
			return m, cmd, cmd != nil
		}
	default:
		return m, nil, false
	}
}

func clampLiveFeedWindowOffset(offset, selected, height, count int) int {
	if count == 0 {
		return 0
	}
	visibleRows := liveFeedVisibleRows(height)
	if selected < offset {
		offset = selected
	}
	if selected >= offset+visibleRows {
		offset = selected - visibleRows + 1
	}
	return clamp(offset, 0, max(0, count-visibleRows))
}

func clampScrollOffset(offset, total, height int) int {
	if total <= 0 {
		return 0
	}
	window := max(1, height)
	return clamp(offset, 0, max(0, total-window))
}

func windowLines(lines []string, offset, height int) []string {
	if len(lines) == 0 {
		return nil
	}
	offset = clampScrollOffset(offset, len(lines), height)
	end := min(len(lines), offset+max(1, height))
	return lines[offset:end]
}

func clampLiveSelection(value, count int) int {
	if count <= 0 {
		return 0
	}
	return clamp(value, 0, count-1)
}

func nextSidebarScreen(current app.Screen, delta int) app.Screen {
	if len(sidebarScreens) == 0 {
		return current
	}
	index := 0
	for i, screen := range sidebarScreens {
		if screen == current {
			index = i
			break
		}
	}
	index = clamp(index+delta, 0, len(sidebarScreens)-1)
	return sidebarScreens[index]
}

func isSidebarScreen(screen app.Screen) bool {
	for _, candidate := range sidebarScreens {
		if candidate == screen {
			return true
		}
	}
	return false
}

func sidebarAction(screen app.Screen) tea.Cmd {
	switch screen {
	case app.ScreenLiveFeed:
		return action(ActionLiveFeed)
	case app.ScreenLookup:
		return action(ActionLookup)
	case app.ScreenSettings:
		return action(ActionSettings)
	default:
		return action(ActionHome)
	}
}

func lookupVisibleRows(height int, lookup app.LookupSnapshot) int {
	baseLines := 10
	return max(1, height-baseLines)
}
