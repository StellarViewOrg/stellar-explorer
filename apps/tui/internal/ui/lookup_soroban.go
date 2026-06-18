package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

func lookupEventDetails(lookup app.LookupSnapshot) []lookupSection {
	if lookup.Event == nil {
		return []lookupSection{{Title: "Empty", Body: "No event payload available.", Muted: true}}
	}

	event := lookup.Event.Event
	mode := lookup.DecodeMode
	if mode == "" {
		mode = app.ContractDecodeModeDecoded
	}

	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Summary"),
		{Title: "Contract", Body: lookup.Event.ParentContractID, Command: "lookup contract " + lookup.Event.ParentContractID, Hint: "enter open parent contract"},
		{Title: "Transaction", Body: event.TransactionHash, Copy: event.TransactionHash, Command: "lookup tx " + event.TransactionHash, Hint: "enter open parent transaction"},
		{Title: "Ledger", Body: fmt.Sprintf("%d", event.LedgerSequence)},
		{Title: "Type", Body: fmt.Sprintf("%d", event.Type)},
		{Title: "Created", Body: renderTimestamp(event.CreatedAt)},
	}
	if strings.TrimSpace(event.DecodeStatus) != "" {
		lines = append(lines, lookupSection{Title: "Decode", Body: strings.TrimSpace(event.DecodeStatus)})
	}
	if strings.TrimSpace(event.SpecDecodeStatus) != "" {
		lines = append(lines, lookupSection{Title: "Spec Decode", Body: strings.TrimSpace(event.SpecDecodeStatus)})
	}
	if strings.TrimSpace(event.EventName) != "" {
		lines = append(lines, lookupSection{Title: "Event", Body: strings.TrimSpace(event.EventName), Emph: true})
	}
	if len(event.FieldsDecoded) > 0 {
		lines = append(lines, sectionHeader(fmt.Sprintf("Fields (%d)", len(event.FieldsDecoded))))
		for index, field := range event.FieldsDecoded {
			title := field.Name
			if title == "" {
				title = fmt.Sprintf("Field %d", index+1)
			}
			body := truncate(strings.TrimSpace(field.Value), 72)
			if strings.TrimSpace(field.Type) != "" {
				body = truncate(fmt.Sprintf("%s (%s)", field.Value, field.Type), 72)
			}
			lines = append(lines, lookupSection{
				Title: title,
				Body:  body,
				Copy:  strings.TrimSpace(field.Value),
				Hint:  "copy decoded field",
			})
		}
	}
	if strings.TrimSpace(event.Summary) != "" && mode != app.ContractDecodeModeRaw {
		lines = append(lines, lookupSection{Title: "Summary Text", Body: truncate(strings.TrimSpace(event.Summary), 72)})
	}

	lines = append(lines, sectionHeader("Topics"))
	if mode == app.ContractDecodeModeRaw {
		if event.TopicsXDR != nil && strings.TrimSpace(*event.TopicsXDR) != "" {
			lines = append(lines, lookupSection{Title: "Topics XDR", Body: truncate(strings.TrimSpace(*event.TopicsXDR), 72), Copy: strings.TrimSpace(*event.TopicsXDR), Hint: "copy topics xdr"})
		} else {
			lines = append(lines, lookupSection{Title: "Topics", Body: "No raw topics payload available.", Muted: true})
		}
	} else if len(event.Topics) > 0 {
		for index, topic := range event.Topics {
			lines = append(lines, lookupSection{Title: fmt.Sprintf("Topic %d", index+1), Body: truncate(strings.TrimSpace(topic), 72), Copy: strings.TrimSpace(topic)})
		}
	} else {
		for index, topic := range []*string{event.Topic1, event.Topic2, event.Topic3, event.Topic4} {
			if topic == nil || strings.TrimSpace(*topic) == "" {
				continue
			}
			lines = append(lines, lookupSection{Title: fmt.Sprintf("Topic %d", index+1), Body: truncate(strings.TrimSpace(*topic), 72), Copy: strings.TrimSpace(*topic)})
		}
		if !hasDecodedEventTopics(event) {
			lines = append(lines, lookupSection{Title: "Topics", Body: "No decoded topics available.", Muted: true})
		}
	}

	lines = append(lines, sectionHeader("Value"))
	if mode == app.ContractDecodeModeRaw {
		if event.ValueXDR != nil && strings.TrimSpace(*event.ValueXDR) != "" {
			lines = append(lines, lookupSection{Title: "Value XDR", Body: truncate(strings.TrimSpace(*event.ValueXDR), 72), Copy: strings.TrimSpace(*event.ValueXDR), Hint: "copy value xdr"})
		} else {
			lines = append(lines, lookupSection{Title: "Value", Body: "No raw value payload available.", Muted: true})
		}
	} else if event.ValueDecoded != nil && strings.TrimSpace(*event.ValueDecoded) != "" {
		lines = append(lines, lookupSection{Title: "Decoded", Body: truncate(strings.TrimSpace(*event.ValueDecoded), 72), Copy: strings.TrimSpace(*event.ValueDecoded), Hint: "copy decoded value"})
	} else if event.ValueXDR != nil && strings.TrimSpace(*event.ValueXDR) != "" {
		lines = append(lines, lookupSection{Title: "Raw Fallback", Body: truncate(strings.TrimSpace(*event.ValueXDR), 72), Copy: strings.TrimSpace(*event.ValueXDR), Hint: "copy raw value"})
	} else {
		lines = append(lines, lookupSection{Title: "Value", Body: "No decoded value available.", Muted: true})
	}

	if mode != app.ContractDecodeModeRaw {
		if event.TopicsXDR != nil && strings.TrimSpace(*event.TopicsXDR) != "" {
			lines = append(lines, sectionHeader("Raw Fallback"), lookupSection{Title: "Topics XDR", Body: truncate(strings.TrimSpace(*event.TopicsXDR), 72), Copy: strings.TrimSpace(*event.TopicsXDR), Hint: "copy topics xdr"})
		}
		if event.ValueXDR != nil && strings.TrimSpace(*event.ValueXDR) != "" {
			lines = append(lines, lookupSection{Title: "Value XDR", Body: truncate(strings.TrimSpace(*event.ValueXDR), 72), Copy: strings.TrimSpace(*event.ValueXDR), Hint: "copy value xdr"})
		}
	}

	return lines
}

