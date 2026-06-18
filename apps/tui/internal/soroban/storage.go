package soroban

import (
	"context"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/sordecode"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// StorageFromInstance decodes contract instance storage entries when present.
func StorageFromInstance(ctx context.Context, loader *SpecLoader, contractID string, instance *xdr.ScContractInstance, lastModifiedLedger int64) []backendclient.ContractStorageSummary {
	if instance == nil || instance.Storage == nil {
		return nil
	}

	registry, _ := loader.Registry(ctx, contractID)
	entries := make([]backendclient.ContractStorageSummary, 0, len(*instance.Storage))
	for _, item := range *instance.Storage {
		decoded := sordecode.ScValToString(item.Val)
		if registry != nil {
			if specDecoded := registry.DecodeStorage(item.Key, item.Val); specDecoded.Key != "" || specDecoded.Value != "" {
				decoded = specDecoded.Value
			}
		}
		keyXDR := encodeScValBase64(item.Key)
		valueXDR := encodeScValBase64(item.Val)
		keyDecoded := sordecode.ScValToString(item.Key)
		valueDecoded := decoded
		entries = append(entries, backendclient.ContractStorageSummary{
			ContractID:         contractID,
			KeyDecoded:         &keyDecoded,
			ValueDecoded:       &valueDecoded,
			KeyXDR:             keyXDR,
			ValueXDR:           valueXDR,
			Durability:         2,
			DurabilityLabel:    "instance",
			LastModifiedLedger: lastModifiedLedger,
			UpdatedAt:          time.Time{},
			DecodeStatus:       storageDecodeStatus(keyDecoded, valueDecoded),
			DisplayKey:         keyDecoded,
			DisplayValue:       valueDecoded,
		})
	}
	return entries
}

func storageDecodeStatus(key string, value string) string {
	switch {
	case strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "":
		return "decoded"
	case strings.TrimSpace(key) != "" || strings.TrimSpace(value) != "":
		return "partial"
	default:
		return "raw"
	}
}
