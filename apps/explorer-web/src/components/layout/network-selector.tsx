"use client";

import { Check, ChevronDown, Globe, TestTube, Rocket } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useNetwork } from "@/lib/providers";
import { NETWORKS } from "@/lib/constants";
import type { NetworkKey } from "@/types";
import { cn } from "@/lib/utils";

const networkIcons: Record<NetworkKey, typeof Globe> = {
  mainnet: Globe,
  testnet: TestTube,
  futurenet: Rocket,
};

const networkColors: Record<NetworkKey, string> = {
  mainnet: "text-success",
  testnet: "text-warning",
  futurenet: "text-primary",
};

export function NetworkSelector() {
  const { network, setNetwork } = useNetwork();
  const Icon = networkIcons[network];

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="min-w-[148px] gap-2">
          <Icon className={cn("size-4", networkColors[network])} />
          <span className="hidden sm:inline">{NETWORKS[network].name}</span>
          <ChevronDown className="size-3 opacity-50" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        {(Object.keys(NETWORKS) as NetworkKey[]).map((key) => {
          const NetworkIcon = networkIcons[key];
          const isActive = network === key;

          return (
            <DropdownMenuItem key={key} onClick={() => setNetwork(key)} className="gap-3">
              <NetworkIcon className={cn("size-4", networkColors[key])} />
              <span className="flex-1">{NETWORKS[key].name}</span>
              {isActive && <Check className="text-success size-4" />}
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
