package transform

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/stellar/go-stellar-sdk/network"
	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

// LedgerFromRPC converts an RPC LedgerEntry into a store.Ledger.
func LedgerFromRPC(entry source.LedgerEntry) (*store.Ledger, error) {
	var headerEntry xdr.LedgerHeaderHistoryEntry
	if err := xdr.SafeUnmarshalBase64(entry.HeaderXDR, &headerEntry); err != nil {
		return nil, fmt.Errorf("unmarshal ledger header: %w", err)
	}
	header := headerEntry.Header

	closedAt, err := parseLedgerCloseTime(entry.LedgerCloseTime)
	if err != nil {
		return nil, fmt.Errorf("parse ledger close time: %w", err)
	}

	headerXDR := entry.HeaderXDR
	return &store.Ledger{
		Sequence:        uint32(header.LedgerSeq),
		Hash:            entry.Hash,
		PrevHash:        hex.EncodeToString(header.PreviousLedgerHash[:]),
		ClosedAt:        closedAt,
		TotalCoins:      int64(header.TotalCoins),
		FeePool:         int64(header.FeePool),
		BaseFee:         int32(header.BaseFee),
		BaseReserve:     int32(header.BaseReserve),
		MaxTxSetSize:    int32(header.MaxTxSetSize),
		ProtocolVersion: int32(header.LedgerVersion),
		HeaderXDR:       &headerXDR,
	}, nil
}

// TransactionFromRPC converts an RPC TransactionEntry into a store.Transaction.
func TransactionFromRPC(entry source.TransactionEntry, networkPassphrase string) (*store.Transaction, error) {
	var envelope xdr.TransactionEnvelope
	if err := xdr.SafeUnmarshalBase64(entry.EnvelopeXDR, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal envelope: %w", err)
	}

	var result xdr.TransactionResult
	if err := xdr.SafeUnmarshalBase64(entry.ResultXDR, &result); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	txHash, err := computeTransactionHash(envelope, networkPassphrase)
	if err != nil {
		return nil, fmt.Errorf("compute tx hash: %w", err)
	}

	sourceAccount := envelope.SourceAccount()
	// Always use the base G-address (ToAccountId strips the muxed memo ID).
	// The full M-address is stored separately in account_muxed when present.
	accountAddr := sourceAccount.ToAccountId().Address()

	memo := envelope.Memo()
	memoType := int16(memo.Type)
	var memoText *string
	var memoHash *string
	switch memo.Type {
	case xdr.MemoTypeMemoText:
		text := memo.MustText()
		memoText = &text
	case xdr.MemoTypeMemoHash:
		h := memo.MustHash()
		hashStr := hex.EncodeToString(h[:])
		memoHash = &hashStr
	case xdr.MemoTypeMemoReturn:
		h := memo.MustRetHash()
		hashStr := hex.EncodeToString(h[:])
		memoHash = &hashStr
	case xdr.MemoTypeMemoId:
		id := memo.MustId()
		idStr := fmt.Sprintf("%d", id)
		memoText = &idStr
	}

	// Determine muxed account fields if applicable
	_, accountMuxed, accountMuxedID := parseMuxedAccount(sourceAccount)

	// Status: 1 = success, 0 = failed
	var status int16
	resultCode := result.Result.Code
	if resultCode == xdr.TransactionResultCodeTxSuccess || resultCode == xdr.TransactionResultCodeTxFeeBumpInnerSuccess {
		status = 1
	}

	// Detect Soroban transactions
	isSoroban := hasSorobanOp(envelope.Operations())

	resultMetaXDR := entry.ResultMetaXDR
	createdAt := time.Unix(entry.CreatedAt, 0).UTC()

	return &store.Transaction{
		Hash:             txHash,
		LedgerSequence:   entry.Ledger,
		ApplicationOrder: entry.ApplicationOrder,
		Account:          accountAddr,
		AccountMuxed:     accountMuxed,
		AccountMuxedID:   accountMuxedID,
		AccountSequence:  envelope.SeqNum(),
		FeeCharged:       int64(result.FeeCharged),
		MaxFee:           int64(envelope.Fee()),
		OperationCount:   int32(envelope.OperationsCount()),
		MemoType:         memoType,
		MemoText:         memoText,
		MemoHash:         memoHash,
		Status:           status,
		IsSoroban:        isSoroban,
		EnvelopeXDR:      entry.EnvelopeXDR,
		ResultXDR:        entry.ResultXDR,
		ResultMetaXDR:    &resultMetaXDR,
		CreatedAt:        createdAt,
	}, nil
}

