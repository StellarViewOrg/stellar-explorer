package sordecode

import (
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
)

func TestScValToString_I128FullWidth(t *testing.T) {
	parts := xdr.Int128Parts{Hi: 1, Lo: 0}
	value := xdr.ScVal{Type: xdr.ScValTypeScvI128, I128: &parts}
	if got := ScValToString(value); got != "18446744073709551616" {
		t.Fatalf("expected full i128 width, got %q", got)
	}
}

func TestScValToString_U128FullWidth(t *testing.T) {
	parts := xdr.UInt128Parts{Hi: 1, Lo: 1}
	value := xdr.ScVal{Type: xdr.ScValTypeScvU128, U128: &parts}
	if got := ScValToString(value); got != "18446744073709551617" {
		t.Fatalf("expected full u128 width, got %q", got)
	}
}

func TestScValToString_NegativeI128(t *testing.T) {
	parts := xdr.Int128Parts{Hi: -1, Lo: 0}
	value := xdr.ScVal{Type: xdr.ScValTypeScvI128, I128: &parts}
	if got := ScValToString(value); got != "-18446744073709551616" {
		t.Fatalf("expected negative i128, got %q", got)
	}
}

func TestFormatTypeDef_OptionVec(t *testing.T) {
	inner := xdr.ScSpecTypeDef{Type: xdr.ScSpecTypeScSpecTypeString}
	optionType, err := xdr.NewScSpecTypeDef(xdr.ScSpecTypeScSpecTypeOption, xdr.ScSpecTypeOption{ValueType: inner})
	if err != nil {
		t.Fatalf("build option type: %v", err)
	}
	if got := FormatTypeDef(optionType); got != "Option<string>" {
		t.Fatalf("unexpected type format: %q", got)
	}
}
