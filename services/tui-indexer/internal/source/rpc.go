package source

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// RPCClient communicates with a Stellar RPC node via JSON-RPC 2.0.
type RPCClient struct {
	endpoint   string
	httpClient *http.Client
}

func NewRPCClient(endpoint string) *RPCClient {
	return &RPCClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- JSON-RPC request / response envelope ---

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *jsonRPCError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

// --- getLatestLedger ---

type LatestLedgerResult struct {
	ID          string `json:"id"`
	Sequence    uint32 `json:"sequence"`
	CloseTime   string `json:"closeTime"`
	HeaderXDR   string `json:"headerXdr"`
	MetadataXDR string `json:"metadataXdr"`
}

func (c *RPCClient) GetLatestLedger(ctx context.Context) (*LatestLedgerResult, error) {
	raw, err := c.call(ctx, "getLatestLedger", nil)
	if err != nil {
		return nil, err
	}
	var result LatestLedgerResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal getLatestLedger: %w", err)
	}
	return &result, nil
}

// --- getLedgers ---

type GetLedgersParams struct {
	StartLedger uint32      `json:"startLedger,omitempty"`
	Pagination  *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type LedgerEntry struct {
	Hash            string `json:"hash"`
	Sequence        uint32 `json:"sequence"`
	LedgerCloseTime string `json:"ledgerCloseTime"`
	HeaderXDR       string `json:"headerXdr"`
	MetadataXDR     string `json:"metadataXdr"`
}

type GetLedgersResult struct {
	Ledgers               []LedgerEntry `json:"ledgers"`
	LatestLedger          uint32        `json:"latestLedger"`
	LatestLedgerCloseTime int64         `json:"latestLedgerCloseTime"`
	OldestLedger          uint32        `json:"oldestLedger"`
	OldestLedgerCloseTime int64         `json:"oldestLedgerCloseTime"`
	Cursor                string        `json:"cursor"`
}

func (c *RPCClient) GetLedgers(ctx context.Context, params GetLedgersParams) (*GetLedgersResult, error) {
	raw, err := c.call(ctx, "getLedgers", params)
	if err != nil {
		return nil, err
	}
	var result GetLedgersResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal getLedgers: %w", err)
	}
	return &result, nil
}

// --- getTransactions ---

type GetTransactionsParams struct {
	StartLedger uint32      `json:"startLedger,omitempty"`
	Pagination  *Pagination `json:"pagination,omitempty"`
}

type TransactionEntry struct {
	Status              string   `json:"status"`
	ApplicationOrder    int32    `json:"applicationOrder"`
	FeeBump             bool     `json:"feeBump"`
	EnvelopeXDR         string   `json:"envelopeXdr"`
	ResultXDR           string   `json:"resultXdr"`
	ResultMetaXDR       string   `json:"resultMetaXdr"`
	DiagnosticEventsXDR []string `json:"diagnosticEventsXdr,omitempty"`
	Ledger              uint32   `json:"ledger"`
	CreatedAt           int64    `json:"createdAt"`
}

type GetTransactionsResult struct {
	Transactions               []TransactionEntry `json:"transactions"`
	LatestLedger               uint32             `json:"latestLedger"`
	LatestLedgerCloseTimestamp int64              `json:"latestLedgerCloseTimestamp"`
	OldestLedger               uint32             `json:"oldestLedger"`
	OldestLedgerCloseTimestamp int64              `json:"oldestLedgerCloseTimestamp"`
	Cursor                     string             `json:"cursor"`
}

func (c *RPCClient) GetTransactions(ctx context.Context, params GetTransactionsParams) (*GetTransactionsResult, error) {
	raw, err := c.call(ctx, "getTransactions", params)
	if err != nil {
		return nil, err
	}
	var result GetTransactionsResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal getTransactions: %w", err)
	}
	return &result, nil
}

// --- getLedgerEntries ---

type GetLedgerEntriesParams struct {
	Keys []string `json:"keys"`
}

type LedgerEntryResult struct {
	Key                string  `json:"key"`
	XDR                string  `json:"xdr"`
	LastModifiedLedger uint32  `json:"lastModifiedLedgerSeq"`
	LiveUntilLedger    *uint32 `json:"liveUntilLedgerSeq,omitempty"`
}

type GetLedgerEntriesResult struct {
	Entries      []LedgerEntryResult `json:"entries"`
	LatestLedger uint32              `json:"latestLedger"`
}

func (c *RPCClient) GetLedgerEntries(ctx context.Context, keys []string) (*GetLedgerEntriesResult, error) {
	raw, err := c.call(ctx, "getLedgerEntries", GetLedgerEntriesParams{Keys: keys})
	if err != nil {
		return nil, err
	}
	var result GetLedgerEntriesResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal getLedgerEntries: %w", err)
	}
	return &result, nil
}

// --- internal transport ---

var requestID atomic.Int64

func (c *RPCClient) call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	id := int(requestID.Add(1))
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}
