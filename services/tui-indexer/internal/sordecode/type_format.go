package sordecode

import (
	"fmt"
	"strings"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// FormatTypeDef renders a contract spec type definition as a readable type string.
func FormatTypeDef(typeDef xdr.ScSpecTypeDef) string {
	switch typeDef.Type {
	case xdr.ScSpecTypeScSpecTypeVal:
		return "Val"
	case xdr.ScSpecTypeScSpecTypeBool:
		return "bool"
	case xdr.ScSpecTypeScSpecTypeVoid:
		return "void"
	case xdr.ScSpecTypeScSpecTypeError:
		return "Error"
	case xdr.ScSpecTypeScSpecTypeU32:
		return "u32"
	case xdr.ScSpecTypeScSpecTypeI32:
		return "i32"
	case xdr.ScSpecTypeScSpecTypeU64:
		return "u64"
	case xdr.ScSpecTypeScSpecTypeI64:
		return "i64"
	case xdr.ScSpecTypeScSpecTypeTimepoint:
		return "Timepoint"
	case xdr.ScSpecTypeScSpecTypeDuration:
		return "Duration"
	case xdr.ScSpecTypeScSpecTypeU128:
		return "u128"
	case xdr.ScSpecTypeScSpecTypeI128:
		return "i128"
	case xdr.ScSpecTypeScSpecTypeU256:
		return "u256"
	case xdr.ScSpecTypeScSpecTypeI256:
		return "i256"
	case xdr.ScSpecTypeScSpecTypeBytes:
		return "Bytes"
	case xdr.ScSpecTypeScSpecTypeString:
		return "string"
	case xdr.ScSpecTypeScSpecTypeSymbol:
		return "Symbol"
	case xdr.ScSpecTypeScSpecTypeAddress:
		return "Address"
	case xdr.ScSpecTypeScSpecTypeMuxedAddress:
		return "MuxedAddress"
	case xdr.ScSpecTypeScSpecTypeOption:
		option := typeDef.MustOption()
		return fmt.Sprintf("Option<%s>", FormatTypeDef(option.ValueType))
	case xdr.ScSpecTypeScSpecTypeResult:
		result := typeDef.MustResult()
		return fmt.Sprintf("Result<%s,%s>", FormatTypeDef(result.OkType), FormatTypeDef(result.ErrorType))
	case xdr.ScSpecTypeScSpecTypeVec:
		vec := typeDef.MustVec()
		return fmt.Sprintf("Vec<%s>", FormatTypeDef(vec.ElementType))
	case xdr.ScSpecTypeScSpecTypeMap:
		mapType := typeDef.MustMap()
		return fmt.Sprintf("Map<%s,%s>", FormatTypeDef(mapType.KeyType), FormatTypeDef(mapType.ValueType))
	case xdr.ScSpecTypeScSpecTypeTuple:
		tuple := typeDef.MustTuple()
		parts := make([]string, 0, len(tuple.ValueTypes))
		for _, item := range tuple.ValueTypes {
			parts = append(parts, FormatTypeDef(item))
		}
		return fmt.Sprintf("(%s)", strings.Join(parts, ","))
	case xdr.ScSpecTypeScSpecTypeBytesN:
		bytesN := typeDef.MustBytesN()
		return fmt.Sprintf("BytesN<%d>", bytesN.N)
	case xdr.ScSpecTypeScSpecTypeUdt:
		return typeDef.MustUdt().Name
	default:
		return typeDef.Type.String()
	}
}
