package soroban

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

// BuildContractSpec converts parsed spec JSON into the TUI contract spec shape.
func BuildContractSpec(contractID string, rawJSON string) *backendclient.ContractSpec {
	spec := &backendclient.ContractSpec{
		ContractID:   contractID,
		Available:    false,
		DecodeStatus: "missing",
		Functions:    []backendclient.ContractSpecFunction{},
		Schemas:      []backendclient.ContractSpecSchema{},
		Events:       []backendclient.ContractSpecEvent{},
		UpdatedAt:    time.Now().UTC(),
	}
	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON == "" || rawJSON == "null" {
		return spec
	}

	spec.Available = true
	spec.DecodeStatus = "raw"
	spec.Raw = &rawJSON

	var entries []map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &entries); err != nil {
		return spec
	}

	for _, entry := range entries {
		kind := strings.ToLower(stringFromMap(entry, "kind"))
		switch {
		case strings.Contains(kind, "function"):
			spec.Functions = append(spec.Functions, backendclient.ContractSpecFunction{
				Name:    stringFromMap(entry, "name"),
				Doc:     optionalStringFromMap(entry, "doc"),
				Inputs:  contractSpecInputs(entry["inputs"]),
				Outputs: contractSpecOutputs(entry["outputs"]),
			})
		case strings.Contains(kind, "event"):
			spec.Events = append(spec.Events, contractSpecEventFromMap(entry))
		default:
			name := stringFromMap(entry, "name")
			if name == "" {
				continue
			}
			raw := marshalCompactJSON(entry)
			spec.Schemas = append(spec.Schemas, backendclient.ContractSpecSchema{
				Kind: kind,
				Name: name,
				Raw:  raw,
			})
		}
	}

	spec.FunctionCount = len(spec.Functions)
	spec.SchemaCount = len(spec.Schemas)
	spec.EventCount = len(spec.Events)
	spec.DecodeStatus = "decoded"
	return spec
}

func stringFromMap(entry map[string]interface{}, key string) string {
	value, ok := entry[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(strings.Trim(fmt.Sprint(value), `"`))
}

func optionalStringFromMap(entry map[string]interface{}, key string) *string {
	value := stringFromMap(entry, key)
	if value == "" {
		return nil
	}
	return &value
}

func contractSpecInputs(value interface{}) []backendclient.ContractSpecValue {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	inputs := make([]backendclient.ContractSpecValue, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		inputs = append(inputs, backendclient.ContractSpecValue{
			Name: stringFromMap(entry, "name"),
			Type: formatSpecType(entry["type"]),
		})
	}
	return inputs
}

func contractSpecOutputs(value interface{}) []string {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	outputs := make([]string, 0, len(items))
	for _, item := range items {
		outputs = append(outputs, formatSpecType(item))
	}
	return outputs
}

func contractSpecEventFromMap(entry map[string]interface{}) backendclient.ContractSpecEvent {
	event := backendclient.ContractSpecEvent{
		Name:         stringFromMap(entry, "name"),
		Doc:          optionalStringFromMap(entry, "doc"),
		DataFormat:   stringFromMap(entry, "data_format"),
		PrefixTopics: stringSliceFromMap(entry["prefix_topics"]),
		Params:       contractSpecInputs(entry["params"]),
	}
	return event
}

func stringSliceFromMap(value interface{}) []string {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text := strings.TrimSpace(fmt.Sprint(item))
		if text != "" {
			result = append(result, text)
		}
	}
	return result
}

func formatSpecType(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]interface{}:
		if primitive := stringFromMap(typed, "type"); primitive != "" {
			return primitive
		}
		if encoded := marshalCompactJSON(typed); encoded != nil {
			return *encoded
		}
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func marshalCompactJSON(value interface{}) *string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	result := string(encoded)
	return &result
}
