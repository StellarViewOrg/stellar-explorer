package transform

import (
	"fmt"
	"log"
	"time"

	assetProto "github.com/stellar/go-stellar-sdk/asset"
	"github.com/stellar/go-stellar-sdk/processors/token_transfer"
	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

const (
	tokenEventTypeTransfer int16 = 0
	tokenEventTypeMint     int16 = 1
	tokenEventTypeBurn     int16 = 2
	tokenEventTypeClawback int16 = 3
	tokenEventTypeFee      int16 = 4

	assetTypeNative       int16 = 0
	assetTypeCredit       int16 = 1
	assetTypeSorobanToken int16 = 2
)

// TokenEventsFromLedgerMeta extracts CAP-67 unified token transfer events from a
// base64-encoded LedgerCloseMeta XDR string (the metadataXdr field from getLedgers RPC).
func TokenEventsFromLedgerMeta(metaXDR string, networkPassphrase string) ([]store.TokenEvent, error) {
	if metaXDR == "" {
		return nil, nil
	}

	var lcm xdr.LedgerCloseMeta
	if err := xdr.SafeUnmarshalBase64(metaXDR, &lcm); err != nil {
		return nil, fmt.Errorf("unmarshal LedgerCloseMeta: %w", err)
	}

	ttp := token_transfer.NewEventsProcessor(networkPassphrase)
	events, err := ttp.EventsFromLedger(lcm)
	if err != nil {
		return nil, fmt.Errorf("extract token events from ledger: %w", err)
	}

	result := make([]store.TokenEvent, 0, len(events))
	for _, event := range events {
		te, err := tokenEventFromProto(event)
		if err != nil {
			log.Printf("token_events: skip event: %v", err)
			continue
		}
		result = append(result, *te)
	}
	return result, nil
}

func tokenEventFromProto(event *token_transfer.TokenTransferEvent) (*store.TokenEvent, error) {
	meta := event.GetMeta()
	if meta == nil {
		return nil, fmt.Errorf("event has no meta")
	}

	eventTypeName := event.GetEventType()
	var eventType int16
	var fromAddr, toAddr *string

	switch e := event.GetEvent().(type) {
	case *token_transfer.TokenTransferEvent_Transfer:
		eventType = tokenEventTypeTransfer
		from := e.Transfer.GetFrom()
		to := e.Transfer.GetTo()
		if from != "" {
			fromAddr = &from
		}
		if to != "" {
			toAddr = &to
		}
	case *token_transfer.TokenTransferEvent_Mint:
		eventType = tokenEventTypeMint
		to := e.Mint.GetTo()
		if to != "" {
			toAddr = &to
		}
	case *token_transfer.TokenTransferEvent_Burn:
		eventType = tokenEventTypeBurn
		from := e.Burn.GetFrom()
		if from != "" {
			fromAddr = &from
		}
	case *token_transfer.TokenTransferEvent_Clawback:
		eventType = tokenEventTypeClawback
		from := e.Clawback.GetFrom()
		if from != "" {
			fromAddr = &from
		}
	case *token_transfer.TokenTransferEvent_Fee:
		eventType = tokenEventTypeFee
		from := e.Fee.GetFrom()
		if from != "" {
			fromAddr = &from
		}
	default:
		return nil, fmt.Errorf("unknown event type: %T", event.GetEvent())
	}

	// Muxed destination info
	var toMuxedID *int64
	if muxedInfo := meta.GetToMuxedInfo(); muxedInfo != nil {
		if idInfo, ok := muxedInfo.Content.(*token_transfer.MuxedInfo_Id); ok {
			id := int64(idInfo.Id)
			toMuxedID = &id
		}
	}

	// Asset classification
	assetType, assetCode, assetIssuer, assetContractID := mapProtoAsset(event.GetAsset(), meta.GetContractAddress())

	// Operation index (nil for fee events, which have no operation)
	var opIndex *int32
	if meta.OperationIndex != nil {
		idx := int32(meta.GetOperationIndex())
		opIndex = &idx
	}

	// Ledger close time
	var createdAt time.Time
	if ts := meta.GetClosedAt(); ts != nil {
		createdAt = ts.AsTime()
	}

	return &store.TokenEvent{
		EventType:       eventType,
		EventTypeName:   eventTypeName,
		FromAddress:     fromAddr,
		ToAddress:       toAddr,
		ToMuxedID:       toMuxedID,
		AssetType:       assetType,
		AssetCode:       assetCode,
		AssetIssuer:     assetIssuer,
		AssetContractID: assetContractID,
		Amount:          event.GetAmount(),
		TransactionHash: meta.GetTxHash(),
		LedgerSequence:  meta.GetLedgerSequence(),
		OperationIndex:  opIndex,
		CreatedAt:       createdAt,
	}, nil
}

// mapProtoAsset determines the DB asset type, code, issuer, and contract ID from the
// proto Asset (which is nil for pure Soroban tokens) and the ContractAddress in EventMeta.
func mapProtoAsset(a *assetProto.Asset, contractAddress string) (assetType int16, code *string, issuer *string, contractID *string) {
	var contractIDPtr *string
	if contractAddress != "" {
		contractIDPtr = &contractAddress
	}

	if a == nil {
		// Pure Soroban token: no underlying classic asset
		return assetTypeSorobanToken, nil, nil, contractIDPtr
	}

	switch a.GetAssetType().(type) {
	case *assetProto.Asset_Native:
		xlm := "XLM"
		return assetTypeNative, &xlm, nil, contractIDPtr
	case *assetProto.Asset_IssuedAsset:
		ia := a.GetIssuedAsset()
		c := ia.GetAssetCode()
		i := ia.GetIssuer()
		return assetTypeCredit, &c, &i, contractIDPtr
	default:
		// Fallback: treat as native if type is unrecognized
		xlm := "XLM"
		return assetTypeNative, &xlm, nil, contractIDPtr
	}
}