// OperationsFromRPC extracts operations from a transaction entry.
func OperationsFromRPC(entry source.TransactionEntry, networkPassphrase string) ([]store.Operation, error) {
	var envelope xdr.TransactionEnvelope
	if err := xdr.SafeUnmarshalBase64(entry.EnvelopeXDR, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal envelope: %w", err)
	}

	txHash, err := computeTransactionHash(envelope, networkPassphrase)
	if err != nil {
		return nil, fmt.Errorf("compute tx hash: %w", err)
	}

	ops := envelope.Operations()
	createdAt := time.Unix(entry.CreatedAt, 0).UTC()
	result := make([]store.Operation, 0, len(ops))

	for i, op := range ops {
		opType := op.Body.Type
		typeName := operationTypeName(opType)

		var sourceAccount *string
		var sourceAccountMuxed *string
		var sourceMuxedID *int64
		if op.SourceAccount != nil {
			base, muxed, muxedID := parseMuxedAccount(*op.SourceAccount)
			sourceAccount = &base
			sourceAccountMuxed = muxed
			sourceMuxedID = muxedID
		}

		details := extractOperationDetails(op)
		detailsJSON, _ := json.Marshal(details)

		storeOp := store.Operation{
			TransactionHash:    txHash,
			ApplicationOrder:   int32(i + 1),
			Type:               int16(opType),
			TypeName:           typeName,
			SourceAccount:      sourceAccount,
			SourceAccountMuxed: sourceAccountMuxed,
			SourceMuxedID:      sourceMuxedID,
			Details:            string(detailsJSON),
			CreatedAt:          createdAt,
		}

		// Extract denormalized fields from specific operation types
		enrichOperation(&storeOp, op, details)

		result = append(result, storeOp)
	}

	return result, nil
}

func computeTransactionHash(envelope xdr.TransactionEnvelope, networkPassphrase string) (string, error) {
	hash, err := network.HashTransactionInEnvelope(envelope, networkPassphrase)
	if err != nil {
		return "", fmt.Errorf("hash transaction: %w", err)
	}
	return hex.EncodeToString(hash[:]), nil
}

func parseLedgerCloseTime(s string) (time.Time, error) {
	// Try unix timestamp (integer as string)
	var ts int64
	if _, err := fmt.Sscanf(s, "%d", &ts); err == nil {
		return time.Unix(ts, 0).UTC(), nil
	}

	// Try RFC3339
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse %q as timestamp", s)
	}
	return t.UTC(), nil
}

func hasSorobanOp(ops []xdr.Operation) bool {
	for _, op := range ops {
		switch op.Body.Type {
		case xdr.OperationTypeInvokeHostFunction,
			xdr.OperationTypeExtendFootprintTtl,
			xdr.OperationTypeRestoreFootprint:
			return true
		}
	}
	return false
}

func operationTypeName(t xdr.OperationType) string {
	s := t.String()
	// Convert from "OperationTypePayment" to "payment"
	s = strings.TrimPrefix(s, "OperationType")
	if s == "" {
		return fmt.Sprintf("unknown_%d", int32(t))
	}
	// Convert CamelCase to snake_case
	return camelToSnake(s)
}

func camelToSnake(s string) string {
	var result []byte
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, byte(c+'a'-'A'))
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

func extractOperationDetails(op xdr.Operation) map[string]interface{} {
	details := map[string]interface{}{
		"type": operationTypeName(op.Body.Type),
	}

	switch op.Body.Type {
	case xdr.OperationTypePayment:
		payment := op.Body.MustPaymentOp()
		details["destination"] = payment.Destination.Address()
		details["amount"] = fmt.Sprintf("%d", payment.Amount)
		details["asset"] = assetString(payment.Asset)
	case xdr.OperationTypeCreateAccount:
		create := op.Body.MustCreateAccountOp()
		details["destination"] = create.Destination.Address()
		details["starting_balance"] = fmt.Sprintf("%d", create.StartingBalance)
	case xdr.OperationTypePathPaymentStrictReceive:
		pp := op.Body.MustPathPaymentStrictReceiveOp()
		details["destination"] = pp.Destination.Address()
		details["dest_amount"] = fmt.Sprintf("%d", pp.DestAmount)
		details["dest_asset"] = assetString(pp.DestAsset)
		details["send_asset"] = assetString(pp.SendAsset)
		details["send_max"] = fmt.Sprintf("%d", pp.SendMax)
	case xdr.OperationTypePathPaymentStrictSend:
		pp := op.Body.MustPathPaymentStrictSendOp()
		details["destination"] = pp.Destination.Address()
		details["send_amount"] = fmt.Sprintf("%d", pp.SendAmount)
		details["send_asset"] = assetString(pp.SendAsset)
		details["dest_asset"] = assetString(pp.DestAsset)
		details["dest_min"] = fmt.Sprintf("%d", pp.DestMin)
	case xdr.OperationTypeInvokeHostFunction:
		invoke := op.Body.MustInvokeHostFunctionOp()
		details["function_type"] = invoke.HostFunction.Type.String()
		for key, value := range extractInvokeHostFunctionDetails(invoke) {
			details[key] = value
		}
	case xdr.OperationTypeChangeTrust:
		ct := op.Body.MustChangeTrustOp()
		details["limit"] = fmt.Sprintf("%d", ct.Limit)
	}

	return details
}

