"use client";

import { useQuery } from "@tanstack/react-query";
import { useNetwork } from "@/lib/providers";
import { LIVE_LEDGER_POLL_INTERVAL } from "@/lib/constants";
import { stellarQueries } from "@/lib/stellar";

// Hook for latest ledger (polls while mounted)
export function useLatestLedger() {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.latestLedger(network),
    refetchInterval: LIVE_LEDGER_POLL_INTERVAL,
  });
}

// Hook for a specific ledger
export function useLedger(sequence: number) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.ledger(network, sequence),
    enabled: sequence > 0,
  });
}

// Hook for ledger transactions
export function useLedgerTransactions(sequence: number, limit?: number) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.ledgerTransactions(network, sequence, limit),
    enabled: sequence > 0,
  });
}

// Hook for recent transactions (polls while mounted)
export function useRecentTransactions(limit?: number) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.recentTransactions(network, limit),
    refetchInterval: LIVE_LEDGER_POLL_INTERVAL,
  });
}

// Hook for a specific transaction
export function useTransaction(hash: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.transaction(network, hash),
    enabled: !!hash && hash.length === 64,
  });
}

// Hook for transaction operations
export function useTransactionOperations(hash: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.transactionOperations(network, hash),
    enabled: !!hash && hash.length === 64,
  });
}

// Hook for transaction effects
export function useTransactionEffects(hash: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.transactionEffects(network, hash),
    enabled: !!hash && hash.length === 64,
  });
}

// Hook for account details
export function useAccount(id: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.account(network, id),
    enabled: !!id && id.startsWith("G") && id.length === 56,
  });
}

// Hook for account transactions
export function useAccountTransactions(id: string, cursor?: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.accountTransactions(network, id, cursor),
    enabled: !!id && id.startsWith("G") && id.length === 56,
  });
}

// Hook for account operations
export function useAccountOperations(id: string, cursor?: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.accountOperations(network, id, cursor),
    enabled: !!id && id.startsWith("G") && id.length === 56,
  });
}

// Fetches last operations for an account that may no longer exist (e.g. merged accounts)
export function useAccountLastOperations(id: string, enabled: boolean) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.accountLastOperations(network, id),
    enabled: enabled && !!id && id.startsWith("G") && id.length === 56,
  });
}

// Hook for asset details
export function useAsset(code: string, issuer: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.asset(network, code, issuer),
    enabled: !!code && !!issuer,
  });
}

// Hook for fee stats (polls while mounted)
export function useFeeStats() {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.feeStats(network),
    refetchInterval: LIVE_LEDGER_POLL_INTERVAL,
  });
}

// Hook for contract info
export function useContractInfo(contractId: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.contractInfo(network, contractId),
    enabled: !!contractId && contractId.startsWith("C") && contractId.length === 56,
  });
}

// Hook for contract events (with live polling)
export function useContractEvents(contractId: string, startLedger?: number, live = false) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.contractEvents(network, contractId, startLedger),
    enabled: !!contractId && contractId.startsWith("C") && contractId.length === 56,
    refetchInterval: live ? 5000 : false,
  });
}

// Hook for contract invocations via Horizon (polls while mounted)
export function useContractInvocations(contractId: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.contractInvocations(network, contractId),
    enabled: !!contractId && contractId.startsWith("C") && contractId.length === 56,
    refetchInterval: LIVE_LEDGER_POLL_INTERVAL,
  });
}

// Hook for contract XLM balance
export function useContractBalance(contractId: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.contractBalance(network, contractId),
    enabled: !!contractId && contractId.startsWith("C") && contractId.length === 56,
  });
}

// Hook for contract code (WASM)
export function useContractCode(contractId: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.contractCode(network, contractId),
    enabled: !!contractId && contractId.startsWith("C") && contractId.length === 56,
  });
}

// Hook for contract storage
export function useContractStorage(contractId: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.contractStorage(network, contractId),
    enabled: !!contractId && contractId.startsWith("C") && contractId.length === 56,
  });
}

// Hook for assets list with pagination
export function useAssetsList(cursor?: string) {
  const { network } = useNetwork();
  return useQuery(stellarQueries.assetsList(network, cursor));
}

