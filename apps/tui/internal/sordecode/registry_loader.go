package sordecode

import (
	"encoding/base64"
	"fmt"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// RegistryFromSpecXDR builds a spec registry from base64-encoded contract spec XDR bytes.
func RegistryFromSpecXDR(specXDRBase64 string) (*SpecRegistry, error) {
	specXDRBase64 = trimSpace(specXDRBase64)
	if specXDRBase64 == "" {
		return nil, fmt.Errorf("empty contract spec xdr")
	}
	specBytes, err := base64.StdEncoding.DecodeString(specXDRBase64)
	if err != nil {
		return nil, fmt.Errorf("decode contract spec xdr: %w", err)
	}
	entries, err := DecodeEntries(specBytes)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("contract spec xdr has no entries")
	}
	return NewRegistry(entries), nil
}

// RegistryFromSpecEntries builds a registry from decoded XDR spec entries.
func RegistryFromSpecEntries(entries []xdr.ScSpecEntry) *SpecRegistry {
	if len(entries) == 0 {
		return nil
	}
	return NewRegistry(entries)
}

func trimSpace(value string) string {
	start := 0
	end := len(value)
	for start < end && (value[start] == ' ' || value[start] == '\n' || value[start] == '\r' || value[start] == '\t') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\n' || value[end-1] == '\r' || value[end-1] == '\t') {
		end--
	}
	return value[start:end]
}