func lookupStorageEntryDetails(lookup app.LookupSnapshot) []lookupSection {
	if lookup.StorageEntry == nil {
		return []lookupSection{{Title: "Empty", Body: "No storage payload available.", Muted: true}}
	}

	entry := lookup.StorageEntry.Entry
	mode := lookup.DecodeMode
	if mode == "" {
		mode = app.ContractDecodeModeDecoded
	}

	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Summary"),
		{Title: "Contract", Body: lookup.StorageEntry.ParentContractID, Command: "lookup contract " + lookup.StorageEntry.ParentContractID, Hint: "enter open parent contract"},
		{Title: "Durability", Body: strings.TrimSpace(entry.DurabilityLabel)},
		{Title: "Modified", Body: fmt.Sprintf("ledger %d", entry.LastModifiedLedger)},
		{Title: "Updated", Body: renderTimestamp(entry.UpdatedAt)},
	}
	if strings.TrimSpace(entry.DecodeStatus) != "" {
		lines = append(lines, lookupSection{Title: "Decode", Body: strings.TrimSpace(entry.DecodeStatus)})
	}
	if entry.TTLLedger != nil {
		lines = append(lines, lookupSection{Title: "TTL Ledger", Body: fmt.Sprintf("%d", *entry.TTLLedger)})
	}
	if strings.TrimSpace(entry.ExpirationProximity) != "" {
		lines = append(lines, lookupSection{Title: "Expiration", Body: strings.TrimSpace(entry.ExpirationProximity)})
	}

	lines = append(lines, sectionHeader("Key"))
	if mode == app.ContractDecodeModeRaw || strings.TrimSpace(entry.DisplayKey) == "" {
		lines = append(lines, lookupSection{Title: "Key XDR", Body: truncate(strings.TrimSpace(entry.KeyXDR), 72), Copy: strings.TrimSpace(entry.KeyXDR), Hint: "copy key xdr"})
	} else {
		lines = append(lines, lookupSection{Title: "Decoded", Body: truncate(strings.TrimSpace(entry.DisplayKey), 72), Copy: strings.TrimSpace(entry.DisplayKey), Hint: "copy decoded key"})
		if strings.TrimSpace(entry.KeyXDR) != "" {
			lines = append(lines, lookupSection{Title: "Key XDR", Body: truncate(strings.TrimSpace(entry.KeyXDR), 72), Copy: strings.TrimSpace(entry.KeyXDR), Hint: "copy key xdr"})
		}
	}

	lines = append(lines, sectionHeader("Value"))
	if mode == app.ContractDecodeModeRaw {
		lines = append(lines, lookupSection{Title: "Value XDR", Body: truncate(strings.TrimSpace(entry.ValueXDR), 72), Copy: strings.TrimSpace(entry.ValueXDR), Hint: "copy value xdr"})
	} else if strings.TrimSpace(entry.DisplayValue) != "" {
		lines = append(lines, lookupSection{Title: "Decoded", Body: truncate(strings.TrimSpace(entry.DisplayValue), 72), Copy: contractStorageCopyValue(entry), Hint: "copy decoded value"})
		if strings.TrimSpace(entry.ValueXDR) != "" {
			lines = append(lines, lookupSection{Title: "Value XDR", Body: truncate(strings.TrimSpace(entry.ValueXDR), 72), Copy: strings.TrimSpace(entry.ValueXDR), Hint: "copy value xdr"})
		}
	} else if entry.ValueDecoded != nil && strings.TrimSpace(*entry.ValueDecoded) != "" {
		lines = append(lines, lookupSection{Title: "Decoded", Body: truncate(strings.TrimSpace(*entry.ValueDecoded), 72), Copy: strings.TrimSpace(*entry.ValueDecoded), Hint: "copy decoded value"})
	} else {
		lines = append(lines, lookupSection{Title: "Value XDR", Body: truncate(strings.TrimSpace(entry.ValueXDR), 72), Copy: strings.TrimSpace(entry.ValueXDR), Hint: "copy value xdr"})
	}

	return lines
}

