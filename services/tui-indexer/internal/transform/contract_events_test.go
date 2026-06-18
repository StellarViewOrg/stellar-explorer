package transform

import (
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
)

func TestScValToString_Symbol(t *testing.T) {
	sym := xdr.ScSymbol("transfer")
	v := xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &sym}
	if got := scValToString(v); got != "transfer" {
		t.Errorf("expected 'transfer', got %q", got)
	}
}

func TestScValToString_Bool(t *testing.T) {
	b := true
	v := xdr.ScVal{Type: xdr.ScValTypeScvBool, B: &b}
	if got := scValToString(v); got != "true" {
		t.Errorf("expected 'true', got %q", got)
	}
}

func TestScValToString_Void(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvVoid}
	if got := scValToString(v); got != "void" {
		t.Errorf("expected 'void', got %q", got)
	}
}

func TestScValToString_Address_Account(t *testing.T) {
	const addr = "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"
	accountID := xdr.MustAddress(addr)
	scAddr := xdr.ScAddress{
		Type:      xdr.ScAddressTypeScAddressTypeAccount,
		AccountId: &accountID,
	}
	v := xdr.ScVal{Type: xdr.ScValTypeScvAddress, Address: &scAddr}
	if got := scValToString(v); got != addr {
		t.Errorf("expected %q, got %q", addr, got)
	}
}

func TestContractEventsFromTransaction_EmptyMeta(t *testing.T) {
	entry := source.TransactionEntry{
		EnvelopeXDR:   "",
		ResultMetaXDR: "",
		Ledger:        100,
		CreatedAt:     1700000000,
	}
	events, err := ContractEventsFromTransaction(entry, "Test SDF Network ; September 2015")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events for empty meta, got %d", len(events))
	}
}

func TestScValToString_I64(t *testing.T) {
	var n xdr.Int64 = 42
	v := xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: &n}
	if got := scValToString(v); got != "42" {
		t.Errorf("expected '42', got %q", got)
	}
}