// Hook for accounts holding a specific asset (top holders)
export function useAssetAccounts(code: string, issuer: string) {
  const { network } = useNetwork();
  const isNative = code === "XLM" && issuer === "native";
  return useQuery({
    ...stellarQueries.assetAccounts(network, code, issuer),
    enabled: !isNative && !!code && !!issuer,
  });
}

// Hook for asset trade aggregations (24h volume, price change)
export function useAssetTrades(code: string, issuer: string) {
  const { network } = useNetwork();
  const isNative = code === "XLM" && issuer === "native";
  return useQuery({
    ...stellarQueries.assetTradeAggregations(network, code, issuer),
    enabled: !isNative && !!code && !!issuer,
  });
}

// Hook for asset orderbook
export function useAssetOrderbook(sellingCode: string, sellingIssuer: string) {
  const { network } = useNetwork();
  const isNative = sellingCode === "XLM" && sellingIssuer === "native";
  return useQuery({
    ...stellarQueries.assetOrderbook(network, sellingCode, sellingIssuer),
    enabled: !isNative && !!sellingCode && !!sellingIssuer,
  });
}

// Hook for top assets
export function useTopAssets() {
  const { network } = useNetwork();
  return useQuery(stellarQueries.topAssets(network));
}

// Hook for account offers (open DEX orders)
export function useAccountOffers(id: string, cursor?: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.accountOffers(network, id, cursor),
    enabled: !!id && id.startsWith("G") && id.length === 56,
  });
}

// Hook for account data entries (key-value pairs stored on account)
export function useAccountDataEntries(id: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.accountDataEntries(network, id),
    enabled: !!id && id.startsWith("G") && id.length === 56,
  });
}

// Hook for a single liquidity pool by ID
export function useLiquidityPool(id: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.liquidityPool(network, id),
    enabled: !!id && id.length === 64,
  });
}

// Hook for paginated list of all liquidity pools
export function useLiquidityPools(cursor?: string) {
  const { network } = useNetwork();
  return useQuery(stellarQueries.liquidityPoolsList(network, cursor));
}

// Hook for liquidity pools containing a specific asset
export function useAssetLiquidityPools(code: string, issuer: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.assetLiquidityPools(network, code, issuer),
    enabled: !!code && !!issuer,
  });
}

// Hook for transactions in a liquidity pool
export function useLiquidityPoolTransactions(id: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.liquidityPoolTransactions(network, id),
    enabled: !!id && id.length === 64,
  });
}

// Hook for claimable balances (optionally filtered by asset)
export function useClaimableBalances(assetCode?: string, assetIssuer?: string, cursor?: string) {
  const { network } = useNetwork();
  return useQuery(stellarQueries.claimableBalancesList(network, assetCode, assetIssuer, cursor));
}

// Hook for claimable balances for a specific asset
export function useAssetClaimableBalances(code: string, issuer: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.assetClaimableBalances(network, code, issuer),
    enabled: !!code && !!issuer,
  });
}

// Hook for Soroban RPC network info
export function useRpcNetworkInfo() {
  const { network } = useNetwork();
  return useQuery(stellarQueries.rpcNetworkInfo(network));
}

// Hook for Stellar Expert asset analytics (public network only)
export function useAssetAnalytics(code: string, issuer: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.assetAnalytics(network, code, issuer),
    enabled: !!code && !!issuer && code !== "XLM",
  });
}

// Hook for Stellar Expert account profile with creation date (mainnet only)
export function useAccountProfile(id: string) {
  const { network } = useNetwork();
  return useQuery({
    ...stellarQueries.accountProfile(network, id),
    enabled: network === "mainnet" && !!id && id.startsWith("G") && id.length === 56,
  });
}

// Hook for Stellar Expert network activity history (public network only)
export function useNetworkActivity() {
  const { network } = useNetwork();
  return useQuery(stellarQueries.networkActivity(network));
}

// Hook for Stellar Expert global network stats (public network only)
export function useNetworkStats() {
  const { network } = useNetwork();
  return useQuery(stellarQueries.networkStats(network));
}
