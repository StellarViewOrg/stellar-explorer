package transform

import (
	"fmt"

	"github.com/stellar/go-stellar-sdk/xdr"
)

func sorobanReturnValuesFromTransactionMeta(resultMetaXDR string) ([]*xdr.ScVal, error) {
	if resultMetaXDR == "" {
		return nil, nil
	}
	var meta xdr.TransactionMeta
	if err := xdr.SafeUnmarshalBase64(resultMetaXDR, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal transaction meta: %w", err)
	}

	switch meta.V {
	case 3:
		if meta.V3 == nil || meta.V3.SorobanMeta == nil {
			return nil, nil
		}
		value := meta.V3.SorobanMeta.ReturnValue
		return []*xdr.ScVal{&value}, nil
	case 4:
		if meta.V4 == nil || meta.V4.SorobanMeta == nil || meta.V4.SorobanMeta.ReturnValue == nil {
			return nil, nil
		}
		return []*xdr.ScVal{meta.V4.SorobanMeta.ReturnValue}, nil
	default:
		return nil, nil
	}
}
