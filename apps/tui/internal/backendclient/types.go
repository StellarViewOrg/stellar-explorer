package backendclient

import "time"

// APIError matches the error payload returned by the tui-indexer read API.
type APIError struct {
	Error string `json:"error"`
}

// HTTPError captures a non-2xx backend response.
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e HTTPError) Error() string {
	return e.Message
}

// HealthResponse mirrors GET /healthz.
type HealthResponse struct {
	Status             string  `json:"status"`
	Database           string  `json:"database"`
	RPC                string  `json:"rpc"`
	LatestRPCLedger    *uint32 `json:"latest_rpc_ledger,omitempty"`
	LastIngestedLedger uint32  `json:"last_ingested_ledger"`
}

// LedgerSummary mirrors the compact ledger shape in the read API.
type LedgerSummary struct {
	Sequence          uint32    `json:"sequence"`
	Hash              string    `json:"hash"`
	ClosedAt          time.Time `json:"closed_at"`
	TransactionCount  int32     `json:"transaction_count"`
	OperationCount    int32     `json:"operation_count"`
	SuccessfulTxCount int32     `json:"successful_tx_count"`
	FailedTxCount     int32     `json:"failed_tx_count"`
}

// TransactionSummary mirrors the live feed transaction shape in the read API.
type TransactionSummary struct {
	Hash             string    `json:"hash"`
	LedgerSequence   uint32    `json:"ledger_sequence"`
	ApplicationOrder int32     `json:"application_order"`
	Account          string    `json:"account"`
	OperationCount   int32     `json:"operation_count"`
	Status           int16     `json:"status"`
	IsSoroban        bool      `json:"is_soroban"`
	CreatedAt        time.Time `json:"created_at"`
	// Optional indexed metadata used by advanced live-feed filters.
	PrimaryContractID    string `json:"primary_contract_id,omitempty"`
	PrimaryAssetCode     string `json:"primary_asset_code,omitempty"`
	PrimaryAssetIssuer   string `json:"primary_asset_issuer,omitempty"`
	PrimaryOperationType string `json:"primary_operation_type,omitempty"`
}

// LiveFeedSummaryResponse mirrors GET /v1/feed/live/summary.
type LiveFeedSummaryResponse struct {
	LastIngestedLedger uint32               `json:"last_ingested_ledger"`
	LatestLedger       *LedgerSummary       `json:"latest_ledger,omitempty"`
	RecentTransactions []TransactionSummary `json:"recent_transactions"`
}

// LedgerLookupResponse mirrors GET /v1/ledgers/:sequence.
type LedgerLookupResponse struct {
	Ledger       *LedgerSummary       `json:"ledger"`
	Transactions []TransactionSummary `json:"transactions"`
}

// LedgerListResponse mirrors GET /v1/ledgers.
type LedgerListResponse struct {
	Ledgers []LedgerSummary `json:"ledgers"`
}

// SearchResult mirrors one TUI search result returned by tui-indexer.
type SearchResult struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Source      string `json:"source,omitempty"`
}

