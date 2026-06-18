package soroban

import (
	"encoding/base64"
	"fmt"

	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

func contractInstanceLedgerKey(contractID string) (string, error) {
	contractIDBytes, err := strkey.Decode(strkey.VersionByteContract, contractID)
	if err != nil {
		return "", fmt.Errorf("decode contract id: %w", err)
	}
	var id xdr.ContractId
	copy(id[:], contractIDBytes)

	scAddr := xdr.ScAddress{
		Type:       xdr.ScAddressTypeScAddressTypeContract,
		ContractId: &id,
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

func extractWasmHashFromInstance(dataXDR string) (xdr.Hash, *xdr.ScContractInstance, error) {
	var data xdr.LedgerEntryData
	if err := xdr.SafeUnmarshalBase64(dataXDR, &data); err != nil {
		return xdr.Hash{}, nil, fmt.Errorf("unmarshal ledger entry: %w", err)
	}
	contractEntry, ok := data.GetContractData()
	if !ok {
		return xdr.Hash{}, nil, fmt.Errorf("ledger entry is not contract data")
	}
	instance, ok := contractEntry.Val.GetInstance()
	if !ok {
		return xdr.Hash{}, nil, fmt.Errorf("contract data value is not an instance")
	}
	if instance.Executable.Type != xdr.ContractExecutableTypeContractExecutableWasm {
		return xdr.Hash{}, &instance, fmt.Errorf("contract is not a wasm contract")
	}
	hash, ok := instance.Executable.GetWasmHash()
	if !ok {
		return xdr.Hash{}, &instance, fmt.Errorf("wasm hash unavailable")
	}
	return hash, &instance, nil
}

func extractWasmBytecode(dataXDR string) ([]byte, error) {
	var data xdr.LedgerEntryData
	if err := xdr.SafeUnmarshalBase64(dataXDR, &data); err != nil {
		return nil, fmt.Errorf("unmarshal ledger entry: %w", err)
	}
	codeEntry, ok := data.GetContractCode()
	if !ok {
		return nil, fmt.Errorf("ledger entry is not contract code")
	}
	return codeEntry.Code, nil
}
