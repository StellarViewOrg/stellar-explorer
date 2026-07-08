"use client";

import { Globe, Loader2, X, ArrowRight } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useCrossNetworkSearch } from "@/lib/hooks/use-cross-network-search";
import { useNetwork } from "@/lib/providers";
import { NETWORKS } from "@/lib/constants";
import { useTranslations } from "next-intl";
import type { EntityType } from "@/types";

interface CrossNetworkBannerProps {
  entityType: EntityType;
  entityId: string;
  enabled: boolean;
}

export function CrossNetworkBanner({ entityType, entityId, enabled }: CrossNetworkBannerProps) {
  const { network: currentNetwork, setNetwork } = useNetwork();
  const { foundOnNetwork, isChecking } = useCrossNetworkSearch(entityType, entityId, enabled);
  const [dismissed, setDismissed] = useState(false);
  const t = useTranslations("crossNetwork");
  const tEntity = useTranslations("entityTypes");

  if (!enabled || dismissed) return null;

  if (isChecking) {
    return (
      <div className="border-border/50 bg-muted/40 flex items-center gap-3 rounded-lg border px-4 py-3 text-sm">
        <Loader2 className="text-muted-foreground size-4 shrink-0 animate-spin" />
        <span className="text-muted-foreground">{t("checking")}</span>
      </div>
    );
  }

  if (!foundOnNetwork) return null;

  const foundNetworkName = NETWORKS[foundOnNetwork].name;
  const currentNetworkName = NETWORKS[currentNetwork].name;
  const entityLabel = tEntity(entityType);

  return (
    <div className="border-primary/30 bg-primary/5 flex items-start gap-3 rounded-lg border px-4 py-3 text-sm">
      <Globe className="text-primary mt-0.5 size-4 shrink-0" />
      <div className="flex-1 space-y-1">
        <p className="text-foreground font-medium">{t("foundOn", { network: foundNetworkName })}</p>
        <p className="text-muted-foreground">
          {t("suggestion", {
            entity: entityLabel,
            currentNetwork: currentNetworkName,
            foundNetwork: foundNetworkName,
          })}
        </p>
      </div>
      <div className="flex shrink-0 items-center gap-2">
        <Button
          size="sm"
          variant="default"
          onClick={() => setNetwork(foundOnNetwork)}
          className="gap-1.5 text-xs"
        >
          {t("switchTo", { network: foundNetworkName })}
          <ArrowRight className="size-3" />
        </Button>
        <Button
          size="icon"
          variant="ghost"
          className="size-7"
          onClick={() => setDismissed(true)}
          aria-label={t("dismiss")}
        >
          <X className="size-3.5" />
        </Button>
      </div>
    </div>
  );
}
