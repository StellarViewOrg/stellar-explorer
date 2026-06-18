package sordecode

import (
	"github.com/stellar/go-stellar-sdk/xdr"
)

// DecodeEntries decodes concatenated contract spec XDR bytes into spec entries.
func DecodeEntries(specBytes []byte) ([]xdr.ScSpecEntry, error) {
	var entries []xdr.ScSpecEntry
	decoder := xdr.NewBytesDecoder()
	remaining := specBytes
	for len(remaining) > 0 {
		var entry xdr.ScSpecEntry
		n, err := decoder.DecodeBytes(&entry, remaining)
		if err != nil || n == 0 {
			break
		}
		entries = append(entries, entry)
		remaining = remaining[n:]
	}
	return entries, nil
}
