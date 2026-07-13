"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Coins, Users, TrendingUp } from "lucide-react";
import { HashDisplay } from "@/components/common/hash-display";
import { formatNumber, formatCompactNumber } from "@/lib/utils";
import { useTranslations } from "next-intl";
import type { AssetRecordExtended } from "./types";

export function AssetSummary({ asset }: { asset: AssetRecordExtended }) {
  const t = useTranslations("assetDetails");

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{t("information")}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          {/* Left column - Stats */}
          <div className="space-y-4">
            <div className="bg-primary/5 border-primary/10 rounded-lg border p-4">
              <div className="text-muted-foreground mb-1 flex items-center gap-2 text-sm">
                <Coins className="size-4" />
                {t("totalSupply")}
              </div>
              <div className="text-2xl font-semibold tabular-nums">
                {formatNumber(asset.amount)}
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="bg-card/50 rounded-lg p-4">
                <div className="text-muted-foreground mb-1 flex items-center gap-2 text-sm">
                  <Users className="size-4" />
                  {t("accounts")}
                </div>
                <div className="text-xl font-semibold tabular-nums">
                  {formatCompactNumber(asset.num_accounts)}
                </div>
              </div>

              <div className="bg-card/50 rounded-lg p-4">
                <div className="text-muted-foreground mb-1 flex items-center gap-2 text-sm">
                  <TrendingUp className="size-4" />
                  {t("claimableBalances")}
                </div>
                <div className="text-xl font-semibold tabular-nums">
                  {formatCompactNumber(asset.claimable_balances_amount || "0")}
                </div>
              </div>
            </div>
          </div>

          {/* Right column - Details */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("assetCode")}</span>
              <span className="font-medium">{asset.asset_code}</span>
            </div>
            <Separator />
            <div className="flex items-start justify-between">
              <span className="text-muted-foreground text-sm">{t("issuer")}</span>
              <HashDisplay
                hash={asset.asset_issuer}
                truncate
                linkTo={`/account/${asset.asset_issuer}`}
                className="text-sm"
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("assetType")}</span>
              <Badge variant="secondary">{asset.asset_type}</Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("liquidityPools")}</span>
              <span className="text-sm font-medium">{asset.num_liquidity_pools || 0}</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
