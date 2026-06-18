package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

const (
	specDecodeStatusDecoded  = "decoded"
	specDecodeStatusPartial  = "partial"
	specDecodeStatusRaw      = "raw"
	specDecodeStatusNoSpec   = "spec_unavailable"
	specDecodeStatusNoInvoke = "not_applicable"
)

type decodedEventPayload struct {
	Text             string                 `json:"text,omitempty"`
	EventName        string                 `json:"event_name,omitempty"`
	Fields           []sordecode.NamedField `json:"fields,omitempty"`
	SpecDecodeStatus string                 `json:"spec_decode_status,omitempty"`
}

// EnrichSorobanOperations applies contract spec XDR decoding to invoke_host_function operation details.
func EnrichSorobanOperations(ctx context.Context, loader ContractSpecRegistryLoader, ops []store.Operation, txEntry source.TransactionEntry) error {
	if loader == nil || len(ops) == 0 {
		return nil
	}

	returnValues, err := sorobanReturnValuesFromTransactionMeta(txEntry.ResultMetaXDR)
	if err != nil {
		return err
	}
	returnIndex := 0

	for index := range ops {
		if ops[index].TypeName != "invoke_host_function" {
			continue
		}
		details, err := parseOperationDetailsMap(ops[index].Details)
		if err != nil {
			continue
		}

		contractID := strings.TrimSpace(stringFromDetails(details, "contract_id"))
		if contractID == "" && ops[index].ContractID != nil {
			contractID = strings.TrimSpace(*ops[index].ContractID)
		}
		functionName := strings.TrimSpace(stringFromDetails(details, "function_name"))
		if functionName == "" && ops[index].FunctionName != nil {
			functionName = strings.TrimSpace(*ops[index].FunctionName)
		}

		var registry *sordecode.SpecRegistry
		if contractID != "" {
			registry, _ = loader.GetSpecRegistryForContract(ctx, contractID)
		}

		status := specDecodeStatusNoSpec
		if registry != nil && functionName != "" {
			if invoke, args, ok := invokeContractArgsFromDetails(details, txEntry, ops[index].ApplicationOrder); ok {
				decoded, decodeErr := registry.DecodeInvocation(functionName, args)
				if decodeErr == nil && len(decoded.Params) > 0 {
					details["arguments_decoded"] = decoded.Params
					status = specDecodeStatusDecoded
				}
				_ = invoke
			} else if args := argsFromDetailStrings(details); len(args) > 0 {
				decoded, decodeErr := registry.DecodeInvocation(functionName, args)
				if decodeErr == nil {
					details["arguments_decoded"] = decoded.Params
					status = specDecodeStatusPartial
				}
			}
		}

		if registry != nil && functionName != "" && returnIndex < len(returnValues) && returnValues[returnIndex] != nil {
			if result, decodeErr := registry.DecodeInvocationResult(functionName, *returnValues[returnIndex]); decodeErr == nil {
				details["result_decoded"] = result
				if status == specDecodeStatusNoSpec {
					status = specDecodeStatusPartial
				} else if status != specDecodeStatusDecoded {
					status = specDecodeStatusPartial
				}
			}
			returnIndex++
		}

		if authEntries, ok := authEntriesFromTx(txEntry, ops[index].ApplicationOrder); ok && len(authEntries) > 0 {
			details["sub_invocations"] = registryAwareAuthTree(authEntries, registry)
			if registry != nil && status == specDecodeStatusNoSpec {
				status = specDecodeStatusPartial
			}
		}

		details["spec_decode_status"] = status
		if encoded, err := json.Marshal(details); err == nil {
			ops[index].Details = string(encoded)
		}
	}

	return nil
}

// EnrichContractEvents applies contract spec XDR decoding to indexed contract event payloads.
func EnrichContractEvents(ctx context.Context, loader ContractSpecRegistryLoader, events []store.ContractEvent) error {
	if loader == nil || len(events) == 0 {
		return nil
	}

	for index := range events {
		registry, _ := loader.GetSpecRegistryForContract(ctx, events[index].ContractID)
		if registry == nil {
			continue
		}

		topics, topicErr := decodeTopicValsFromStoredEvent(events[index])
		value, valueErr := decodeValueValFromStoredEvent(events[index])
		if topicErr != nil || valueErr != nil {
			continue
		}

		decoded, err := registry.DecodeEvent(topics, value)
		if err != nil {
			continue
		}

		payload := decodedEventPayload{
			Text:             sordecode.ScValToString(value),
			EventName:        decoded.Name,
			Fields:           decoded.Fields,
			SpecDecodeStatus: specDecodeStatusForEvent(decoded),
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			continue
		}
		encodedStr := string(encoded)
		events[index].ValueDecoded = &encodedStr
		if decoded.Name != "" {
			events[index].Topic1 = &decoded.Name
		}
	}

	return nil
}

func registryAwareAuthTree(entries []xdr.SorobanAuthorizationEntry, registry *sordecode.SpecRegistry) []sordecode.DecodedSubInvocation {
	if registry == nil {
		return sordecode.NewRegistry(nil).DecodeAuthorizationTree(entries)
	}
	return registry.DecodeAuthorizationTree(entries)
}

