export * from "./stellar";

// Network types
export type NetworkKey = "mainnet" | "testnet" | "futurenet";

export interface NetworkConfig {
  name: string;
  horizonUrl: string;
  rpcUrl: string;
  passphrase: string;
}

// Entity types for search
export type EntityType = "transaction" | "account" | "contract" | "asset" | "ledger" | "unknown";

// Transaction status
export type TransactionStatus = "success" | "failed";

// Operation types from Stellar
export type OperationType =
  | "create_account"
  | "payment"
  | "path_payment_strict_receive"
  | "path_payment_strict_send"
  | "manage_sell_offer"
  | "manage_buy_offer"
  | "create_passive_sell_offer"
  | "set_options"
  | "change_trust"
  | "allow_trust"
  | "account_merge"
  | "inflation"
  | "manage_data"
  | "bump_sequence"
  | "create_claimable_balance"
  | "claim_claimable_balance"
  | "begin_sponsoring_future_reserves"
  | "end_sponsoring_future_reserves"
  | "revoke_sponsorship"
  | "clawback"
  | "clawback_claimable_balance"
  | "set_trust_line_flags"
  | "liquidity_pool_deposit"
  | "liquidity_pool_withdraw"
  | "invoke_host_function"
  | "extend_footprint_ttl"
  | "restore_footprint";

// Formatted display types
export interface FormattedAmount {
  value: string;
  code: string;
  issuer?: string;
}

export interface FormattedHash {
  full: string;
  truncated: string;
}

// Watchlist item
export interface WatchlistItem {
  type: "account" | "asset" | "contract";
  id: string;
  label?: string;
  addedAt: number;
}

// Search result
export interface SearchResult {
  type: EntityType;
  id: string;
  label?: string;
  subtitle?: string;
}

// Asset identifier
export interface AssetIdentifier {
  code: string;
  issuer: string;
}

// Pagination cursor
export interface PaginationState {
  cursor?: string;
  hasNext: boolean;
  hasPrev: boolean;
}
