package soroban

import (
	"fmt"
	"strings"

	"github.com/stellar/go-stellar-sdk/xdr"
)

func sorobanReturnValuesFromTransactionMeta(resultMetaXDR string) ([]*xdr.ScVal, error) {
	if strings.TrimSpace(resultMetaXDR) == "" {
		return nil, nil
	}
	var meta xdr.TransactionMeta
	if err := xdr.SafeUnmarshalBase64(resultMetaXDR, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal transaction meta: %w", err)
	}

	switch meta.V {
	case 3:
		if meta.V3 == nil || meta.V3.SorobanMeta == nil {
			return nil, nil
		}
		value := meta.V3.SorobanMeta.ReturnValue
		return []*xdr.ScVal{&value}, nil
	case 4:
		if meta.V4 == nil || meta.V4.SorobanMeta == nil || meta.V4.SorobanMeta.ReturnValue == nil {
			return nil, nil
		}
		return []*xdr.ScVal{meta.V4.SorobanMeta.ReturnValue}, nil
	default:
		return nil, nil
	}
}

func transactionEnvelopeOps(envelopeXDR string) ([]xdr.Operation, bool) {
	if strings.TrimSpace(envelopeXDR) == "" {
		return nil, false
	}
	var envelope xdr.TransactionEnvelope
	if err := xdr.SafeUnmarshalBase64(envelopeXDR, &envelope); err != nil {
		return nil, false
	}
	if envelope.Type == xdr.EnvelopeTypeEnvelopeTypeTx && envelope.V1 != nil {
		return envelope.V1.Tx.Operations, true
	}
	if envelope.Type == xdr.EnvelopeTypeEnvelopeTypeTxV0 && envelope.V0 != nil {
		return envelope.V0.Tx.Operations, true
	}
	return nil, false
}

func invokeContractArgsFromEnvelope(envelopeXDR string, applicationOrder int32) ([]xdr.ScVal, string, bool) {
	ops, ok := transactionEnvelopeOps(envelopeXDR)
	if !ok || applicationOrder <= 0 || int(applicationOrder) > len(ops) {
		return nil, "", false
	}
	op := ops[applicationOrder-1]
	if op.Body.Type != xdr.OperationTypeInvokeHostFunction {
		return nil, "", false
	}
	invoke := op.Body.MustInvokeHostFunctionOp()
	if invoke.HostFunction.Type != xdr.HostFunctionTypeHostFunctionTypeInvokeContract {
		return nil, "", false
	}
	args := invoke.HostFunction.MustInvokeContract()
	functionName := strings.TrimSpace(string(args.FunctionName))
	return args.Args, functionName, true
}

func authEntriesFromEnvelope(envelopeXDR string, applicationOrder int32) ([]xdr.SorobanAuthorizationEntry, bool) {
	ops, ok := transactionEnvelopeOps(envelopeXDR)
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
