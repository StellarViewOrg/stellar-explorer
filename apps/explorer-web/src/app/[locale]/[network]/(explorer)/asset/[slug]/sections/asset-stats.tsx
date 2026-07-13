"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCompactNumber } from "@/lib/utils";
import { useTranslations } from "next-intl";
import type { AssetRecordExtended } from "./types";

export function AssetStats({ asset }: { asset: AssetRecordExtended }) {
  const t = useTranslations("assetDetails");

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{t("statistics")}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-4">
          <div className="bg-card/50 rounded-lg p-4 text-center">
            <div className="text-2xl font-semibold tabular-nums">
              {formatCompactNumber(asset.num_accounts)}
            </div>
            <div className="text-muted-foreground mt-1 text-xs">{t("totalHolders")}</div>
          </div>
          <div className="bg-card/50 rounded-lg p-4 text-center">
            <div className="text-2xl font-semibold tabular-nums">
              {formatCompactNumber(asset.amount)}
            </div>
            <div className="text-muted-foreground mt-1 text-xs">{t("circulating")}</div>
          </div>
          <div className="bg-card/50 rounded-lg p-4 text-center">
            <div className="text-2xl font-semibold tabular-nums">
              {formatCompactNumber(asset.num_liquidity_pools || 0)}
            </div>
            <div className="text-muted-foreground mt-1 text-xs">{t("liquidityPools")}</div>
          </div>
          <div className="bg-card/50 rounded-lg p-4 text-center">
            <div className="text-2xl font-semibold tabular-nums">
              {formatCompactNumber(asset.accounts?.authorized || 0)}
            </div>
            <div className="text-muted-foreground mt-1 text-xs">{t("authorized")}</div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
