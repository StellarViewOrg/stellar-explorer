package sordecode

import (
	"fmt"
	"strings"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// DecodedSubInvocation is one node in a Soroban authorization invocation tree.
type DecodedSubInvocation struct {
	ContractID     string                 `json:"contract_id,omitempty"`
	Function       string                 `json:"function,omitempty"`
	HostAction     string                 `json:"host_action,omitempty"`
	Params         []NamedParam           `json:"params,omitempty"`
	SubInvocations []DecodedSubInvocation `json:"sub_invocations,omitempty"`
}

// DecodeAuthorizationTree renders Soroban auth entries as a structured invocation tree.
func (r *SpecRegistry) DecodeAuthorizationTree(entries []xdr.SorobanAuthorizationEntry) []DecodedSubInvocation {
	if len(entries) == 0 {
		return nil
	}
	result := make([]DecodedSubInvocation, 0, len(entries))
	for _, entry := range entries {
		result = append(result, decodeAuthorizedInvocation(entry.RootInvocation, r))
	}
	return result
}

func decodeAuthorizedInvocation(invocation xdr.SorobanAuthorizedInvocation, registry *SpecRegistry) DecodedSubInvocation {
	node := decodeAuthorizedFunction(invocation.Function, registry)
	if len(invocation.SubInvocations) > 0 {
		node.SubInvocations = make([]DecodedSubInvocation, 0, len(invocation.SubInvocations))
		for _, sub := range invocation.SubInvocations {
			node.SubInvocations = append(node.SubInvocations, decodeAuthorizedInvocation(sub, registry))
		}
	}
	return node
}

func decodeAuthorizedFunction(function xdr.SorobanAuthorizedFunction, registry *SpecRegistry) DecodedSubInvocation {
	switch function.Type {
	case xdr.SorobanAuthorizedFunctionTypeSorobanAuthorizedFunctionTypeContractFn:
		args := function.MustContractFn()
		contractID := ScAddressToString(args.ContractAddress)
		functionName := strings.TrimSpace(string(args.FunctionName))
		node := DecodedSubInvocation{
			ContractID: contractID,
			Function:   functionName,
		}
		if registry != nil && functionName != "" {
			if decoded, err := registry.DecodeInvocation(functionName, args.Args); err == nil {
				node.Params = decoded.Params
			}
		}
		if len(node.Params) == 0 && len(args.Args) > 0 {
			node.Params = fallbackParams(args.Args)
		}
		return node
	case xdr.SorobanAuthorizedFunctionTypeSorobanAuthorizedFunctionTypeCreateContractHostFn,
		xdr.SorobanAuthorizedFunctionTypeSorobanAuthorizedFunctionTypeCreateContractV2HostFn:
		return DecodedSubInvocation{HostAction: "create_contract"}
	default:
		return DecodedSubInvocation{HostAction: function.Type.String()}
	}
}

func fallbackParams(args []xdr.ScVal) []NamedParam {
	params := make([]NamedParam, 0, len(args))
	for index, arg := range args {
		params = append(params, NamedParam{
			Name:  formatArgName(index),
			Type:  "Val",
			Value: ScValToString(arg),
		})
	}
	return params
}

func formatArgName(index int) string {
	return fmt.Sprintf("arg%d", index+1)
}
