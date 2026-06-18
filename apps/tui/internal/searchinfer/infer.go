package searchinfer

import (
	"strconv"
	"strings"

	"github.com/stellar/go-stellar-sdk/strkey"
)

// Candidate is a locally inferred search hit before source labeling.
type Candidate struct {
	Kind        string
	Title       string
	Description string
	Command     string
}

// FromQuery infers executable lookup commands from partial or complete Stellar identifiers.
func FromQuery(query string) []Candidate {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	switch {
	case isTransactionHash(query):
		return []Candidate{lookupCandidate("transaction", "Transaction", query, "lookup tx "+query)}
	case isPartialTransactionHash(query):
		return []Candidate{lookupCandidate("transaction", "Transaction (partial)", query, "lookup tx "+query)}
	case isAccountAddress(query):
		return []Candidate{lookupCandidate("account", "Account", query, "lookup account "+query)}
	case isContractAddress(query):
		return []Candidate{lookupCandidate("contract", "Contract", query, "lookup contract "+query)}
	case isLedgerSequence(query):
		return []Candidate{lookupCandidate("ledger", "Ledger", query, "lookup ledger "+query)}
	case isAssetQuery(query):
		return []Candidate{lookupCandidate("asset", "Asset", query, "lookup asset "+query)}
	default:
		return nil
	}
}

func lookupCandidate(kind, title, query, command string) Candidate {
	return Candidate{
		Kind:        kind,
		Title:       title,
		Description: truncateValue(query),
		Command:     command,
	}
}

func isTransactionHash(value string) bool {
	if len(value) != 64 {
		return false
	}
	return isHexString(value)
}

func isPartialTransactionHash(value string) bool {
	if len(value) < 8 || len(value) >= 64 {
		return false
	}
	return isHexString(value)
}

func isAccountAddress(value string) bool {
	_, err := strkey.Decode(strkey.VersionByteAccountID, value)
	return err == nil
}

func isContractAddress(value string) bool {
	_, err := strkey.Decode(strkey.VersionByteContract, value)
	return err == nil
}

func isLedgerSequence(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	return err == nil && parsed > 0
}

func isAssetQuery(value string) bool {
	code, issuer, ok := strings.Cut(value, ":")
	if !ok {
		return false
	}
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || len(code) > 12 || !strings.HasPrefix(issuer, "G") {
		return false
	}
	_, err := strkey.Decode(strkey.VersionByteAccountID, issuer)
	return err == nil
}

func isHexString(value string) bool {
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func truncateValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 20 {
		return value
	}
	return value[:10] + "..." + value[len(value)-6:]
}
