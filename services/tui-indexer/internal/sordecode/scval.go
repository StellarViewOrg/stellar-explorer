package sordecode

import (
	"fmt"
	"strings"

	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// ScValToString renders an ScVal as a human-readable string without contract spec context.
func ScValToString(v xdr.ScVal) string {
	switch v.Type {
	case xdr.ScValTypeScvSymbol:
		if v.Sym != nil {
			return string(*v.Sym)
		}
	case xdr.ScValTypeScvString:
		if v.Str != nil {
			return string(*v.Str)
		}
	case xdr.ScValTypeScvBool:
		if v.B != nil {
			if *v.B {
				return "true"
			}
			return "false"
		}
	case xdr.ScValTypeScvI128:
		if v.I128 != nil {
			return formatInt128(*v.I128)
		}
	case xdr.ScValTypeScvU128:
		if v.U128 != nil {
			return formatUInt128(*v.U128)
		}
	case xdr.ScValTypeScvI256:
		if v.I256 != nil {
			return formatInt256(*v.I256)
		}
	case xdr.ScValTypeScvU256:
		if v.U256 != nil {
			return formatUInt256(*v.U256)
		}
	case xdr.ScValTypeScvI64:
		if v.I64 != nil {
			return fmt.Sprintf("%d", *v.I64)
		}
	case xdr.ScValTypeScvU64:
		if v.U64 != nil {
			return fmt.Sprintf("%d", *v.U64)
		}
	case xdr.ScValTypeScvI32:
		if v.I32 != nil {
			return fmt.Sprintf("%d", *v.I32)
		}
	case xdr.ScValTypeScvU32:
		if v.U32 != nil {
			return fmt.Sprintf("%d", *v.U32)
		}
	case xdr.ScValTypeScvTimepoint:
		if v.Timepoint != nil {
			return fmt.Sprintf("timepoint(%d)", *v.Timepoint)
		}
	case xdr.ScValTypeScvDuration:
		if v.Duration != nil {
			return fmt.Sprintf("duration(%d)", *v.Duration)
		}
	case xdr.ScValTypeScvAddress:
		if v.Address != nil {
			return formatScAddress(*v.Address)
		}
	case xdr.ScValTypeScvBytes:
		if v.Bytes != nil {
			return fmt.Sprintf("0x%x", []byte(*v.Bytes))
		}
	case xdr.ScValTypeScvVoid:
		return "void"
	case xdr.ScValTypeScvError:
		if v.Error != nil {
			return formatScError(*v.Error)
		}
		return "error"
	case xdr.ScValTypeScvMap:
		if v.Map != nil && *v.Map != nil {
			m := **v.Map
			parts := make([]string, 0, len(m))
			for _, entry := range m {
				parts = append(parts, fmt.Sprintf("%s:%s", ScValToString(entry.Key), ScValToString(entry.Val)))
			}
			return "{" + strings.Join(parts, ",") + "}"
		}
	case xdr.ScValTypeScvVec:
		if v.Vec != nil && *v.Vec != nil {
			vec := **v.Vec
			parts := make([]string, 0, len(vec))
			for _, item := range vec {
				parts = append(parts, ScValToString(item))
			}
			return "[" + strings.Join(parts, ",") + "]"
		}
	case xdr.ScValTypeScvContractInstance:
		if v.Instance != nil {
			if v.Instance.Executable.Type == xdr.ContractExecutableTypeContractExecutableWasm {
				hash := v.Instance.Executable.MustWasmHash()
				return fmt.Sprintf("contract_instance(0x%x)", hash[:])
			}
			return "contract_instance"
		}
	case xdr.ScValTypeScvLedgerKeyContractInstance:
		return "ledger_key_contract_instance"
	case xdr.ScValTypeScvLedgerKeyNonce:
		if v.NonceKey != nil {
			return fmt.Sprintf("nonce(%d)", v.NonceKey.Nonce)
		}
	}
	return fmt.Sprintf("scval_type_%d", v.Type)
}

// ScAddressToString renders a Soroban ScAddress as a strkey or readable label.
func ScAddressToString(addr xdr.ScAddress) string {
	return formatScAddress(addr)
}

func formatScAddress(addr xdr.ScAddress) string {
	switch addr.Type {
	case xdr.ScAddressTypeScAddressTypeAccount:
		if addr.AccountId != nil {
			return addr.AccountId.Address()
		}
	case xdr.ScAddressTypeScAddressTypeContract:
		if addr.ContractId != nil {
			encoded, err := strkey.Encode(strkey.VersionByteContract, (*addr.ContractId)[:])
			if err == nil {
				return encoded
			}
		}
	case xdr.ScAddressTypeScAddressTypeMuxedAccount:
		if addr.MuxedAccount != nil {
			accountID := xdr.AccountId{
				Type:    xdr.PublicKeyTypePublicKeyTypeEd25519,
				Ed25519: &addr.MuxedAccount.Ed25519,
			}
			return fmt.Sprintf("muxed(%s:%d)", accountID.Address(), addr.MuxedAccount.Id)
		}
	}
	return "unknown_address"
}
