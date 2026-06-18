package transform

import "testing"

func TestOperationDetailsSpecImproved(t *testing.T) {
	before := `{"spec_decode_status":"spec_unavailable","arguments":["arg1=World"]}`
	after := `{"spec_decode_status":"decoded","arguments_decoded":[{"name":"to","value":"World"}]}`
	if !OperationDetailsSpecImproved(before, after) {
		t.Fatal("expected operation details to be considered improved")
	}
	if OperationDetailsSpecImproved(after, after) {
		t.Fatal("expected identical details to skip backfill")
	}
}

func TestEventValueSpecImproved(t *testing.T) {
	before := `[]`
	after := `{"text":"[]","event_name":"hello","fields":[],"spec_decode_status":"decoded"}`
	beforePtr := &before
	afterPtr := &after
	if !EventValueSpecImproved(beforePtr, afterPtr) {
		t.Fatal("expected event payload to be considered improved")
	}
}
