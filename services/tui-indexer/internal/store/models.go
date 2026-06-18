package store

import "time"

// Ledger represents a row in the ledgers hypertable.
type Ledger struct {
	Sequence            uint32    `db:"sequence"`
	Hash                string    `db:"hash"`
	PrevHash            string    `db:"prev_hash"`
	ClosedAt            time.Time `db:"closed_at"`
	TotalCoins          int64     `db:"total_coins"`
	FeePool             int64     `db:"fee_pool"`
	BaseFee             int32     `db:"base_fee"`
	BaseReserve         int32     `db:"base_reserve"`
	MaxTxSetSize        int32     `db:"max_tx_set_size"`
	ProtocolVersion     int32     `db:"protocol_version"`
	TransactionCount    int32     `db:"transaction_count"`
	OperationCount      int32     `db:"operation_count"`
	SuccessfulTxCount   int32     `db:"successful_tx_count"`
	FailedTxCount       int32     `db:"failed_tx_count"`
	TxSetOperationCount *int32    `db:"tx_set_operation_count"`
	HeaderXDR           *string   `db:"header_xdr"`
}

// Transaction represents a row in the transactions hypertable.
type Transaction struct {
	Hash             string    `db:"hash"`
	LedgerSequence   uint32    `db:"ledger_sequence"`
	ApplicationOrder int32     `db:"application_order"`
	Account          string    `db:"account"`
	AccountMuxed     *string   `db:"account_muxed"`
	AccountMuxedID   *int64    `db:"account_muxed_id"`
	AccountSequence  int64     `db:"account_sequence"`
	FeeCharged       int64     `db:"fee_charged"`
	MaxFee           int64     `db:"max_fee"`
	OperationCount   int32     `db:"operation_count"`
	MemoType         int16     `db:"memo_type"`
	MemoText         *string   `db:"memo_text"`
	MemoHash         *string   `db:"memo_hash"`
	Status           int16     `db:"status"`
	IsSoroban        bool      `db:"is_soroban"`
	SorobanResources *string   `db:"soroban_resources"` // JSON
	EnvelopeXDR      string    `db:"envelope_xdr"`
	ResultXDR        string    `db:"result_xdr"`
	ResultMetaXDR    *string   `db:"result_meta_xdr"`
	FeeMetaXDR       *string   `db:"fee_meta_xdr"`
	CreatedAt        time.Time `db:"created_at"`
}

// Operation represents a row in the operations hypertable.
type Operation struct {
	TransactionID      int64     `db:"transaction_id"`
	TransactionHash    string    `db:"transaction_hash"`
	ApplicationOrder   int32     `db:"application_order"`
	Type               int16     `db:"type"`
	TypeName           string    `db:"type_name"`
	SourceAccount      *string   `db:"source_account"`
	SourceAccountMuxed *string   `db:"source_account_muxed"`
	SourceMuxedID      *int64    `db:"source_muxed_id"`
	AssetCode          *string   `db:"asset_code"`
	AssetIssuer        *string   `db:"asset_issuer"`
	Amount             *string   `db:"amount"`
	Destination        *string   `db:"destination"`
	DestinationMuxed   *string   `db:"destination_muxed"`
	DestinationMuxedID *int64    `db:"destination_muxed_id"`
	ContractID         *string   `db:"contract_id"`
	FunctionName       *string   `db:"function_name"`
	Details            string    `db:"details"` // JSON
	CreatedAt          time.Time `db:"created_at"`
}

// Effect represents a row in the effects hypertable.
type Effect struct {
	OperationID     int64     `db:"operation_id"`
	TransactionHash string    `db:"transaction_hash"`
	Type            int16     `db:"type"`
	TypeName        string    `db:"type_name"`
	Account         string    `db:"account"`
	Details         string    `db:"details"` // JSON
	CreatedAt       time.Time `db:"created_at"`
}

