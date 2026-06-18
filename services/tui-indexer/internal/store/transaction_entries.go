package store

import (
	"context"
	"strings"

	"github.com/lib/pq"
)

// LoadTransactionEntriesByHash returns envelope and meta XDR for the requested transaction hashes.
func (s *PostgresStore) LoadTransactionEntriesByHash(ctx context.Context, hashes []string) (map[string]ReadTransactionDetail, error) {
	if len(hashes) == 0 {
		return map[string]ReadTransactionDetail{}, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT hash, envelope_xdr, result_meta_xdr
		FROM transactions
		WHERE hash = ANY($1)`, pq.Array(hashes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make(map[string]ReadTransactionDetail, len(hashes))
	for rows.Next() {
		var detail ReadTransactionDetail
		if err := rows.Scan(&detail.Hash, &detail.EnvelopeXDR, &detail.ResultMetaXDR); err != nil {
			return nil, err
		}
		detail.Hash = strings.TrimSpace(detail.Hash)
		if detail.Hash == "" {
			continue
		}
		entries[detail.Hash] = detail
	}
	return entries, rows.Err()
}
