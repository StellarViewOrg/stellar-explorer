"use client";

import { memo } from "react";
import { Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";
import { Card, CardContent } from "@/components/ui/card";
import { AssetLogo } from "@/components/common/asset-logo";
import { truncateHash, formatCompactNumber } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { CheckCircle2 } from "lucide-react";
import type { StellarAsset } from "@/types";

interface AssetCardProps {
  asset: StellarAsset;
  className?: string;
}

export const AssetCard = memo(function AssetCard({ asset, className }: AssetCardProps) {
  const t = useTranslations("cards.asset");
  const assetPath = `/asset/${asset.asset_code}-${asset.asset_issuer}`;

  return (
    <Link href={assetPath}>
      <Card variant="elevated" interactive className={cn("group border-0 py-0", className)}>
        <CardContent className="flex items-center justify-between p-4">
          <div className="flex min-w-0 items-center gap-3">
            <div className="relative">
              <AssetLogo
                code={asset.asset_code}
                issuer={asset.asset_issuer}
                tomlUrl={asset._links?.toml?.href}
                size="md"
              />
              {/* Glow effect on hover */}
              <div className="bg-chart-3/20 absolute inset-0 rounded-full opacity-0 blur-md transition-opacity duration-300 group-hover:opacity-60" />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">{asset.asset_code}</span>
                {asset.flags?.auth_required === false && asset.flags?.auth_revocable === false && (
                  <CheckCircle2 className="text-success size-3.5" />
                )}
              </div>
              <span className="text-muted-foreground text-xs">
                {truncateHash(asset.asset_issuer, 4, 4)}
              </span>
            </div>
          </div>
          <div className="shrink-0 text-right">
            <span className="text-sm font-medium tabular-nums">
              {formatCompactNumber(asset.num_accounts)}
            </span>
            <span className="text-muted-foreground ml-1 text-xs">{t("holders")}</span>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
});

export function AssetCardSkeleton({ className }: { className?: string }) {
  return (
    <Card variant="elevated" className={cn("border-0 py-0", className)}>
      <CardContent className="flex items-center justify-between p-4">
        <div className="flex items-center gap-3">
          <div className="bg-muted/50 size-10 animate-pulse rounded-xl" />
          <div className="space-y-2">
            <div className="bg-muted/50 h-4 w-16 animate-pulse rounded" />
            <div className="bg-muted/50 h-3 w-20 animate-pulse rounded" />
          </div>
        </div>
        <div className="bg-muted/50 h-4 w-16 animate-pulse rounded" />
      </CardContent>
    </Card>
  );
}