// TokenEvent represents a row in the token_events hypertable (CAP-67 unified events).
type TokenEvent struct {
	EventType       int16     `db:"event_type"`      // 0=transfer, 1=mint, 2=burn, 3=clawback, 4=fee
	EventTypeName   string    `db:"event_type_name"` // "transfer", "mint", etc.
	FromAddress     *string   `db:"from_address"`
	FromMuxed       *string   `db:"from_muxed"`
	ToAddress       *string   `db:"to_address"`
	ToMuxed         *string   `db:"to_muxed"`
	ToMuxedID       *int64    `db:"to_muxed_id"`
	AssetType       int16     `db:"asset_type"` // 0=native, 1=credit, 2=soroban_token
	AssetCode       *string   `db:"asset_code"`
	AssetIssuer     *string   `db:"asset_issuer"`
	AssetContractID *string   `db:"asset_contract_id"`
	Amount          string    `db:"amount"`           // i128 as decimal string
	AmountFormatted *string   `db:"amount_formatted"` // formatted with decimals (optional)
	TransactionHash string    `db:"transaction_hash"`
	LedgerSequence  uint32    `db:"ledger_sequence"`
	OperationIndex  *int32    `db:"operation_index"`
	CreatedAt       time.Time `db:"created_at"`
}

// Contract represents a row in the contracts table.
type Contract struct {
	ContractID         string    `db:"contract_id"`
	WasmHash           *string   `db:"wasm_hash"`
	CreatorAccount     *string   `db:"creator_account"`
	CreatedLedger      uint32    `db:"created_ledger"`
	CreatedAt          time.Time `db:"created_at"`
	LastModifiedLedger uint32    `db:"last_modified_ledger"`
	ContractType       int16     `db:"contract_type"` // 0=wasm, 1=stellar_asset, 2=custom
	IsSep41Token       bool      `db:"is_sep41_token"`
	IsSep50NFT         bool      `db:"is_sep50_nft"`
	TokenName          *string   `db:"token_name"`
	TokenSymbol        *string   `db:"token_symbol"`
	TokenDecimals      *int32    `db:"token_decimals"`
	ContractSpec       *string   `db:"contract_spec"` // JSON
}

// ContractCode represents a row in the contract_code table.
type ContractCode struct {
	WasmHash      string    `db:"wasm_hash"`
	WasmBytecode  []byte    `db:"wasm_bytecode"`
	WasmSize      int32     `db:"wasm_size"`
	SpecXDR       *string   `db:"spec_xdr"`    // base64-encoded raw spec XDR
	SpecParsed    *string   `db:"spec_parsed"` // JSON parsed spec
	CreatedLedger uint32    `db:"created_ledger"`
	CreatedAt     time.Time `db:"created_at"`
}

// ContractEvent represents a row in the contract_events hypertable.
type ContractEvent struct {
	ContractID      string    `db:"contract_id"`
	TransactionHash string    `db:"transaction_hash"`
	LedgerSequence  uint32    `db:"ledger_sequence"`
	Type            int16     `db:"type"` // 0=contract, 1=system, 2=diagnostic
	Topic1          *string   `db:"topic_1"`
	Topic2          *string   `db:"topic_2"`
	Topic3          *string   `db:"topic_3"`
	Topic4          *string   `db:"topic_4"`
	TopicsXDR       string    `db:"topics_xdr"`     // base64 encoded
	ValueXDR        string    `db:"value_xdr"`      // base64 encoded
	TopicsDecoded   *string   `db:"topics_decoded"` // JSON
	ValueDecoded    *string   `db:"value_decoded"`  // JSON
	CreatedAt       time.Time `db:"created_at"`
}

// ReadLedgerSummary is a compact ledger shape for read APIs.
type ReadLedgerSummary struct {
	Sequence          uint32    `json:"sequence"`
	Hash              string    `json:"hash"`
	ClosedAt          time.Time `json:"closed_at"`
	TransactionCount  int32     `json:"transaction_count"`
	OperationCount    int32     `json:"operation_count"`
	SuccessfulTxCount int32     `json:"successful_tx_count"`
	FailedTxCount     int32     `json:"failed_tx_count"`
}

