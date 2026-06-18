package transform

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

const contractTypWasm int16 = 0

// sep41FunctionNames are the required functions for a SEP-41 token contract.
var sep41FunctionNames = []string{"balance", "transfer", "name", "symbol", "decimals"}

// sep50FunctionNames are the required functions for a SEP-50 NFT contract.
var sep50FunctionNames = []string{"owner_of", "token_uri", "balance"}

// DetectedContract holds data extracted from a createContract operation.
type DetectedContract struct {
	ContractID     string
	CreatorAccount string
	CreatedLedger  uint32
	CreatedAt      time.Time
}

// collectTxApplyChanges extracts ledger entry changes from a TransactionMeta.
func collectTxApplyChanges(ledgerSeq uint32, meta xdr.TransactionMeta, out *[]xdr.LedgerEntryChange) {
	log.Printf("detect_contracts: ledger %d tx_meta_version=%d", ledgerSeq, meta.V)
	switch meta.V {
	case 3:
		if meta.V3 != nil {
			*out = append(*out, meta.V3.TxChangesAfter...)
			for _, op := range meta.V3.Operations {
				*out = append(*out, op.Changes...)
			}
		}
	case 4:
		if meta.V4 != nil {
			*out = append(*out, meta.V4.TxChangesAfter...)
			for _, op := range meta.V4.Operations {
				*out = append(*out, op.Changes...)
			}
		}
	}
}

// DetectNewContracts scans TransactionMeta for newly created contract instances.
// Returns a list of contracts detected in this ledger.
func DetectNewContracts(metaXDR string, ledgerSeq uint32, closedAt time.Time) ([]DetectedContract, error) {
	if metaXDR == "" {
		return nil, nil
	}

	var lcm xdr.LedgerCloseMeta
	if err := xdr.SafeUnmarshalBase64(metaXDR, &lcm); err != nil {
		return nil, fmt.Errorf("unmarshal LedgerCloseMeta: %w", err)
	}

	var allChanges []xdr.LedgerEntryChange
	switch lcm.V {
	case 0:
		if lcm.V0 != nil {
			log.Printf("detect_contracts: ledger %d lcm=V0 txs=%d", ledgerSeq, len(lcm.V0.TxProcessing))
			for _, txMeta := range lcm.V0.TxProcessing {
				allChanges = append(allChanges, txMeta.TxApplyProcessing.V3.TxChangesAfter...)
				for _, op := range txMeta.TxApplyProcessing.V3.Operations {
					allChanges = append(allChanges, op.Changes...)
				}
			}
		}
	case 1:
		if lcm.V1 != nil {
			log.Printf("detect_contracts: ledger %d lcm=V1 txs=%d", ledgerSeq, len(lcm.V1.TxProcessing))
			for _, txMeta := range lcm.V1.TxProcessing {
				collectTxApplyChanges(ledgerSeq, txMeta.TxApplyProcessing, &allChanges)
			}
		}
	case 2:
		if lcm.V2 != nil {
			log.Printf("detect_contracts: ledger %d lcm=V2 txs=%d", ledgerSeq, len(lcm.V2.TxProcessing))
			for _, txMeta := range lcm.V2.TxProcessing {
				collectTxApplyChanges(ledgerSeq, txMeta.TxApplyProcessing, &allChanges)
			}
		}
	}
	log.Printf("detect_contracts: ledger %d total_changes=%d", ledgerSeq, len(allChanges))

	var detected []DetectedContract
	for _, change := range allChanges {
		if change.Type != xdr.LedgerEntryChangeTypeLedgerEntryCreated {
			continue
		}
		entry := change.Created
		if entry == nil || entry.Data.Type != xdr.LedgerEntryTypeContractData {
			continue
		}
		cd := entry.Data.ContractData
		if cd == nil {
			continue
		}
		// We're looking for instance entries (key = ScvLedgerKeyContractInstance)
		if cd.Key.Type != xdr.ScValTypeScvLedgerKeyContractInstance {
			continue
		}
		if cd.Contract.Type != xdr.ScAddressTypeScAddressTypeContract {
			continue
		}
		if cd.Contract.ContractId == nil {
			continue
		}

		contractID, err := strkey.Encode(strkey.VersionByteContract, (*cd.Contract.ContractId)[:])
		if err != nil {
			continue
		}

		detected = append(detected, DetectedContract{
			ContractID:    contractID,
			CreatedLedger: ledgerSeq,
			CreatedAt:     closedAt,
		})
	}
	return detected, nil
}

