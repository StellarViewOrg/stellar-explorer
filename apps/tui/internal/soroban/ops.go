package soroban

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/sordecode"
	"github.com/stellar/go-stellar-sdk/xdr"
)

const (
	specDecodeStatusDecoded = "decoded"
	specDecodeStatusPartial = "partial"
	specDecodeStatusNoSpec  = "spec_unavailable"
)

type decodedEventPayload struct {
	Text             string                 `json:"text,omitempty"`
	EventName        string                 `json:"event_name,omitempty"`
	Fields           []sordecode.NamedField `json:"fields,omitempty"`
	SpecDecodeStatus string                 `json:"spec_decode_status,omitempty"`
}

// EnrichTransactionOperations applies spec-aware decoding to invoke_host_function operations.
func EnrichTransactionOperations(ctx context.Context, loader *SpecLoader, response *backendclient.TransactionLookupResponse) error {
	if loader == nil || response == nil || response.Transaction == nil {
		return nil
	}

	envelopeXDR := strings.TrimSpace(response.Transaction.EnvelopeXDR)
	resultMeta := ""
	if response.Transaction.ResultMetaXDR != nil {
		resultMeta = strings.TrimSpace(*response.Transaction.ResultMetaXDR)
	}
	returnValues, _ := sorobanReturnValuesFromTransactionMeta(resultMeta)
	returnIndex := 0

	for index := range response.Operations {
		op := &response.Operations[index]
		if strings.TrimSpace(op.TypeName) != "invoke_host_function" {
			continue
		}

		details := parseOperationDetails(op.Details)
		contractID := firstNonEmpty(
			stringFromDetails(details, "contract_id"),
			stringPtr(op.ContractID),
		)
		functionName := firstNonEmpty(
			stringFromDetails(details, "function_name"),
			stringPtr(op.FunctionName),
		)
		if contractID == "" || functionName == "" {
			if args, fn, ok := invokeContractArgsFromEnvelope(envelopeXDR, op.ApplicationOrder); ok {
				if functionName == "" {
					functionName = fn
				}
				if contractID == "" {
					contractID = stringPtr(op.ContractID)
				}
				if len(args) > 0 && details["arguments"] == nil {
					details["arguments"] = summarizeArgs(args)
				}
			}
		}
		if contractID != "" {
			details["contract_id"] = contractID
		}
		if functionName != "" {
			details["function_name"] = functionName
		}

		registry, _ := loader.Registry(ctx, contractID)
		status := specDecodeStatusNoSpec
		if registry != nil && functionName != "" {
			if args, _, ok := invokeContractArgsFromEnvelope(envelopeXDR, op.ApplicationOrder); ok && len(args) > 0 {
				if decoded, err := registry.DecodeInvocation(functionName, args); err == nil && len(decoded.Params) > 0 {
					details["arguments_decoded"] = decoded.Params
					status = specDecodeStatusDecoded
				}
			} else if args := argsFromDetailStrings(details); len(args) > 0 {
				if decoded, err := registry.DecodeInvocation(functionName, args); err == nil {
					details["arguments_decoded"] = decoded.Params
					status = specDecodeStatusPartial
				}
			}
		}

		if registry != nil && functionName != "" && returnIndex < len(returnValues) && returnValues[returnIndex] != nil {
			if result, err := registry.DecodeInvocationResult(functionName, *returnValues[returnIndex]); err == nil {
				details["result_decoded"] = result
				if status == specDecodeStatusNoSpec {
					status = specDecodeStatusPartial
				}
			}
			returnIndex++
		}

		if authEntries, ok := authEntriesFromEnvelope(envelopeXDR, op.ApplicationOrder); ok && len(authEntries) > 0 {
			if registry != nil {
				details["sub_invocations"] = registry.DecodeAuthorizationTree(authEntries)
			}
			details["auth_count"] = len(authEntries)
		}

		details["spec_decode_status"] = status
		if encoded, err := json.Marshal(details); err == nil {
			op.Details = string(encoded)
		}
	}

	return nil
}

func parseOperationDetails(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	details := map[string]interface{}{}
	if raw == "" {
		return details
	}
	_ = json.Unmarshal([]byte(raw), &details)
	return details
}

func stringFromDetails(details map[string]interface{}, key string) string {
	value, ok := details[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
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
		sym := xdr.ScSymbol(text)
		args = append(args, xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &sym})
	}
	return args
}

func summarizeArgs(args []xdr.ScVal) []string {
	summaries := make([]string, 0, len(args))
	for index, arg := range args {
		summaries = append(summaries, fmt.Sprintf("arg%d=%s", index+1, sordecode.ScValToString(arg)))
	}
	return summaries
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
