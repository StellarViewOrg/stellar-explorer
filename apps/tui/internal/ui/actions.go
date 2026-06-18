package ui

import tea "github.com/charmbracelet/bubbletea"

type ActionKind string

const (
	ActionNone                     ActionKind = ""
	ActionQuit                     ActionKind = "quit"
	ActionHome                     ActionKind = "home"
	ActionLiveFeed                 ActionKind = "live-feed"
	ActionLookup                   ActionKind = "lookup"
	ActionSettings                 ActionKind = "settings"
	ActionBack                     ActionKind = "back"
	ActionForward                  ActionKind = "forward"
	ActionToggleHelp               ActionKind = "toggle-help"
	ActionRefresh                  ActionKind = "refresh"
	ActionOpenCommandPalette       ActionKind = "open-command-palette"
	ActionFocusNext                ActionKind = "focus-next"
	ActionMoveSelection            ActionKind = "move-selection"
	ActionActivateSelection        ActionKind = "activate-selection"
	ActionCloseCommand             ActionKind = "close-command"
	ActionSubmitCommand            ActionKind = "submit-command"
	ActionCommandBackspace         ActionKind = "command-backspace"
	ActionMoveCommandResult        ActionKind = "move-command-result"
	ActionAppendCommandInput       ActionKind = "append-command-input"
	ActionPasteClipboard           ActionKind = "paste-clipboard"
	ActionCopySelection            ActionKind = "copy-selection"
	ActionToggleLivePause          ActionKind = "toggle-live-pause"
	ActionCycleLiveFilter          ActionKind = "cycle-live-filter"
	ActionShiftLiveReplay          ActionKind = "shift-live-replay"
	ActionMarkBookmark             ActionKind = "mark-bookmark"
	ActionQuickBookmark            ActionKind = "quick-bookmark"
	ActionRemoveBookmark           ActionKind = "remove-bookmark"
	ActionOpenNotePalette          ActionKind = "open-note-palette"
	ActionOpenLabelPalette         ActionKind = "open-label-palette"
	ActionCycleContractTab         ActionKind = "cycle-contract-tab"
	ActionSetContractDecodeRaw     ActionKind = "set-contract-decode-raw"
	ActionSetContractDecodeDecoded ActionKind = "set-contract-decode-decoded"
	ActionToggleLookupExpand       ActionKind = "toggle-lookup-expand"
	ActionToggleLookupVisual       ActionKind = "toggle-lookup-visual"
)

type ActionMsg struct {
	Kind  ActionKind
	Text  string
	Delta int
}

func action(kind ActionKind) tea.Cmd {
	return func() tea.Msg {
		return ActionMsg{Kind: kind}
	}
}

func actionText(kind ActionKind, text string) tea.Cmd {
	return func() tea.Msg {
		return ActionMsg{Kind: kind, Text: text}
	}
}

func actionDelta(kind ActionKind, delta int) tea.Cmd {
	return func() tea.Msg {
		return ActionMsg{Kind: kind, Delta: delta}
	}
}

func globalKeyAction(key string) tea.Cmd {
	switch key {
	case "ctrl+c", "q":
		return action(ActionQuit)
	case "h":
		return action(ActionHome)
	case "l":
		return action(ActionLiveFeed)
	case "u":
		return action(ActionLookup)
	case "s":
		return action(ActionSettings)
	case "b":
		return action(ActionBack)
	case "f":
		return action(ActionForward)
	case "?":
		return action(ActionToggleHelp)
	case "r":
		return action(ActionRefresh)
	case "/":
		return action(ActionOpenCommandPalette)
	case "tab":
		return action(ActionFocusNext)
	case "y":
		return action(ActionCopySelection)
	default:
		return nil
	}
}

func printableTeaKey(key string) string {
	if key == "space" {
		return " "
	}
	if len(key) != 1 {
		return ""
	}
	for _, r := range key {
		if r < 32 || r == 127 {
			return ""
		}
	}
	return key
}