// ProcessContractSpec fetches WASM, parses the contract spec, classifies the contract,
// and upserts it into the store. This is designed to run asynchronously.
func ProcessContractSpec(ctx context.Context, rpc *source.RPCClient, db *store.PostgresStore, contract DetectedContract) {
	log.Printf("contract_spec: starting processing for %s (ledger %d)", contract.ContractID, contract.CreatedLedger)
	// Step 1: Fetch contract instance to get wasm_hash
	instanceKey, err := contractInstanceLedgerKey(contract.ContractID)
	if err != nil {
		log.Printf("contract_spec: build instance key for %s: %v", contract.ContractID, err)
		return
	}

	instanceResult, err := rpc.GetLedgerEntries(ctx, []string{instanceKey})
	if err != nil || len(instanceResult.Entries) == 0 {
		log.Printf("contract_spec: fetch instance for %s: %v", contract.ContractID, err)
		return
	}

	wasmHash, err := extractWasmHashFromInstance(instanceResult.Entries[0].XDR)
	if err != nil {
		log.Printf("contract_spec: extract wasm hash for %s: %v", contract.ContractID, err)
		return
	}

	// Step 2: Fetch WASM bytecode using the wasm_hash
	codeKey, err := contractCodeLedgerKey(wasmHash)
	if err != nil {
		log.Printf("contract_spec: build code key for %s: %v", contract.ContractID, err)
		return
	}

	codeResult, err := rpc.GetLedgerEntries(ctx, []string{codeKey})
	if err != nil || len(codeResult.Entries) == 0 {
		log.Printf("contract_spec: fetch code for %s: %v", contract.ContractID, err)
		return
	}

	wasmBytes, err := extractWasmBytecode(codeResult.Entries[0].XDR)
	if err != nil {
		log.Printf("contract_spec: extract wasm for %s: %v", contract.ContractID, err)
		return
	}

	// Step 3: Parse WASM custom section for contract spec XDR
	specXDRBytes, err := sordecode.ExtractSpecFromWASM(wasmBytes)
	if err != nil {
		// Not all contracts have a spec section (e.g., SACs)
		log.Printf("contract_spec: no spec section for %s: %v", contract.ContractID, err)
	}

	// Step 4: Decode spec entries and classify contract
	var specEntries []xdr.ScSpecEntry
	var specParsedJSON *string
	var specXDRBase64 *string

	if len(specXDRBytes) > 0 {
		specEntries, err = sordecode.DecodeEntries(specXDRBytes)
		if err != nil {
			log.Printf("contract_spec: decode spec for %s: %v", contract.ContractID, err)
		} else {
			// Encode raw XDR for storage
			b64 := base64.StdEncoding.EncodeToString(specXDRBytes)
			specXDRBase64 = &b64

			// Marshal parsed spec to JSON
			parsed := sordecode.EntriesToJSON(specEntries)
			if j, err := json.Marshal(parsed); err == nil {
				s := string(j)
				specParsedJSON = &s
			}
		}
	}

	isSep41 := classifyAsSep41(specEntries)
	isSep50 := classifyAsSep50(specEntries)

	wasmHashHex := hex.EncodeToString(wasmHash[:])

	// Upsert contract code
	contractCode := &store.ContractCode{
		WasmHash:      wasmHashHex,
		WasmBytecode:  wasmBytes,
		WasmSize:      int32(len(wasmBytes)),
		SpecXDR:       specXDRBase64,
		SpecParsed:    specParsedJSON,
		CreatedLedger: contract.CreatedLedger,
		CreatedAt:     contract.CreatedAt,
	}
	if err := db.UpsertContractCode(ctx, contractCode); err != nil {
		log.Printf("contract_spec: upsert code for %s: %v", contract.ContractID, err)
	}

	// Upsert contract record
	storeContract := &store.Contract{
		ContractID:         contract.ContractID,
		WasmHash:           &wasmHashHex,
		CreatorAccount:     nilIfEmpty(contract.CreatorAccount),
		CreatedLedger:      contract.CreatedLedger,
		CreatedAt:          contract.CreatedAt,
		LastModifiedLedger: contract.CreatedLedger,
		ContractType:       contractTypWasm,
		IsSep41Token:       isSep41,
		IsSep50NFT:         isSep50,
		ContractSpec:       specParsedJSON,
	}
	if err := db.UpsertContract(ctx, storeContract); err != nil {
		log.Printf("contract_spec: upsert contract %s: %v", contract.ContractID, err)
	}

	log.Printf("contract_spec: processed %s (sep41=%v sep50=%v wasm=%d bytes)",
		contract.ContractID, isSep41, isSep50, len(wasmBytes))
}

