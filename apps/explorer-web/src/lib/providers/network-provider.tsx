"use client";

import { createContext, useContext, type ReactNode } from "react";
import { useParams } from "next/navigation";
import type { NetworkKey } from "@/types";
import { DEFAULT_NETWORK, NETWORKS, VALID_NETWORKS } from "@/lib/constants";
import { usePathname } from "@/i18n/navigation";

interface NetworkContextValue {
  network: NetworkKey;
  setNetwork: (network: NetworkKey) => void;
  networkConfig: (typeof NETWORKS)[NetworkKey];
}

const NetworkContext = createContext<NetworkContextValue | null>(null);

export function NetworkProvider({ children }: { children: ReactNode }) {
  const params = useParams();
  const pathname = usePathname();

  // Read network from URL params, fallback to default
  const rawNetwork = params?.network as string | undefined;
  const network: NetworkKey =
    rawNetwork && VALID_NETWORKS.includes(rawNetwork as NetworkKey)
      ? (rawNetwork as NetworkKey)
      : DEFAULT_NETWORK;

  const setNetwork = (newNetwork: NetworkKey) => {
    const newPath = `/${newNetwork}${pathname}`;
    window.location.href = `/${(params?.locale as string) || "en"}${newPath}`;
  };

  const value: NetworkContextValue = {
    network,
    setNetwork,
    networkConfig: NETWORKS[network],
  };

  return <NetworkContext.Provider value={value}>{children}</NetworkContext.Provider>;
}

export function useNetwork() {
  const context = useContext(NetworkContext);
  if (!context) {
    throw new Error("useNetwork must be used within a NetworkProvider");
  }
  return context;
}