func appendSorobanOperationSections(lines []lookupSection, op backendclient.OperationSummary, mode app.ContractDecodeMode, displayLimit int) []lookupSection {
	if strings.TrimSpace(op.TypeName) != "invoke_host_function" {
		return lines
	}
	if displayLimit <= 0 {
		displayLimit = 72
	}
	if mode == "" {
		mode = app.ContractDecodeModeDecoded
	}

	details := parseOperationDetailsJSON(op.Details)
	if len(details) == 0 && op.FunctionName == nil && op.ContractID == nil {
		return lines
	}

	lines = append(lines, sectionHeader("Soroban Invocation"))
	if op.FunctionName != nil && strings.TrimSpace(*op.FunctionName) != "" {
		lines = append(lines, lookupSection{Title: "Function", Body: strings.TrimSpace(*op.FunctionName), Emph: true})
	}
	if hostType, ok := details["host_function_type"].(string); ok && strings.TrimSpace(hostType) != "" {
		lines = append(lines, lookupSection{Title: "Host Fn", Body: strings.TrimSpace(hostType)})
	}
	if op.ContractID != nil && strings.TrimSpace(*op.ContractID) != "" {
		lines = append(lines, relatedEntityRow("Contract", strings.TrimSpace(*op.ContractID), strings.TrimSpace(*op.ContractID), "lookup contract "+strings.TrimSpace(*op.ContractID), "enter open invoked contract"))
	}
	showDecoded := mode != app.ContractDecodeModeRaw
	if showDecoded {
		if decodedArgs, ok := details["arguments_decoded"].([]interface{}); ok && len(decodedArgs) > 0 {
			lines = append(lines, sectionHeader(fmt.Sprintf("Arguments (%d)", len(decodedArgs))))
			for index, raw := range decodedArgs {
				entry, ok := raw.(map[string]interface{})
				if !ok {
					lines = append(lines, lookupSection{Title: fmt.Sprintf("Arg %d", index+1), Body: truncate(fmt.Sprint(raw), displayLimit)})
					continue
				}
				name := strings.TrimSpace(fmt.Sprint(entry["name"]))
				if name == "" {
					name = fmt.Sprintf("arg%d", index+1)
				}
				value := strings.TrimSpace(fmt.Sprint(entry["value"]))
				typeName := strings.TrimSpace(fmt.Sprint(entry["type"]))
				body := value
				if typeName != "" && typeName != "<nil>" {
					body = fmt.Sprintf("%s (%s)", value, typeName)
				}
				lines = append(lines, lookupSection{
					Title: name,
					Body:  truncate(body, displayLimit),
					Copy:  value,
					Hint:  "copy decoded argument",
				})
			}
		}
		if result, ok := details["result_decoded"].(string); ok && strings.TrimSpace(result) != "" {
			lines = append(lines, sectionHeader("Result"), lookupSection{
				Title: "Decoded",
				Body:  truncate(strings.TrimSpace(result), displayLimit),
				Copy:  strings.TrimSpace(result),
				Hint:  "copy decoded result",
			})
		}
	}
	if args, ok := details["arguments"].([]interface{}); ok && len(args) > 0 && (!showDecoded || details["arguments_decoded"] == nil) {
		lines = append(lines, sectionHeader(fmt.Sprintf("Arguments (%d)", len(args))))
		for index, arg := range args {
			lines = append(lines, lookupSection{Title: fmt.Sprintf("Arg %d", index+1), Body: truncate(fmt.Sprint(arg), displayLimit)})
		}
	}
	if subInvocations, ok := details["sub_invocations"].([]interface{}); ok && len(subInvocations) > 0 && showDecoded {
		lines = append(lines, sectionHeader(fmt.Sprintf("Sub Invocations (%d)", len(subInvocations))))
		for index, raw := range subInvocations {
			lines = append(lines, lookupSection{
				Title: fmt.Sprintf("Invocation %d", index+1),
				Body:  truncate(summarizeSubInvocation(raw), displayLimit),
			})
		}
	}
	if status, ok := details["spec_decode_status"].(string); ok && strings.TrimSpace(status) != "" && showDecoded {
		lines = append(lines, lookupSection{Title: "Spec Decode", Body: strings.TrimSpace(status)})
	}
	if authCount, ok := details["auth_count"].(float64); ok && int(authCount) > 0 {
		lines = append(lines, sectionHeader(fmt.Sprintf("Authorization (%d)", int(authCount))))
		if authorizations, ok := details["authorizations"].([]interface{}); ok {
			for index, auth := range authorizations {
				lines = append(lines, lookupSection{Title: fmt.Sprintf("Auth %d", index+1), Body: truncate(fmt.Sprint(auth), displayLimit)})
			}
		}
	}
	if wasmBytes, ok := details["wasm_bytes"].(float64); ok && int(wasmBytes) > 0 {
		lines = append(lines, lookupSection{Title: "WASM Upload", Body: fmt.Sprintf("%d bytes", int(wasmBytes))})
	}
	if hostAction, ok := details["host_action"].(string); ok && strings.TrimSpace(hostAction) != "" {
		lines = append(lines, lookupSection{Title: "Host Action", Body: strings.TrimSpace(hostAction)})
	}

	return lines
}

