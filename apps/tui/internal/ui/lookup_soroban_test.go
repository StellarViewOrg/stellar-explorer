package ui

import (
	"strings"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

func TestLookupEventDetailsRendersTopicsAndValue(t *testing.T) {
	topic := "transfer"
	value := "100"
	sections := lookupEventDetails(app.LookupSnapshot{
		Kind:       app.LookupEvent,
		DecodeMode: app.ContractDecodeModeDecoded,
		Event: &backendclient.ContractEventLookupSnapshot{
			ParentContractID: "CCONTRACT",
			Event: backendclient.ContractEventSummary{
				TransactionHash: "tx-event-1",
				LedgerSequence:  88,
				Type:            1,
				Topic1:          &topic,
				ValueDecoded:    &value,
				DecodeStatus:    "decoded",
			},
		},
	})

	foundTopic := false
	foundValue := false
	for _, section := range sections {
		if section.Title == "Topic 1" && strings.Contains(section.Body, "transfer") {
			foundTopic = true
		}
		if section.Title == "Decoded" && strings.Contains(section.Body, "100") {
			foundValue = true
		}
	}
	if !foundTopic || !foundValue {
		t.Fatalf("expected decoded event topics and value, got %#v", sections)
	}
}

func TestLookupStorageEntryDetailsRendersDecodedKeyValue(t *testing.T) {
	sections := lookupStorageEntryDetails(app.LookupSnapshot{
		Kind:       app.LookupStorage,
		DecodeMode: app.ContractDecodeModeDecoded,
		StorageEntry: &backendclient.ContractStorageLookupSnapshot{
			ParentContractID: "CCONTRACT",
			Entry: backendclient.ContractStorageSummary{
				ContractID:      "CCONTRACT",
				DisplayKey:      "balance",
				DisplayValue:    "100",
				DurabilityLabel: "persistent",
				KeyXDR:          "key-xdr",
				ValueXDR:        "value-xdr",
			},
		},
	})

	foundKey := false
	foundValue := false
	for _, section := range sections {
		if section.Title == "Decoded" && strings.Contains(section.Body, "balance") {
			foundKey = true
		}
		if section.Title == "Decoded" && strings.Contains(section.Body, "100") {
			foundValue = true
		}
	}
	if !foundKey || !foundValue {
		t.Fatalf("expected decoded storage key and value, got %#v", sections)
	}
}

func TestAppendSorobanOperationSectionsRendersDecodedArguments(t *testing.T) {
	fn := "hello"
	contract := "CCONTRACT"
	sections := appendSorobanOperationSections(nil, backendclient.OperationSummary{
		TypeName:     "invoke_host_function",
		FunctionName: &fn,
		ContractID:   &contract,
		Details: `{
			"function_name":"hello",
			"arguments_decoded":[{"name":"to","type":"string","value":"World"}],
			"result_decoded":"Hello, World!",
			"spec_decode_status":"decoded"
		}`,
	}, app.ContractDecodeModeDecoded, 72)

	foundArg := false
	foundResult := false
	foundStatus := false
	for _, section := range sections {
		if section.Title == "to" && strings.Contains(section.Body, "World") {
			foundArg = true
		}
		if section.Title == "Decoded" && strings.Contains(section.Body, "Hello, World!") {
			foundResult = true
		}
		if section.Title == "Spec Decode" && section.Body == "decoded" {
			foundStatus = true
		}
	}
	if !foundArg || !foundResult || !foundStatus {
		t.Fatalf("expected decoded arguments, result, and spec status, got %#v", sections)
	}
}

func TestLookupEventDetailsRendersSpecDecodedFields(t *testing.T) {
	topic := "transfer"
	value := `{"text":"100","event_name":"transfer","fields":[{"name":"amount","type":"i128","location":"value","value":"100"}],"spec_decode_status":"decoded"}`
	sections := lookupEventDetails(app.LookupSnapshot{
		Kind:       app.LookupEvent,
		DecodeMode: app.ContractDecodeModeDecoded,
		Event: &backendclient.ContractEventLookupSnapshot{
			ParentContractID: "CCONTRACT",
			Event: backendclient.ContractEventSummary{
				TransactionHash:  "tx-event-2",
				LedgerSequence:   90,
				Type:             1,
				Topic1:           &topic,
				ValueDecoded:     &value,
				DecodeStatus:     "decoded",
				EventName:        "transfer",
				SpecDecodeStatus: "decoded",
				FieldsDecoded: []backendclient.ContractEventField{
					{Name: "amount", Type: "i128", Location: "value", Value: "100"},
				},
			},
		},
	})

	foundEvent := false
	foundField := false
	for _, section := range sections {
		if section.Title == "Event" && section.Body == "transfer" {
			foundEvent = true
		}
		if section.Title == "amount" && strings.Contains(section.Body, "100") {
			foundField = true
		}
	}
	if !foundEvent || !foundField {
		t.Fatalf("expected spec-decoded event fields, got %#v", sections)
	}
}

func TestAppendSorobanOperationSectionsRendersAuthorization(t *testing.T) {
	fn := "transfer"
	contract := "CCONTRACT"
	sections := appendSorobanOperationSections(nil, backendclient.OperationSummary{
		TypeName:     "invoke_host_function",
		FunctionName: &fn,
		ContractID:   &contract,
		Details:      `{"function_name":"transfer","auth_count":1,"authorizations":["auth 1: source_account"],"arguments":["arg1=100"]}`,
	}, app.ContractDecodeModeRaw, 72)

	foundAuth := false
	foundArgs := false
	for _, section := range sections {
		if section.Title == "Auth 1" {
			foundAuth = true
		}
		if section.Title == "Arg 1" {
			foundArgs = true
		}
	}
	if !foundAuth || !foundArgs {
		t.Fatalf("expected soroban authorization and argument sections, got %#v", sections)
	}
}
