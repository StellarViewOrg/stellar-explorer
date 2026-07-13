"use client";

import { useQueries } from "@tanstack/react-query";
import { getHorizonClient } from "@/lib/stellar/client";
import { useNetwork } from "@/lib/providers";
import { STALE_TIME, VALID_NETWORKS } from "@/lib/constants";
import type { EntityType, NetworkKey } from "@/types";

interface CrossNetworkResult {
  foundOnNetwork: NetworkKey | null;
  isChecking: boolean;
}

export function useCrossNetworkSearch(
  entityType: EntityType,
  entityId: string,
  enabled: boolean
): CrossNetworkResult {
  const { network: currentNetwork } = useNetwork();
  const otherNetworks = VALID_NETWORKS.filter((n) => n !== currentNetwork);

  const results = useQueries({
    queries: otherNetworks.map((network) => ({
      queryKey: ["cross-network-search", network, entityType, entityId],
      queryFn: async () => {
        const horizon = getHorizonClient(network);
        switch (entityType) {
          case "account":
            return horizon.accounts().accountId(entityId).call();
          case "contract":
            return horizon.operations().forAccount(entityId).limit(1).call();
          case "transaction":
            return horizon.transactions().transaction(entityId).call();
          case "ledger": {
            const seq = parseInt(entityId, 10);
            return horizon.ledgers().ledger(seq).call();
          }
          default:
            throw new Error("unsupported entity type for cross-network search");
        }
      },
      enabled: enabled && !!entityId,
      retry: 0,
      staleTime: STALE_TIME,
    })),
  });

  const foundIndex = results.findIndex((r) => r.isSuccess && r.data);
  const foundOnNetwork = foundIndex >= 0 ? otherNetworks[foundIndex] : null;
  const isChecking = enabled && results.some((r) => r.isLoading);

  return { foundOnNetwork, isChecking };
}