func enrichOperation(storeOp *store.Operation, op xdr.Operation, details map[string]interface{}) {
	switch op.Body.Type {
	case xdr.OperationTypePayment:
		payment := op.Body.MustPaymentOp()
		base, muxed, muxedID := parseMuxedAccount(payment.Destination)
		storeOp.Destination = &base
		storeOp.DestinationMuxed = muxed
		storeOp.DestinationMuxedID = muxedID
		amount := fmt.Sprintf("%d", payment.Amount)
		storeOp.Amount = &amount
		code, issuer := assetParts(payment.Asset)
		storeOp.AssetCode = code
		storeOp.AssetIssuer = issuer
	case xdr.OperationTypeCreateAccount:
		create := op.Body.MustCreateAccountOp()
		dest := create.Destination.Address()
		storeOp.Destination = &dest
		amount := fmt.Sprintf("%d", create.StartingBalance)
		storeOp.Amount = &amount
	case xdr.OperationTypePathPaymentStrictReceive:
		pp := op.Body.MustPathPaymentStrictReceiveOp()
		base, muxed, muxedID := parseMuxedAccount(pp.Destination)
		storeOp.Destination = &base
		storeOp.DestinationMuxed = muxed
		storeOp.DestinationMuxedID = muxedID
		amount := fmt.Sprintf("%d", pp.DestAmount)
		storeOp.Amount = &amount
		code, issuer := assetParts(pp.DestAsset)
		storeOp.AssetCode = code
		storeOp.AssetIssuer = issuer
	case xdr.OperationTypePathPaymentStrictSend:
		pp := op.Body.MustPathPaymentStrictSendOp()
		base, muxed, muxedID := parseMuxedAccount(pp.Destination)
		storeOp.Destination = &base
		storeOp.DestinationMuxed = muxed
		storeOp.DestinationMuxedID = muxedID
		amount := fmt.Sprintf("%d", pp.SendAmount)
		storeOp.Amount = &amount
		code, issuer := assetParts(pp.SendAsset)
		storeOp.AssetCode = code
		storeOp.AssetIssuer = issuer
	case xdr.OperationTypeInvokeHostFunction:
		invoke := op.Body.MustInvokeHostFunctionOp()
		enrichInvokeHostFunction(storeOp, invoke, details)
		if storeOp.FunctionName == nil {
			fnType := invoke.HostFunction.Type.String()
			storeOp.FunctionName = &fnType
		}
	}
}

func extractInvokeHostFunctionDetails(invoke xdr.InvokeHostFunctionOp) map[string]interface{} {
	details := map[string]interface{}{}
	enrichInvokeHostFunction(&store.Operation{}, invoke, details)
	return details
}

// parseMuxedAccount splits a MuxedAccount into:
//   - base: the plain 56-char G-address (always present)
//   - muxed: the full M-address (only when muxed, nil otherwise)
//   - muxedID: the 64-bit integer muxed ID (only when muxed, nil otherwise)
func parseMuxedAccount(m xdr.MuxedAccount) (base string, muxed *string, muxedID *int64) {
	base = m.ToAccountId().Address()
	if m.Type == xdr.CryptoKeyTypeKeyTypeMuxedEd25519 {
		addr := m.Address()
		muxed = &addr
		if rawID, ok := m.GetMed25519(); ok {
			id64 := int64(rawID.Id)
			muxedID = &id64
		}
	}
	return
}

func assetString(asset xdr.Asset) string {
	switch asset.Type {
	case xdr.AssetTypeAssetTypeNative:
		return "native"
	case xdr.AssetTypeAssetTypeCreditAlphanum4:
		a4 := asset.MustAlphaNum4()
		return fmt.Sprintf("%s:%s", strings.TrimRight(string(a4.AssetCode[:]), "\x00"), a4.Issuer.Address())
	case xdr.AssetTypeAssetTypeCreditAlphanum12:
		a12 := asset.MustAlphaNum12()
		return fmt.Sprintf("%s:%s", strings.TrimRight(string(a12.AssetCode[:]), "\x00"), a12.Issuer.Address())
	default:
		return "unknown"
	}
}

func assetParts(asset xdr.Asset) (*string, *string) {
	switch asset.Type {
	case xdr.AssetTypeAssetTypeNative:
		code := "XLM"
		return &code, nil
	case xdr.AssetTypeAssetTypeCreditAlphanum4:
		a4 := asset.MustAlphaNum4()
		code := strings.TrimRight(string(a4.AssetCode[:]), "\x00")
		issuer := a4.Issuer.Address()
		return &code, &issuer
	case xdr.AssetTypeAssetTypeCreditAlphanum12:
		a12 := asset.MustAlphaNum12()
		code := strings.TrimRight(string(a12.AssetCode[:]), "\x00")
		issuer := a12.Issuer.Address()
		return &code, &issuer
	default:
		return nil, nil
	}
}
