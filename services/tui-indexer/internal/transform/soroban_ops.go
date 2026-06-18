package transform

import (
	"fmt"
	"strings"

	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

func enrichInvokeHostFunction(storeOp *store.Operation, invoke xdr.InvokeHostFunctionOp, details map[string]interface{}) {
	hostFn := invoke.HostFunction
	details["host_function_type"] = hostFn.Type.String()

	switch hostFn.Type {
	case xdr.HostFunctionTypeHostFunctionTypeInvokeContract:
		args := hostFn.MustInvokeContract()
		contractID := sordecode.ScAddressToString(args.ContractAddress)
		functionName := strings.TrimSpace(string(args.FunctionName))
		if contractID != "" && contractID != "unknown_address" {
			details["contract_id"] = contractID
			storeOp.ContractID = &contractID
		}
		if functionName != "" {
			details["function_name"] = functionName
			storeOp.FunctionName = &functionName
		}
		if len(args.Args) > 0 {
			argSummaries := make([]string, 0, len(args.Args))
			for index, arg := range args.Args {
				argSummaries = append(argSummaries, fmt.Sprintf("arg%d=%s", index+1, truncateScValSummary(arg, 48)))
			}
			details["arguments"] = argSummaries
		}
	case xdr.HostFunctionTypeHostFunctionTypeCreateContract, xdr.HostFunctionTypeHostFunctionTypeCreateContractV2:
		details["host_action"] = "create_contract"
	case xdr.HostFunctionTypeHostFunctionTypeUploadContractWasm:
		if wasm, ok := hostFn.GetWasm(); ok {
			details["wasm_bytes"] = len(wasm)
		}
	}

	if authorizations := summarizeSorobanAuthorizations(invoke.Auth); len(authorizations) > 0 {
		details["auth_count"] = len(authorizations)
		details["authorizations"] = authorizations
	}
}

func summarizeSorobanAuthorizations(entries []xdr.SorobanAuthorizationEntry) []string {
	if len(entries) == 0 {
		return nil
	}
	summaries := make([]string, 0, len(entries))
	for index, entry := range entries {
		summaries = append(summaries, fmt.Sprintf("auth %d: %s", index+1, summarizeAuthorizationEntry(entry)))
	}
	return summaries
}

func summarizeAuthorizationEntry(entry xdr.SorobanAuthorizationEntry) string {
	var credential string
	switch entry.Credentials.Type {
	case xdr.SorobanCredentialsTypeSorobanCredentialsSourceAccount:
		credential = "source_account"
	case xdr.SorobanCredentialsTypeSorobanCredentialsAddress:
		creds := entry.Credentials.MustAddress()
		parts := []string{"address:" + sordecode.ScAddressToString(creds.Address)}
		if creds.Nonce != 0 {
			parts = append(parts, fmt.Sprintf("nonce=%d", creds.Nonce))
		}
		if creds.SignatureExpirationLedger != 0 {
			parts = append(parts, fmt.Sprintf("sig_exp=%d", creds.SignatureExpirationLedger))
		}
		credential = strings.Join(parts, " ")
	default:
		credential = entry.Credentials.Type.String()
	}

	if fn := authorizedFunctionName(entry.RootInvocation.Function); fn != "" {
		return credential + " fn=" + fn
	}
	return credential
}

func authorizedFunctionName(fn xdr.SorobanAuthorizedFunction) string {
	if fn.Type == xdr.SorobanAuthorizedFunctionTypeSorobanAuthorizedFunctionTypeContractFn {
		return strings.TrimSpace(string(fn.MustContractFn().FunctionName))
	}
	return ""
}

func truncateScValSummary(value xdr.ScVal, max int) string {
	summary := scValToString(value)
	if max <= 0 || len(summary) <= max {
		return summary
	}
	return summary[:max] + "..."
}
