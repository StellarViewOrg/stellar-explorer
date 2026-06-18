package sordecode

import (
	"fmt"
	"math/big"

	"github.com/stellar/go-stellar-sdk/xdr"
)

func int128PartsToBigInt(parts xdr.Int128Parts) *big.Int {
	hi := big.NewInt(int64(parts.Hi))
	hi.Lsh(hi, 64)
	lo := new(big.Int).SetUint64(uint64(parts.Lo))
	return hi.Add(hi, lo)
}

func uint128PartsToBigInt(parts xdr.UInt128Parts) *big.Int {
	hi := new(big.Int).SetUint64(uint64(parts.Hi))
	lo := new(big.Int).SetUint64(uint64(parts.Lo))
	hi.Lsh(hi, 64)
	hi.Add(hi, lo)
	return hi
}

func int256PartsToBigInt(parts xdr.Int256Parts) *big.Int {
	unsigned := uint256PartsToBigInt(xdr.UInt256Parts{
		HiHi: xdr.Uint64(uint64(parts.HiHi)),
		HiLo: xdr.Uint64(uint64(parts.HiLo)),
		LoHi: xdr.Uint64(uint64(parts.LoHi)),
		LoLo: xdr.Uint64(uint64(parts.LoLo)),
	})
	if int64(parts.HiHi) < 0 {
		unsigned.Sub(unsigned, new(big.Int).Lsh(big.NewInt(1), 256))
	}
	return unsigned
}

func uint256PartsToBigInt(parts xdr.UInt256Parts) *big.Int {
	words := []uint64{
		uint64(parts.HiHi),
		uint64(parts.HiLo),
		uint64(parts.LoHi),
		uint64(parts.LoLo),
	}
	value := big.NewInt(0)
	for _, word := range words {
		value.Lsh(value, 64)
		value.Add(value, new(big.Int).SetUint64(word))
	}
	return value
}

func formatInt128(parts xdr.Int128Parts) string {
	return int128PartsToBigInt(parts).String()
}

func formatUInt128(parts xdr.UInt128Parts) string {
	return uint128PartsToBigInt(parts).String()
}

func formatInt256(parts xdr.Int256Parts) string {
	return int256PartsToBigInt(parts).String()
}

func formatUInt256(parts xdr.UInt256Parts) string {
	return uint256PartsToBigInt(parts).String()
}

func formatScError(err xdr.ScError) string {
	switch err.Type {
	case xdr.ScErrorTypeSceContract:
		if err.ContractCode != nil {
			return fmt.Sprintf("contract_error(%d)", *err.ContractCode)
		}
	case xdr.ScErrorTypeSceWasmVm, xdr.ScErrorTypeSceContext, xdr.ScErrorTypeSceStorage,
		xdr.ScErrorTypeSceObject, xdr.ScErrorTypeSceCrypto, xdr.ScErrorTypeSceEvents,
		xdr.ScErrorTypeSceBudget, xdr.ScErrorTypeSceValue, xdr.ScErrorTypeSceAuth:
		if err.Code != nil {
			return fmt.Sprintf("%s(%d)", err.Type.String(), *err.Code)
		}
	}
	return err.Type.String()
}
