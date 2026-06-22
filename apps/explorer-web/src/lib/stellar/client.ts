import { Horizon, rpc } from "@stellar/stellar-sdk";
import { NETWORKS } from "@/lib/constants";
import type { NetworkKey } from "@/types";

export type StellarExpertResult<T> =
  | { available: true; data: T }
  | { available: false; reason: "testnet" | "futurenet" | "error" };

export interface StellarClients {
  horizon: Horizon.Server;
  rpc: rpc.Server;
}

const clientCache = new Map<NetworkKey, StellarClients>();

export function createStellarClient(network: NetworkKey): StellarClients {
  const cached = clientCache.get(network);
  if (cached) return cached;

  const config = NETWORKS[network];
  const clients: StellarClients = {
    horizon: new Horizon.Server(config.horizonUrl),
    rpc: new rpc.Server(config.rpcUrl),
  };

  clientCache.set(network, clients);
  return clients;
}

export function getHorizonClient(network: NetworkKey): Horizon.Server {
  return createStellarClient(network).horizon;
}

export function getRpcClient(network: NetworkKey): rpc.Server {
  return createStellarClient(network).rpc;
}

export async function fetchStellarExpert<T>(
  network: NetworkKey,
  endpoint: string
): Promise<StellarExpertResult<T>> {
  if (network === "testnet") return { available: false, reason: "testnet" };
  if (network === "futurenet") return { available: false, reason: "futurenet" };

  try {
    // Stellar Expert API uses "public" as the network slug, not "mainnet"
    const seNetwork = network === "mainnet" ? "public" : network;
    const path = `explorer/${seNetwork}/${endpoint}`;
    const url = `/api/stellar-expert?path=${encodeURIComponent(path)}`;
    const response = await fetch(url, { signal: AbortSignal.timeout(12_000) });

    if (!response.ok) {
      return { available: false, reason: "error" };
    }

    const data = (await response.json()) as T & { fallback?: boolean };
    if (data.fallback) return { available: false, reason: "error" };

    return { available: true, data: data as T };
  } catch {
    return { available: false, reason: "error" };
  }
}