// ReadTransactionSummary is a compact transaction shape for live feed reads.
type ReadTransactionSummary struct {
	Hash                 string    `json:"hash"`
	LedgerSequence       uint32    `json:"ledger_sequence"`
	ApplicationOrder     int32     `json:"application_order"`
	Account              string    `json:"account"`
	OperationCount       int32     `json:"operation_count"`
	Status               int16     `json:"status"`
	IsSoroban            bool      `json:"is_soroban"`
	CreatedAt            time.Time `json:"created_at"`
	PrimaryContractID    string    `json:"primary_contract_id,omitempty"`
	PrimaryAssetCode     string    `json:"primary_asset_code,omitempty"`
	PrimaryAssetIssuer   string    `json:"primary_asset_issuer,omitempty"`
	PrimaryOperationType string    `json:"primary_operation_type,omitempty"`
}

// ReadTransactionDetail is the first transaction lookup shape exposed by the read API.
type ReadTransactionDetail struct {
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

// ReadEffectSummary is a compact effect shape for transaction lookups.
type ReadEffectSummary struct {
	TransactionHash string    `json:"transaction_hash"`
	Type            int16     `json:"type"`
	TypeName        string    `json:"type_name"`
	Account         string    `json:"account"`
	Details         string    `json:"details"`
	CreatedAt       time.Time `json:"created_at"`
}

// ReadOperationSummary is a compact operation shape for transaction lookups.
type ReadOperationSummary struct {
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

// ReadAccountDetail is the first account lookup shape exposed by the read API.
type ReadAccountDetail struct {
	ID                 string    `json:"id"`
	Sequence           int64     `json:"sequence"`
	Balance            string    `json:"balance"`
	BuyingLiabilities  string    `json:"buying_liabilities"`
	SellingLiabilities string    `json:"selling_liabilities"`
	NumSubentries      int32     `json:"num_subentries"`
	HomeDomain         *string   `json:"home_domain,omitempty"`
	Flags              int32     `json:"flags"`
	InflationDest      *string   `json:"inflation_dest,omitempty"`
	Thresholds         *string   `json:"thresholds,omitempty"`
	LastModifiedLedger int64     `json:"last_modified_ledger"`
	Sponsor            *string   `json:"sponsor,omitempty"`
	NumSponsored       int32     `json:"num_sponsored"`
	NumSponsoring      int32     `json:"num_sponsoring"`
	DataEntries        *string   `json:"data_entries,omitempty"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ReadTrustlineSummary is a compact trustline shape for account lookups.
type ReadTrustlineSummary struct {
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

// ReadAccountSignerSummary is a compact signer shape for account lookups.
type ReadAccountSignerSummary struct {
	SignerKey          string  `json:"signer_key"`
	Weight             int32   `json:"weight"`
	Type               string  `json:"type"`
	Sponsor            *string `json:"sponsor,omitempty"`
	LastModifiedLedger int64   `json:"last_modified_ledger"`
}

// ReadContractDetail is the first contract lookup shape exposed by the read API.
type ReadContractDetail struct {
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

// ReadContractSpec is a structured contract spec shape for Soroban-oriented views.
type ReadContractSpec struct {
	ContractID    string                     `json:"contract_id"`
	WasmHash      *string                    `json:"wasm_hash,omitempty"`
	Available     bool                       `json:"available"`
	DecodeStatus  string                     `json:"decode_status"`
	FunctionCount int                        `json:"function_count"`
	SchemaCount   int                        `json:"schema_count"`
	EventCount    int                        `json:"event_count"`
	Functions     []ReadContractSpecFunction `json:"functions"`
	Schemas       []ReadContractSpecSchema   `json:"schemas"`
	Events        []ReadContractSpecEvent    `json:"events"`
	Raw           *string                    `json:"raw,omitempty"`
	SpecXDR       *string                    `json:"spec_xdr,omitempty"`
	UpdatedAt     time.Time                  `json:"updated_at"`
}

// ReadContractSpecEvent describes one event entry from the contract spec XDR.
type ReadContractSpecEvent struct {
	Name         string                  `json:"name"`
	Doc          *string                 `json:"doc,omitempty"`
	PrefixTopics []string                `json:"prefix_topics,omitempty"`
	DataFormat   string                  `json:"data_format,omitempty"`
	Params       []ReadContractSpecValue `json:"params"`
}

// ReadContractSpecFunction describes one callable function from the contract spec.
type ReadContractSpecFunction struct {
	Name    string                  `json:"name"`
	Doc     *string                 `json:"doc,omitempty"`
	Inputs  []ReadContractSpecValue `json:"inputs"`
	Outputs []string                `json:"outputs"`
}

// ReadContractSpecValue describes one function input.
type ReadContractSpecValue struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ReadContractSpecSchema describes one named UDT/schema entry from the contract spec.
type ReadContractSpecSchema struct {
	Kind string  `json:"kind"`
	Name string  `json:"name"`
	Raw  *string `json:"raw,omitempty"`
}

// ReadContractStorageSummary is a compact storage-entry shape for contract explorer views.
type ReadContractStorageSummary struct {
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

// ReadContractEventField is one spec-decoded event field for read API views.
type ReadContractEventField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Location string `json:"location,omitempty"`
	Value    string `json:"value"`
}

// ReadContractEventSummary is a compact contract-event shape for contract explorer views.
type ReadContractEventSummary struct {
	ContractID       string                   `json:"contract_id,omitempty"`
	TransactionHash  string                   `json:"transaction_hash"`
	LedgerSequence   uint32                   `json:"ledger_sequence"`
	Type             int16                    `json:"type"`
	Topic1           *string                  `json:"topic_1,omitempty"`
	Topic2           *string                  `json:"topic_2,omitempty"`
	Topic3           *string                  `json:"topic_3,omitempty"`
	Topic4           *string                  `json:"topic_4,omitempty"`
	Topics           []string                 `json:"topics,omitempty"`
	TopicsXDR        *string                  `json:"topics_xdr,omitempty"`
	ValueXDR         *string                  `json:"value_xdr,omitempty"`
	TopicsDecoded    *string                  `json:"topics_decoded,omitempty"`
	ValueDecoded     *string                  `json:"value_decoded,omitempty"`
	EventName        string                   `json:"event_name,omitempty"`
	FieldsDecoded    []ReadContractEventField `json:"fields_decoded,omitempty"`
	SpecDecodeStatus string                   `json:"spec_decode_status,omitempty"`
	DecodeStatus     string                   `json:"decode_status"`
	Summary          string                   `json:"summary,omitempty"`
	CreatedAt        time.Time                `json:"created_at"`
}

// ReadAssetHolderSummary is a compact asset-holder shape for asset explorer views.
type ReadAssetHolderSummary struct {
	AccountID          string    `json:"account_id"`
	Balance            string    `json:"balance"`
	LimitAmount        string    `json:"limit_amount"`
	BuyingLiabilities  string    `json:"buying_liabilities"`
	SellingLiabilities string    `json:"selling_liabilities"`
	LastModifiedLedger int64     `json:"last_modified_ledger"`
	Sponsor            *string   `json:"sponsor,omitempty"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ReadAssetDetail is the first asset lookup shape exposed by the read API.
type ReadAssetDetail struct {
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

// ReadSearchResult is a compact result shape for TUI discovery.
type ReadSearchResult struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command"`
}

// ReadTimelineItem is a normalized, executable activity row for entity timelines.
type ReadTimelineItem struct {
	Kind        string    `json:"kind"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Command     string    `json:"command"`
	OccurredAt  time.Time `json:"occurred_at"`
}
