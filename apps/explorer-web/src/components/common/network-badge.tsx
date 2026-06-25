"use client";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { NetworkKey } from "@/types";
import { useTranslations } from "next-intl";

interface NetworkBadgeProps {
  network: NetworkKey;
  className?: string;
}

const networkStyles: Record<NetworkKey, string> = {
  mainnet: "bg-success/15 text-success border-success/25",
  testnet: "bg-warning/15 text-warning border-warning/25",
  futurenet: "bg-primary/15 text-primary border-primary/25",
};

const networkLabelKeys: Record<NetworkKey, "mainnet" | "testnet" | "futurenet"> = {
  mainnet: "mainnet",
  testnet: "testnet",
  futurenet: "futurenet",
};

export function NetworkBadge({ network, className }: NetworkBadgeProps) {
  const t = useTranslations("network");

  return (
    <Badge variant="outline" className={cn(networkStyles[network], "font-medium", className)}>
      {t(networkLabelKeys[network])}
    </Badge>
  );
}
