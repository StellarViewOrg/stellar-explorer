package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// LabelAttachment describes one imported label for an entity target.
type LabelAttachment struct {
	Kind   string
	Target string
	Names  []string
}

type labelsFile struct {
	Accounts     map[string]labelEntry `toml:"accounts"`
	Addresses    map[string]labelEntry `toml:"addresses"`
	Transactions map[string]labelEntry `toml:"transactions"`
	Contracts    map[string]labelEntry `toml:"contracts"`
	Assets       map[string]labelEntry `toml:"assets"`
}

type labelEntry struct {
	Name string   `toml:"name"`
	Tags []string `toml:"tags"`
}

// LoadLabelsFile parses a snbeat-style labels.toml file.
func LoadLabelsFile(path string) ([]LabelAttachment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read labels file: %w", err)
	}

	var file labelsFile
	if err := toml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode labels file: %w", err)
	}

	attachments := make([]LabelAttachment, 0)
	attachments = append(attachments, parseLabelSection("account", file.Accounts)...)
	attachments = append(attachments, parseLabelSection("account", file.Addresses)...)
	attachments = append(attachments, parseLabelSection("transaction", file.Transactions)...)
	attachments = append(attachments, parseLabelSection("contract", file.Contracts)...)
	attachments = append(attachments, parseLabelSection("asset", file.Assets)...)

	return attachments, nil
}

func parseLabelSection(kind string, entries map[string]labelEntry) []LabelAttachment {
	if len(entries) == 0 {
		return nil
	}

	attachments := make([]LabelAttachment, 0, len(entries))
	for target, raw := range entries {
		names := labelNames(raw)
		if len(names) == 0 {
			continue
		}
		attachments = append(attachments, LabelAttachment{
			Kind:   kind,
			Target: target,
			Names:  names,
		})
	}

	return attachments
}

func labelNames(raw labelEntry) []string {
	if raw.Name != "" {
		names := []string{raw.Name}
		names = append(names, raw.Tags...)
		return uniqueStrings(names)
	}
	return uniqueStrings(raw.Tags)
}

// UnmarshalTOML supports both plain strings and structured label entries.
func (e *labelEntry) UnmarshalTOML(data any) error {
	switch value := data.(type) {
	case string:
		e.Name = value
		return nil
	case map[string]any:
		if name, ok := value["name"].(string); ok {
			e.Name = name
		}
		if tags, ok := value["tags"].([]any); ok {
			for _, tag := range tags {
				if text, ok := tag.(string); ok && text != "" {
					e.Tags = append(e.Tags, text)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported label entry type %T", data)
	}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