func specDecodeStatusForEvent(decoded sordecode.DecodedEvent) string {
	if decoded.Name == "" || len(decoded.Fields) == 0 {
		return specDecodeStatusRaw
	}
	for _, field := range decoded.Fields {
		if field.Value == "" || field.Value == "<missing>" {
			return specDecodeStatusPartial
		}
	}
	return specDecodeStatusDecoded
}

func parseOperationDetailsMap(raw string) (map[string]interface{}, error) {
	details := map[string]interface{}{}
	if strings.TrimSpace(raw) == "" {
		return details, nil
	}
	if err := json.Unmarshal([]byte(raw), &details); err != nil {
		return nil, err
	}
	return details, nil
}

func stringFromDetails(details map[string]interface{}, key string) string {
	value, ok := details[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func invokeContractArgsFromDetails(details map[string]interface{}, txEntry source.TransactionEntry, applicationOrder int32) (xdr.InvokeHostFunctionOp, []xdr.ScVal, bool) {
	_, ops, ok := transactionEnvelopeOps(txEntry.EnvelopeXDR)
	if !ok || applicationOrder <= 0 || int(applicationOrder) > len(ops) {
		return xdr.InvokeHostFunctionOp{}, nil, false
	}
	op := ops[applicationOrder-1]
	if op.Body.Type != xdr.OperationTypeInvokeHostFunction {
		return xdr.InvokeHostFunctionOp{}, nil, false
	}
	invoke := op.Body.MustInvokeHostFunctionOp()
	if invoke.HostFunction.Type != xdr.HostFunctionTypeHostFunctionTypeInvokeContract {
		return xdr.InvokeHostFunctionOp{}, nil, false
	}
	args := invoke.HostFunction.MustInvokeContract()
	return invoke, args.Args, true
}

func transactionEnvelopeOps(envelopeXDR string) (xdr.TransactionEnvelope, []xdr.Operation, bool) {
	if strings.TrimSpace(envelopeXDR) == "" {
		return xdr.TransactionEnvelope{}, nil, false
	}
	var envelope xdr.TransactionEnvelope
	if err := xdr.SafeUnmarshalBase64(envelopeXDR, &envelope); err != nil {
		return xdr.TransactionEnvelope{}, nil, false
	}
	if envelope.Type == xdr.EnvelopeTypeEnvelopeTypeTx && envelope.V1 != nil {
		return envelope, envelope.V1.Tx.Operations, true
	}
	if envelope.Type == xdr.EnvelopeTypeEnvelopeTypeTxV0 && envelope.V0 != nil {
		return envelope, envelope.V0.Tx.Operations, true
	}
	return envelope, nil, false
}

func argsFromDetailStrings(details map[string]interface{}) []xdr.ScVal {
	rawArgs, ok := details["arguments"].([]interface{})
	if !ok || len(rawArgs) == 0 {
		return nil
	}
	args := make([]xdr.ScVal, 0, len(rawArgs))
	for _, rawArg := range rawArgs {
		text := strings.TrimSpace(fmt.Sprint(rawArg))
		if eq := strings.Index(text, "="); eq > 0 {
			text = strings.TrimSpace(text[eq+1:])
			if strings.HasSuffix(text, "...") {
				text = strings.TrimSuffix(text, "...")
			}
		}
		args = append(args, xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: scSymbolPtr(text)})
	}
	return args
}

func scSymbolPtr(value string) *xdr.ScSymbol {
	sym := xdr.ScSymbol(value)
	return &sym
}

func authEntriesFromTx(txEntry source.TransactionEntry, applicationOrder int32) ([]xdr.SorobanAuthorizationEntry, bool) {
	_, ops, ok := transactionEnvelopeOps(txEntry.EnvelopeXDR)
	if !ok || applicationOrder <= 0 || int(applicationOrder) > len(ops) {
		return nil, false
	}
	op := ops[applicationOrder-1]
	if op.Body.Type != xdr.OperationTypeInvokeHostFunction {
		return nil, false
	}
	invoke := op.Body.MustInvokeHostFunctionOp()
	return invoke.Auth, true
}

func decodeTopicValsFromStoredEvent(event store.ContractEvent) ([]xdr.ScVal, error) {
	if strings.TrimSpace(event.TopicsXDR) == "" {
		return nil, fmt.Errorf("missing topics xdr")
	}
	var topicXDRs []string
	if err := json.Unmarshal([]byte(event.TopicsXDR), &topicXDRs); err != nil {
		return nil, err
	}
	topics := make([]xdr.ScVal, 0, len(topicXDRs))
	for _, topicXDR := range topicXDRs {
		var topic xdr.ScVal
		if err := xdr.SafeUnmarshalBase64(topicXDR, &topic); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, nil
}

func decodeValueValFromStoredEvent(event store.ContractEvent) (xdr.ScVal, error) {
	if strings.TrimSpace(event.ValueXDR) == "" {
		return xdr.ScVal{}, fmt.Errorf("missing value xdr")
	}
	var value xdr.ScVal
	if err := xdr.SafeUnmarshalBase64(event.ValueXDR, &value); err != nil {
		return xdr.ScVal{}, err
	}
	return value, nil
}
