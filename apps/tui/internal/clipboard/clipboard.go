package clipboard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	commandContext = exec.CommandContext
	lookPath       = exec.LookPath
)

type commandSpec struct {
	name string
	args []string
}

// Read returns clipboard text using the first supported OS command.
func Read(ctx context.Context) (string, error) {
	output, _, err := runClipboardCommand(ctx, readCommands(runtime.GOOS), "")
	if err != nil {
		return "", err
	}
	return strings.TrimRight(output, "\r\n"), nil
}

// Write stores clipboard text using the first supported OS command.
func Write(ctx context.Context, value string) error {
	_, command, err := runClipboardCommand(ctx, writeCommands(runtime.GOOS), value)
	if err != nil {
		return err
	}
	if strings.TrimSpace(command.name) == "" {
		return errors.New("clipboard write command unavailable")
	}
	return nil
}

func readCommands(goos string) []commandSpec {
	switch goos {
	case "darwin":
		return []commandSpec{{name: "pbpaste"}}
	case "windows":
		return []commandSpec{{name: "powershell", args: []string{"-NoProfile", "-Command", "Get-Clipboard"}}}
	default:
		return []commandSpec{
			{name: "wl-paste", args: []string{"--no-newline"}},
			{name: "xclip", args: []string{"-selection", "clipboard", "-o"}},
			{name: "xsel", args: []string{"--clipboard", "--output"}},
		}
	}
}

func writeCommands(goos string) []commandSpec {
	switch goos {
	case "darwin":
		return []commandSpec{{name: "pbcopy"}}
	case "windows":
		return []commandSpec{{name: "clip"}}
	default:
		return []commandSpec{
			{name: "wl-copy"},
			{name: "xclip", args: []string{"-selection", "clipboard"}},
			{name: "xsel", args: []string{"--clipboard", "--input"}},
		}
	}
}

func runClipboardCommand(ctx context.Context, commands []commandSpec, stdin string) (string, commandSpec, error) {
	if len(commands) == 0 {
		return "", commandSpec{}, errors.New("clipboard command unavailable")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var lastErr error
	for _, candidate := range commands {
		if _, err := lookPath(candidate.name); err != nil {
			lastErr = err
			continue
		}

		cmd := commandContext(timeoutCtx, candidate.name, candidate.args...)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if stdin != "" {
			cmd.Stdin = strings.NewReader(stdin)
		}
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			lastErr = fmt.Errorf("%s: %w", candidate.name, err)
			continue
		}
		return stdout.String(), candidate, nil
	}

	if lastErr == nil {
		lastErr = errors.New("clipboard command unavailable")
	}
	return "", commandSpec{}, lastErr
}