// SearchResponse mirrors GET /v1/search.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// TransactionDetail mirrors the current transaction lookup payload.
type TransactionDetail struct {
	Hash             string    `json:"hash"`
	LedgerSequence   uint32    `json:"ledger_sequence"`
	ApplicationOrder int32     `json:"application_order"`
	Account          string    `json:"account"`
	AccountMuxed     *string   `json:"account_muxed,omitempty"`
	AccountMuxedID   *int64    `json:"account_muxed_id,omitempty"`
	AccountSequence  int64     `json:"account_sequence"`
	FeeCharged       int64     `json:"fee_charged"`
	MaxFee           int64     `json:"max_fee"`
	OperationCount   int32     `json:"operation_count"`
	MemoType         int16     `json:"memo_type"`
	MemoText         *string   `json:"memo_text,omitempty"`
	MemoHash         *string   `json:"memo_hash,omitempty"`
	Status           int16     `json:"status"`
	IsSoroban        bool      `json:"is_soroban"`
	SorobanResources *string   `json:"soroban_resources,omitempty"`
	EnvelopeXDR      string    `json:"envelope_xdr"`
	ResultXDR        string    `json:"result_xdr"`
	ResultMetaXDR    *string   `json:"result_meta_xdr,omitempty"`
	FeeMetaXDR       *string   `json:"fee_meta_xdr,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// EffectSummary mirrors indexed effects returned for a transaction lookup.
type EffectSummary struct {
	TransactionHash string    `json:"transaction_hash"`
	Type            int16     `json:"type"`
	TypeName        string    `json:"type_name"`
	Account         string    `json:"account"`
	Details         string    `json:"details"`
	CreatedAt       time.Time `json:"created_at"`
}

// OperationSummary mirrors the compact operation list returned for a transaction.
type OperationSummary struct {
	TransactionHash    string    `json:"transaction_hash"`
	ApplicationOrder   int32     `json:"application_order"`
	Type               int16     `json:"type"`
	TypeName           string    `json:"type_name"`
	SourceAccount      *string   `json:"source_account,omitempty"`
	SourceAccountMuxed *string   `json:"source_account_muxed,omitempty"`
	SourceMuxedID      *int64    `json:"source_muxed_id,omitempty"`
	AssetCode          *string   `json:"asset_code,omitempty"`
	AssetIssuer        *string   `json:"asset_issuer,omitempty"`
	Amount             *string   `json:"amount,omitempty"`
	Destination        *string   `json:"destination,omitempty"`
	DestinationMuxed   *string   `json:"destination_muxed,omitempty"`
	DestinationMuxedID *int64    `json:"destination_muxed_id,omitempty"`
	ContractID         *string   `json:"contract_id,omitempty"`
	FunctionName       *string   `json:"function_name,omitempty"`
	Details            string    `json:"details"`
	CreatedAt          time.Time `json:"created_at"`
}

// TransactionLookupResponse mirrors GET /v1/transactions/:hash.
type TransactionLookupResponse struct {
	Transaction *TransactionDetail `json:"transaction"`
	Operations  []OperationSummary `json:"operations"`
	Effects     []EffectSummary    `json:"effects"`
}

// OperationLookupSnapshot is the renderer-facing operation detail payload.
type OperationLookupSnapshot struct {
	ParentTransactionHash string           `json:"parent_transaction_hash"`
	Operation             OperationSummary `json:"operation"`
}

// AccountDetail mirrors GET /v1/accounts/:id.
type AccountDetail struct {
	ID                 string     `json:"id"`
	Sequence           int64      `json:"sequence"`
	SequenceLedger     *int64     `json:"sequence_ledger,omitempty"`
	SequenceTime       *time.Time `json:"sequence_time,omitempty"`
	Balance            string     `json:"balance"`
	BuyingLiabilities  string     `json:"buying_liabilities"`
	SellingLiabilities string     `json:"selling_liabilities"`
	NumSubentries      int32      `json:"num_subentries"`
	HomeDomain         *string    `json:"home_domain,omitempty"`
	Flags              int32      `json:"flags"`
	InflationDest      *string    `json:"inflation_dest,omitempty"`
	Thresholds         *string    `json:"thresholds,omitempty"`
	LastModifiedLedger int64      `json:"last_modified_ledger"`
	Sponsor            *string    `json:"sponsor,omitempty"`
	NumSponsored       int32      `json:"num_sponsored"`
	NumSponsoring      int32      `json:"num_sponsoring"`
	DataEntries        *string    `json:"data_entries,omitempty"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// TrustlineSummary mirrors the trustlines section of GET /v1/accounts/:id.
type TrustlineSummary struct {
	AssetType          int16     `json:"asset_type"`
	AssetCode          string    `json:"asset_code"`
	AssetIssuer        string    `json:"asset_issuer"`
	Balance            string    `json:"balance"`
	LimitAmount        string    `json:"limit_amount"`
	BuyingLiabilities  string    `json:"buying_liabilities"`
	SellingLiabilities string    `json:"selling_liabilities"`
	Flags              int32     `json:"flags"`
	LastModifiedLedger int64     `json:"last_modified_ledger"`
	Sponsor            *string   `json:"sponsor,omitempty"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AccountSignerSummary mirrors the signers section of GET /v1/accounts/:id.
type AccountSignerSummary struct {
	SignerKey          string  `json:"signer_key"`
	Weight             int32   `json:"weight"`
	Type               string  `json:"type"`
	Sponsor            *string `json:"sponsor,omitempty"`
	LastModifiedLedger int64   `json:"last_modified_ledger"`
}

// AccountLookupResponse mirrors GET /v1/accounts/:id.
type AccountLookupResponse struct {
	Account            *AccountDetail         `json:"account"`
	Trustlines         []TrustlineSummary     `json:"trustlines"`
	Signers            []AccountSignerSummary `json:"signers"`
	RecentTransactions []TransactionSummary   `json:"recent_transactions"`
	RecentOperations   []OperationSummary     `json:"recent_operations"`
}

// AccountListResponse mirrors GET /v1/accounts.
type AccountListResponse struct {
	Accounts []AccountDetail `json:"accounts"`
}

// TimelineItem mirrors one normalized entity timeline item from tui-indexer.
type TimelineItem struct {
	Kind        string    `json:"kind"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Command     string    `json:"command"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// TimelineResponse mirrors timeline list endpoints.
type TimelineResponse struct {
	Items []TimelineItem `json:"items"`
}

// AssetHolderSummary mirrors the top holders section of GET /v1/assets/:code::issuer.
type AssetHolderSummary struct {
	AccountID          string    `json:"account_id"`
	Balance            string    `json:"balance"`
	LimitAmount        string    `json:"limit_amount"`
	BuyingLiabilities  string    `json:"buying_liabilities"`
	SellingLiabilities string    `json:"selling_liabilities"`
	LastModifiedLedger int64     `json:"last_modified_ledger"`
	Sponsor            *string   `json:"sponsor,omitempty"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AssetDetail mirrors GET /v1/assets/:code::issuer.
type AssetDetail struct {
	AssetType            int16     `json:"asset_type"`
	AssetCode            string    `json:"asset_code"`
	AssetIssuer          string    `json:"asset_issuer"`
	NumAccounts          int32     `json:"num_accounts"`
	TotalSupply          string    `json:"total_supply"`
	NumClaimableBalances int32     `json:"num_claimable_balances"`
	NumLiquidityPools    int32     `json:"num_liquidity_pools"`
	NumContracts         int32     `json:"num_contracts"`
	Flags                int32     `json:"flags"`
	AuthRequired         bool      `json:"auth_required"`
	AuthRevocable        bool      `json:"auth_revocable"`
	AuthImmutable        bool      `json:"auth_immutable"`
	ClawbackEnabled      bool      `json:"clawback_enabled"`
	HomeDomain           *string   `json:"home_domain,omitempty"`
	TomlName             *string   `json:"toml_name,omitempty"`
	TomlDescription      *string   `json:"toml_description,omitempty"`
	SACContractID        *string   `json:"sac_contract_id,omitempty"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// AssetLookupResponse mirrors GET /v1/assets/:code::issuer.
type AssetLookupResponse struct {
	Asset              *AssetDetail         `json:"asset"`
	TopHolders         []AssetHolderSummary `json:"top_holders"`
	RecentTransactions []TransactionSummary `json:"recent_transactions"`
}

// AssetListResponse mirrors GET /v1/assets.
type AssetListResponse struct {
	Assets []AssetDetail `json:"assets"`
}

// ContractEventSummary mirrors the recent events section of GET /v1/contracts/:id.
type ContractEventSummary struct {
	ContractID       string               `json:"contract_id,omitempty"`
	TransactionHash  string               `json:"transaction_hash"`
	LedgerSequence   uint32               `json:"ledger_sequence"`
	Type             int16                `json:"type"`
	Topic1           *string              `json:"topic_1,omitempty"`
	Topic2           *string              `json:"topic_2,omitempty"`
	Topic3           *string              `json:"topic_3,omitempty"`
	Topic4           *string              `json:"topic_4,omitempty"`
	Topics           []string             `json:"topics,omitempty"`
	TopicsXDR        *string              `json:"topics_xdr,omitempty"`
	ValueXDR         *string              `json:"value_xdr,omitempty"`
	TopicsDecoded    *string              `json:"topics_decoded,omitempty"`
	ValueDecoded     *string              `json:"value_decoded,omitempty"`
	EventName        string               `json:"event_name,omitempty"`
	FieldsDecoded    []ContractEventField `json:"fields_decoded,omitempty"`
	SpecDecodeStatus string               `json:"spec_decode_status,omitempty"`
	DecodeStatus     string               `json:"decode_status"`
	Summary          string               `json:"summary,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
}

// ContractEventField is one spec-decoded event field from the indexer read API.
type ContractEventField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Location string `json:"location,omitempty"`
	Value    string `json:"value"`
}

// ContractDetail mirrors GET /v1/contracts/:id.
type ContractDetail struct {
	ContractID         string    `json:"contract_id"`
	WasmHash           *string   `json:"wasm_hash,omitempty"`
	CreatorAccount     *string   `json:"creator_account,omitempty"`
	CreatedLedger      int64     `json:"created_ledger"`
	CreatedAt          time.Time `json:"created_at"`
	LastModifiedLedger int64     `json:"last_modified_ledger"`
	ContractType       int16     `json:"contract_type"`
	IsSep41Token       bool      `json:"is_sep41_token"`
	IsSep50NFT         bool      `json:"is_sep50_nft"`
	TokenName          *string   `json:"token_name,omitempty"`
	TokenSymbol        *string   `json:"token_symbol,omitempty"`
	TokenDecimals      *int32    `json:"token_decimals,omitempty"`
	ContractSpec       *string   `json:"contract_spec,omitempty"`
	StorageEntryCount  int32     `json:"storage_entry_count"`
	EventCount         int64     `json:"event_count"`
	InvocationCount    int64     `json:"invocation_count"`
	Label              *string   `json:"label,omitempty"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ContractSpec mirrors the structured contract spec payload returned by tui-indexer.
type ContractSpec struct {
	ContractID    string                 `json:"contract_id"`
	WasmHash      *string                `json:"wasm_hash,omitempty"`
	Available     bool                   `json:"available"`
	DecodeStatus  string                 `json:"decode_status"`
	FunctionCount int                    `json:"function_count"`
	SchemaCount   int                    `json:"schema_count"`
	EventCount    int                    `json:"event_count"`
	Functions     []ContractSpecFunction `json:"functions"`
	Schemas       []ContractSpecSchema   `json:"schemas"`
	Events        []ContractSpecEvent    `json:"events"`
	Raw           *string                `json:"raw,omitempty"`
	SpecXDR       *string                `json:"spec_xdr,omitempty"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ContractSpecFunction describes one callable Soroban method.
type ContractSpecFunction struct {
	Name    string              `json:"name"`
	Doc     *string             `json:"doc,omitempty"`
	Inputs  []ContractSpecValue `json:"inputs"`
	Outputs []string            `json:"outputs"`
}

// ContractSpecValue describes one contract method input.
type ContractSpecValue struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ContractSpecSchema describes one named UDT/schema entry.
type ContractSpecSchema struct {
	Kind string  `json:"kind"`
	Name string  `json:"name"`
	Raw  *string `json:"raw,omitempty"`
}

// ContractSpecEvent describes one event entry from the contract spec XDR.
type ContractSpecEvent struct {
	Name         string              `json:"name"`
	Doc          *string             `json:"doc,omitempty"`
	PrefixTopics []string            `json:"prefix_topics,omitempty"`
	DataFormat   string              `json:"data_format,omitempty"`
	Params       []ContractSpecValue `json:"params"`
}

// ContractStorageSummary mirrors one compact contract storage row.
type ContractStorageSummary struct {
	ContractID          string    `json:"contract_id"`
	KeyDecoded          *string   `json:"key_decoded,omitempty"`
	ValueDecoded        *string   `json:"value_decoded,omitempty"`
	KeyXDR              string    `json:"key_xdr"`
	ValueXDR            string    `json:"value_xdr"`
	Durability          int16     `json:"durability"`
	DurabilityLabel     string    `json:"durability_label"`
	TTLLedger           *int64    `json:"ttl_ledger,omitempty"`
	LastModifiedLedger  int64     `json:"last_modified_ledger"`
	UpdatedAt           time.Time `json:"updated_at"`
	DecodeStatus        string    `json:"decode_status"`
	DisplayKey          string    `json:"display_key"`
	DisplayValue        string    `json:"display_value"`
	ExpirationProximity string    `json:"expiration_proximity,omitempty"`
}

// ContractEventLookupSnapshot is the renderer-facing contract event detail payload.
type ContractEventLookupSnapshot struct {
	ParentContractID string               `json:"parent_contract_id"`
	Event            ContractEventSummary `json:"event"`
}

// ContractStorageLookupSnapshot is the renderer-facing contract storage entry detail payload.
type ContractStorageLookupSnapshot struct {
	ParentContractID string                 `json:"parent_contract_id"`
	Entry            ContractStorageSummary `json:"entry"`
}

// ContractLookupResponse mirrors GET /v1/contracts/:id.
type ContractLookupResponse struct {
	Contract           *ContractDetail          `json:"contract"`
	Spec               *ContractSpec            `json:"spec,omitempty"`
	Storage            []ContractStorageSummary `json:"storage"`
	RecentTransactions []TransactionSummary     `json:"recent_transactions"`
	RecentEvents       []ContractEventSummary   `json:"recent_events"`
}

// ContractListResponse mirrors GET /v1/contracts.
type ContractListResponse struct {
	Contracts []ContractDetail `json:"contracts"`
}
