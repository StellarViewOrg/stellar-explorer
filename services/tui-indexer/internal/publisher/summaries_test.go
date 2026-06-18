package publisher

import (
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

func TestSummariesFromStoreUsesPrimaryOperationMetadata(t *testing.T) {
	code := "USDC"
	issuer := "GABC"
	summaries := SummariesFromStore(
		[]store.Transaction{{Hash: "tx-1", LedgerSequence: 10, Account: "GDEF", OperationCount: 2, Status: 1}},
		[]store.Operation{
			{TransactionHash: "tx-1", ApplicationOrder: 2, TypeName: "payment", AssetCode: &code, AssetIssuer: &issuer},
			{TransactionHash: "tx-1", ApplicationOrder: 1, TypeName: "invoke_host_function", ContractID: strPtr("CCONTRACT")},
		},
	)

	if len(summaries) != 1 {
		t.Fatalf("expected one summary, got %d", len(summaries))
	}
	if summaries[0].PrimaryContractID != "CCONTRACT" {
		t.Fatalf("expected primary contract, got %q", summaries[0].PrimaryContractID)
	}
	if summaries[0].PrimaryOperationType != "invoke_host_function" {
		t.Fatalf("expected invoke_host_function, got %q", summaries[0].PrimaryOperationType)
	}
}
