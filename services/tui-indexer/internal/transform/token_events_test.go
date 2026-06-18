package transform

import (
	"strings"
	"testing"
	"time"

	assetProto "github.com/stellar/go-stellar-sdk/asset"
	"github.com/stellar/go-stellar-sdk/network"
	"github.com/stellar/go-stellar-sdk/processors/token_transfer"
	"github.com/stellar/go-stellar-sdk/xdr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// makeTestMeta creates a minimal EventMeta for tests.
func makeTestMeta(ledger uint32, txHash string, opIdx *uint32) *token_transfer.EventMeta {
	return &token_transfer.EventMeta{
		LedgerSequence: ledger,
		ClosedAt:       timestamppb.New(time.Unix(1700000000, 0)),
		TxHash:         txHash,
		OperationIndex: opIdx,
	}
}

func opIndexPtr(i uint32) *uint32 { return &i }

func nativeXDRAsset() xdr.Asset {
	return xdr.Asset{Type: xdr.AssetTypeAssetTypeNative}
}

func creditXDRAsset(code, issuer string) xdr.Asset {
	if len(code) <= 4 {
		var c [4]byte
		copy(c[:], strings.TrimRight(code, "\x00"))
		return xdr.Asset{
			Type: xdr.AssetTypeAssetTypeCreditAlphanum4,
			AlphaNum4: &xdr.AlphaNum4{
				AssetCode: xdr.AssetCode4(c),
				Issuer:    xdr.MustAddress(issuer),
			},
		}
	}
	var c [12]byte
	copy(c[:], strings.TrimRight(code, "\x00"))
	return xdr.Asset{
		Type: xdr.AssetTypeAssetTypeCreditAlphanum12,
		AlphaNum12: &xdr.AlphaNum12{
			AssetCode: xdr.AssetCode12(c),
			Issuer:    xdr.MustAddress(issuer),
		},
	}
}

func TestTokenEventFromProto_Transfer(t *testing.T) {
	meta := makeTestMeta(100, "abc123", opIndexPtr(1))
	protoAsset := assetProto.NewProtoAsset(nativeXDRAsset())
	event := token_transfer.NewTransferEvent(meta, "GABC", "GDEF", "5000000", protoAsset)

	te, err := tokenEventFromProto(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if te.EventType != tokenEventTypeTransfer {
		t.Errorf("expected transfer type (0), got %d", te.EventType)
	}
	if te.EventTypeName != "transfer" {
		t.Errorf("expected 'transfer', got %q", te.EventTypeName)
	}
	if te.FromAddress == nil || *te.FromAddress != "GABC" {
		t.Errorf("expected from=GABC, got %v", te.FromAddress)
	}
	if te.ToAddress == nil || *te.ToAddress != "GDEF" {
		t.Errorf("expected to=GDEF, got %v", te.ToAddress)
	}
	if te.Amount != "5000000" {
		t.Errorf("expected amount=5000000, got %q", te.Amount)
	}
	if te.AssetType != assetTypeNative {
		t.Errorf("expected native asset type, got %d", te.AssetType)
	}
	if te.LedgerSequence != 100 {
		t.Errorf("expected ledger 100, got %d", te.LedgerSequence)
	}
	if te.OperationIndex == nil || *te.OperationIndex != 1 {
		t.Errorf("expected operation_index=1, got %v", te.OperationIndex)
	}
}

func TestTokenEventFromProto_Mint(t *testing.T) {
	const usdcIssuer = "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"
	meta := makeTestMeta(200, "def456", opIndexPtr(2))
	usdcXDR := creditXDRAsset("USDC", usdcIssuer)
	usdcProto := assetProto.NewProtoAsset(usdcXDR)
	event := token_transfer.NewMintEvent(meta, "GXYZ", "100000000000", usdcProto)

	te, err := tokenEventFromProto(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if te.EventType != tokenEventTypeMint {
		t.Errorf("expected mint type (1), got %d", te.EventType)
	}
	if te.FromAddress != nil {
		t.Errorf("expected nil from for mint, got %v", te.FromAddress)
	}
	if te.AssetType != assetTypeCredit {
		t.Errorf("expected credit asset type, got %d", te.AssetType)
	}
	if te.AssetCode == nil || *te.AssetCode != "USDC" {
		t.Errorf("expected USDC, got %v", te.AssetCode)
	}
	if te.AssetIssuer == nil || *te.AssetIssuer != usdcIssuer {
		t.Errorf("expected issuer %q, got %v", usdcIssuer, te.AssetIssuer)
	}
}

func TestTokenEventFromProto_SorobanToken(t *testing.T) {
	contractAddr := "CABC123XYZ"
	meta := makeTestMeta(300, "ghi789", opIndexPtr(1))
	meta.ContractAddress = contractAddr
	// Pure Soroban token: asset is nil
	event := token_transfer.NewTransferEvent(meta, "GABC", "GDEF", "999", nil)

	te, err := tokenEventFromProto(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if te.AssetType != assetTypeSorobanToken {
		t.Errorf("expected soroban_token asset type (2), got %d", te.AssetType)
	}
	if te.AssetContractID == nil || *te.AssetContractID != contractAddr {
		t.Errorf("expected contract_id=%q, got %v", contractAddr, te.AssetContractID)
	}
	if te.AssetCode != nil {
		t.Errorf("expected nil asset_code for soroban token, got %v", te.AssetCode)
	}
}

func TestTokenEventFromProto_FeeHasNoOperationIndex(t *testing.T) {
	meta := makeTestMeta(400, "jkl012", nil) // fee events have no operation index
	xlmProto := assetProto.NewProtoAsset(nativeXDRAsset())
	event := token_transfer.NewFeeEvent(meta, "GABC", "100", xlmProto)

	te, err := tokenEventFromProto(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if te.EventType != tokenEventTypeFee {
		t.Errorf("expected fee type (4), got %d", te.EventType)
	}
	if te.OperationIndex != nil {
		t.Errorf("expected nil operation_index for fee event, got %v", te.OperationIndex)
	}
}

func TestTokenEventsFromLedgerMeta_EmptyMeta(t *testing.T) {
	events, err := TokenEventsFromLedgerMeta("", network.PublicNetworkPassphrase)
	if err != nil {
		t.Fatalf("unexpected error for empty meta: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events for empty meta, got %d", len(events))
	}
}