// --- Ledger key builders ---

func contractInstanceLedgerKey(contractID string) (string, error) {
	contractIDBytes, err := strkey.Decode(strkey.VersionByteContract, contractID)
	if err != nil {
		return "", fmt.Errorf("decode contract ID: %w", err)
	}
	var cID xdr.ContractId
	copy(cID[:], contractIDBytes)

	scAddr := xdr.ScAddress{
		Type:       xdr.ScAddressTypeScAddressTypeContract,
		ContractId: &cID,
	}
	key := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract:   scAddr,
			Key:        xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyContractInstance},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}
	eb := xdr.NewEncodingBuffer()
	b, err := eb.MarshalBinary(key)
	if err != nil {
		return "", fmt.Errorf("marshal ledger key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func contractCodeLedgerKey(wasmHash xdr.Hash) (string, error) {
	key := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractCode,
		ContractCode: &xdr.LedgerKeyContractCode{
			Hash: wasmHash,
		},
	}
	eb := xdr.NewEncodingBuffer()
	b, err := eb.MarshalBinary(key)
	if err != nil {
		return "", fmt.Errorf("marshal code ledger key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// --- XDR extraction helpers ---

func extractWasmHashFromInstance(entryXDR string) (xdr.Hash, error) {
	// getLedgerEntries returns LedgerEntryData (not the full LedgerEntry wrapper).
	var data xdr.LedgerEntryData
	if err := xdr.SafeUnmarshalBase64(entryXDR, &data); err != nil {
		return xdr.Hash{}, fmt.Errorf("unmarshal ledger entry: %w", err)
	}
	if data.Type != xdr.LedgerEntryTypeContractData {
		return xdr.Hash{}, fmt.Errorf("expected ContractData entry, got %v", data.Type)
	}
	cd := data.ContractData
	if cd == nil {
		return xdr.Hash{}, fmt.Errorf("nil contract data")
	}
	// The contract instance value contains an ScContractInstance with executable info
	instance, ok := cd.Val.GetInstance()
	if !ok {
		return xdr.Hash{}, fmt.Errorf("contract data value is not an instance")
	}
	if instance.Executable.Type != xdr.ContractExecutableTypeContractExecutableWasm {
		return xdr.Hash{}, fmt.Errorf("contract is not a WASM contract")
	}
	return instance.Executable.MustWasmHash(), nil
}

func extractWasmBytecode(entryXDR string) ([]byte, error) {
	// getLedgerEntries returns LedgerEntryData (not the full LedgerEntry wrapper).
	var data xdr.LedgerEntryData
	if err := xdr.SafeUnmarshalBase64(entryXDR, &data); err != nil {
		return nil, fmt.Errorf("unmarshal ledger entry: %w", err)
	}
	if data.Type != xdr.LedgerEntryTypeContractCode {
		return nil, fmt.Errorf("expected ContractCode entry, got %v", data.Type)
	}
	code := data.ContractCode
	if code == nil {
		return nil, fmt.Errorf("nil contract code")
	}
	return code.Code, nil
}

// --- Contract classifiers ---

// classifyAsSep41 checks whether the spec contains all required SEP-41 token functions.
func classifyAsSep41(entries []xdr.ScSpecEntry) bool {
	return hasAllFunctions(entries, sep41FunctionNames)
}

// classifyAsSep50 checks whether the spec contains all required SEP-50 NFT functions.
func classifyAsSep50(entries []xdr.ScSpecEntry) bool {
	return hasAllFunctions(entries, sep50FunctionNames)
}

func hasAllFunctions(entries []xdr.ScSpecEntry, required []string) bool {
	if len(entries) == 0 {
		return false
	}
	found := make(map[string]bool)
	for _, entry := range entries {
		if entry.Kind == xdr.ScSpecEntryKindScSpecEntryFunctionV0 {
			fn := entry.MustFunctionV0()
			found[strings.ToLower(string(fn.Name))] = true
		}
	}
	for _, name := range required {
		if !found[name] {
			return false
		}
	}
	return true
}

// --- Utility ---

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
