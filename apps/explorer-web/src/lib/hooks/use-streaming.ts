"use client";

import { useEffect, useCallback, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useNetwork } from "@/lib/providers";
import { getHorizonClient } from "@/lib/stellar";
import { stellarKeys } from "@/lib/stellar/queries";
import type { Horizon } from "@stellar/stellar-sdk";

interface StreamingOptions {
  enabled?: boolean;
  onMessage?: (data: unknown) => void;
  onError?: (error: Error) => void;
}

interface StreamingState {
  isConnected: boolean;
  error: Error | null;
}

/**
 * Hook for streaming ledger updates in real-time
 */
export function useLedgerStream(options: StreamingOptions = {}) {
  const { enabled = true, onMessage, onError } = options;
  const { network } = useNetwork();
  const queryClient = useQueryClient();
  const closeRef = useRef<(() => void) | null>(null);
  const disconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [state, setState] = useState<StreamingState>({
    isConnected: false,
    error: null,
  });

  const connect = useCallback(() => {
    if (!enabled) return;

    try {
      const horizon = getHorizonClient(network);

      // Close existing connection
      if (closeRef.current) {
        closeRef.current();
      }

      // Start streaming from the current cursor
      const close = horizon
        .ledgers()
        .cursor("now")
        .stream({
          onmessage: (ledger: Horizon.ServerApi.LedgerRecord) => {
            // Clear any pending disconnect timeout
            if (disconnectTimeoutRef.current) {
              clearTimeout(disconnectTimeoutRef.current);
              disconnectTimeoutRef.current = null;
            }

            // Update the latest ledger in the query cache
            queryClient.setQueryData(stellarKeys.latestLedger(network), ledger);

            // Call custom onMessage handler
            onMessage?.(ledger);

            setState((prev) => ({
              ...prev,
              isConnected: true,
              error: null,
            }));
          },
          onerror: (event: MessageEvent) => {
            // Only treat as error if there's actual error data
            // Empty events are normal disconnections (SSE reconnect behavior)
            const hasErrorData = event?.data?.message || event?.data?.error;
            if (hasErrorData) {
              const error = new Error(event.data?.message || "Stream error");
              onError?.(error);
              setState((prev) => ({
                ...prev,
                isConnected: false,
                error,
              }));
            } else {
              // Debounce silent reconnections to avoid flickering
              if (!disconnectTimeoutRef.current) {
                disconnectTimeoutRef.current = setTimeout(() => {
                  setState((prev) => ({
                    ...prev,
                    isConnected: false,
                  }));
                  disconnectTimeoutRef.current = null;
                }, 8000);
              }
            }
          },
        });

      closeRef.current = close;
      setState((prev) => ({ ...prev, isConnected: true, error: null }));
    } catch (error) {
      const err = error instanceof Error ? error : new Error("Stream connection failed");
      setState((prev) => ({ ...prev, isConnected: false, error: err }));
      onError?.(err);
    }
  }, [enabled, network, queryClient, onMessage, onError]);

  const disconnect = useCallback(() => {
    if (closeRef.current) {
      closeRef.current();
      closeRef.current = null;
      setState((prev) => ({ ...prev, isConnected: false }));
    }
  }, []);

  // Connect on mount and when network changes
  useEffect(() => {
    if (!enabled) return;

    // Use microtask to avoid synchronous setState in effect
    queueMicrotask(() => {
      connect();
    });

    return () => {
      disconnect();
    };
  }, [enabled, network, connect, disconnect]);

  return {
    ...state,
    connect,
    disconnect,
  };
}

/**
 * Hook for streaming transaction updates in real-time
 */