func parseOperationDetailsJSON(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var details map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &details); err != nil {
		return nil
	}
	return details
}

func hasDecodedEventTopics(event backendclient.ContractEventSummary) bool {
	if len(event.Topics) > 0 {
		return true
	}
	for _, topic := range []*string{event.Topic1, event.Topic2, event.Topic3, event.Topic4} {
		if topic != nil && strings.TrimSpace(*topic) != "" {
			return true
		}
	}
	return false
}

func eventOpenCommand(index int) string {
	return fmt.Sprintf("open event %d", index+1)
}

func storageOpenCommand(index int) string {
	return fmt.Sprintf("open storage %d", index+1)
}

func invocationOpenCommand(index int) string {
	return fmt.Sprintf("open invocation %d", index+1)
}

func summarizeSubInvocation(raw interface{}) string {
	entry, ok := raw.(map[string]interface{})
	if !ok {
		return fmt.Sprint(raw)
	}
	parts := make([]string, 0, 4)
	if contractID := strings.TrimSpace(fmt.Sprint(entry["contract_id"])); contractID != "" && contractID != "<nil>" {
		parts = append(parts, contractID)
	}
	if function := strings.TrimSpace(fmt.Sprint(entry["function"])); function != "" && function != "<nil>" {
		parts = append(parts, function+"()")
	}
	if hostAction := strings.TrimSpace(fmt.Sprint(entry["host_action"])); hostAction != "" && hostAction != "<nil>" {
		parts = append(parts, hostAction)
	}
	if params, ok := entry["params"].([]interface{}); ok && len(params) > 0 {
		parts = append(parts, fmt.Sprintf("%d params", len(params)))
	}
	if len(parts) == 0 {
		return "sub invocation"
	}
	return strings.Join(parts, " ")
}
