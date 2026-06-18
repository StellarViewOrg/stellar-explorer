package rpcclient

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetEventsFilter narrows getEvents results.
type GetEventsFilter struct {
	Type        string   `json:"type,omitempty"`
	ContractIds []string `json:"contractIds,omitempty"`
}

// GetEventsParams configures a getEvents RPC request.
type GetEventsParams struct {
	StartLedger uint32            `json:"startLedger,omitempty"`
	EndLedger   uint32            `json:"endLedger,omitempty"`
	Filters     []GetEventsFilter `json:"filters,omitempty"`
	Pagination  *Pagination       `json:"pagination,omitempty"`
}

// EventEntry mirrors one Soroban event from getEvents.
type EventEntry struct {
	Type             string   `json:"type"`
	Ledger           uint32   `json:"ledger"`
	LedgerClosedAt   string   `json:"ledgerClosedAt"`
	ContractID       string   `json:"contractId"`
	ID               string   `json:"id"`
	TxHash           string   `json:"txHash"`
	Topic            []string `json:"topic"`
	Value            string   `json:"value"`
	TransactionIndex int32    `json:"transactionIndex"`
	OperationIndex   int32    `json:"operationIndex"`
}

// GetEventsResult mirrors the getEvents RPC response payload.
type GetEventsResult struct {
	Events       []EventEntry `json:"events"`
	LatestLedger uint32       `json:"latestLedger"`
	OldestLedger uint32       `json:"oldestLedger,omitempty"`
	Cursor       string       `json:"cursor"`
}

// GetEvents fetches contract events from Stellar RPC.
func (c *Client) GetEvents(ctx context.Context, params GetEventsParams) (*GetEventsResult, error) {
	raw, err := c.call(ctx, "getEvents", params)
	if err != nil {
		return nil, err
	}

	var result GetEventsResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal getEvents: %w", err)
	}
	return &result, nil
}