export function useTransactionStream(options: StreamingOptions = {}) {
  const { enabled = true, onMessage, onError } = options;
  const { network } = useNetwork();
  const queryClient = useQueryClient();
  const closeRef = useRef<(() => void) | null>(null);
  const disconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [state, setState] = useState<StreamingState>({
    isConnected: false,
    error: null,
  });

  const connect = useCallback(() => {
    if (!enabled) return;

    try {
      const horizon = getHorizonClient(network);

      // Close existing connection
      if (closeRef.current) {
        closeRef.current();
      }

      // Start streaming transactions
      const close = horizon
        .transactions()
        .cursor("now")
        .stream({
          onmessage: (transaction: Horizon.ServerApi.TransactionRecord) => {
            // Clear any pending disconnect timeout
            if (disconnectTimeoutRef.current) {
              clearTimeout(disconnectTimeoutRef.current);
              disconnectTimeoutRef.current = null;
            }

            // Invalidate recent transactions and fee stats to refetch
            queryClient.invalidateQueries({
              queryKey: stellarKeys.transactions(network),
            });
            queryClient.invalidateQueries({
              queryKey: stellarKeys.feeStats(network),
            });

            // Call custom onMessage handler
            onMessage?.(transaction);

            setState((prev) => ({
              ...prev,
              isConnected: true,
              error: null,
            }));
          },
          onerror: (event: MessageEvent) => {
            const hasErrorData = event?.data?.message || event?.data?.error;
            if (hasErrorData) {
              const error = new Error(event.data?.message || "Stream error");
              onError?.(error);
              setState((prev) => ({
                ...prev,
                isConnected: false,
                error,
              }));
            } else {
              // Debounce silent reconnections to avoid flickering
              if (!disconnectTimeoutRef.current) {
                disconnectTimeoutRef.current = setTimeout(() => {
                  setState((prev) => ({
                    ...prev,
                    isConnected: false,
                  }));
                  disconnectTimeoutRef.current = null;
                }, 8000);
              }
            }
          },
        });

      closeRef.current = close;
      setState((prev) => ({ ...prev, isConnected: true, error: null }));
    } catch (error) {
      const err = error instanceof Error ? error : new Error("Stream connection failed");
      setState((prev) => ({ ...prev, isConnected: false, error: err }));
      onError?.(err);
    }
  }, [enabled, network, queryClient, onMessage, onError]);

  const disconnect = useCallback(() => {
    if (closeRef.current) {
      closeRef.current();
      closeRef.current = null;
      setState((prev) => ({ ...prev, isConnected: false }));
    }
  }, []);

  useEffect(() => {
    if (!enabled) return;

    // Use microtask to avoid synchronous setState in effect
    queueMicrotask(() => {
      connect();
    });

    return () => {
      disconnect();
    };
  }, [enabled, network, connect, disconnect]);

  return {
    ...state,
    connect,
    disconnect,
  };
}

/**
 * Hook for streaming account operations in real-time
 */
export function useAccountOperationsStream(accountId: string, options: StreamingOptions = {}) {
  const { enabled = true, onMessage, onError } = options;
  const { network } = useNetwork();
  const queryClient = useQueryClient();
  const closeRef = useRef<(() => void) | null>(null);
  const disconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [state, setState] = useState<StreamingState>({
    isConnected: false,
    error: null,
  });

  const isValidAccount = accountId?.startsWith("G") && accountId?.length === 56;

  const connect = useCallback(() => {
    if (!enabled || !isValidAccount) return;

    try {
      const horizon = getHorizonClient(network);

      // Close existing connection
      if (closeRef.current) {
        closeRef.current();
      }

      // Start streaming operations for this account
      const close = horizon
        .operations()
        .forAccount(accountId)
        .cursor("now")
        .stream({
          onmessage: (operation: Horizon.ServerApi.OperationRecord) => {
            // Clear any pending disconnect timeout
            if (disconnectTimeoutRef.current) {
              clearTimeout(disconnectTimeoutRef.current);
              disconnectTimeoutRef.current = null;
            }

            // Invalidate account queries — operations, balances, and offers may all change
            queryClient.invalidateQueries({
              queryKey: stellarKeys.accountOperations(network, accountId),
            });
            queryClient.invalidateQueries({
              queryKey: stellarKeys.account(network, accountId),
            });
            queryClient.invalidateQueries({
              queryKey: stellarKeys.accountOffers(network, accountId),
            });

            // Call custom onMessage handler
            onMessage?.(operation);

            setState((prev) => ({
              ...prev,
              isConnected: true,
              error: null,
            }));
          },
          onerror: (event: MessageEvent) => {
            const hasErrorData = event?.data?.message || event?.data?.error;
            if (hasErrorData) {
              const error = new Error(event.data?.message || "Stream error");
              onError?.(error);
              setState((prev) => ({
                ...prev,
                isConnected: false,
                error,
              }));
            } else {
              // Debounce silent reconnections to avoid flickering
              if (!disconnectTimeoutRef.current) {
                disconnectTimeoutRef.current = setTimeout(() => {
                  setState((prev) => ({
                    ...prev,
                    isConnected: false,
                  }));
                  disconnectTimeoutRef.current = null;
                }, 8000);
              }
            }
          },
        });

      closeRef.current = close;
      setState((prev) => ({ ...prev, isConnected: true, error: null }));
    } catch (error) {
      const err = error instanceof Error ? error : new Error("Stream connection failed");
      setState((prev) => ({ ...prev, isConnected: false, error: err }));
      onError?.(err);
    }
  }, [enabled, isValidAccount, accountId, network, queryClient, onMessage, onError]);

  const disconnect = useCallback(() => {
    if (closeRef.current) {
      closeRef.current();
      closeRef.current = null;
      setState((prev) => ({ ...prev, isConnected: false }));
    }
  }, []);

  useEffect(() => {
    if (!enabled || !isValidAccount) return;

    // Use microtask to avoid synchronous setState in effect
    queueMicrotask(() => {
      connect();
    });

    return () => {
      disconnect();
    };
  }, [enabled, isValidAccount, accountId, network, connect, disconnect]);

  return {
    ...state,
    connect,
    disconnect,
  };
}
