"use client";

import { notFound } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/layout/page-header";
import { AssetLogo } from "@/components/common/asset-logo";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { useAsset, useAssetMetadata } from "@/lib/hooks";
import { parseAssetSlug } from "@/lib/utils";
import { Lock, Unlock, Building2, BarChart3, BookOpen, Star, Users } from "lucide-react";
import { useTranslations } from "next-intl";
import type { AssetRecordExtended } from "./sections/types";
import { AssetSummary } from "./sections/asset-summary";
import { AssetFlags } from "./sections/asset-flags";
import { AssetStats } from "./sections/asset-stats";
import { AssetHolders } from "./sections/asset-holders";
import { AssetTrades } from "./sections/asset-trades";
import { AssetOrderbook } from "./sections/asset-orderbook";
import { AssetMarketData } from "./sections/asset-market-data";

interface AssetContentProps {
  slug: string;
}

export function AssetContent({ slug }: AssetContentProps) {
  const parsed = parseAssetSlug(slug);
  const t = useTranslations("assetDetails");
  const tCommon = useTranslations("common");

  const isNative = parsed?.issuer === "native";

  const {
    data: asset,
    isLoading,
    error,
    refetch,
  } = useAsset(parsed?.code || "", isNative ? "" : parsed?.issuer || "");

  const assetRecord = asset as { _links?: { toml?: { href?: string } } } | undefined;
  const tomlUrl = assetRecord?._links?.toml?.href;
  const { data: metadata } = useAssetMetadata(asset?.asset_code, asset?.asset_issuer, tomlUrl);

  if (!parsed) {
    return notFound();
  }

  if (isNative) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <AssetLogo code="XLM" size="xl" />
          <div>
            <h1 className="text-2xl font-bold">{t("xlmTitle")}</h1>
            <p className="text-muted-foreground">{t("xlmSubtitle")}</p>
          </div>
        </div>

        <Card variant="elevated" className="border-0">
          <CardHeader>
            <CardTitle className="text-base">{t("aboutXlm")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-muted-foreground">{t("xlmDescription")}</p>
            <ul className="text-muted-foreground list-inside list-disc space-y-2">
              <li>{t("xlmPurpose1")}</li>
              <li>{t("xlmPurpose2")}</li>
              <li>{t("xlmPurpose3")}</li>
              <li>{t("xlmPurpose4")}</li>
            </ul>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="space-y-2">
          <div className="bg-muted h-8 w-32 animate-pulse rounded" />
          <div className="bg-muted h-4 w-64 animate-pulse rounded" />
        </div>
        <LoadingCard rows={6} />
      </div>
    );
  }

  if (error || !asset) {
    return (
      <div className="space-y-6">
        <PageHeader
          title={parsed.code}
          subtitle={t("notFound")}
          backHref="/"
          backLabel={tCommon("home")}
        />
        <ErrorState title={t("notFound")} message={t("notFoundMessage")} onRetry={refetch} />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Asset Header with Logo */}
      <div className="flex items-start gap-4">
        <AssetLogo
          code={asset.asset_code}
          issuer={asset.asset_issuer}
          tomlUrl={tomlUrl}
          size="xl"
        />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-bold">{metadata?.name || asset.asset_code}</h1>
            {!asset.flags.auth_required && !asset.flags.auth_revocable ? (
              <Badge variant="outline" className="border-blue-500/25 text-blue-500">
                <Unlock className="mr-1 size-3" />
                {t("openAccess")}
              </Badge>
            ) : (
              <Badge variant="outline" className="border-amber-500/25 text-amber-500">
                <Lock className="mr-1 size-3" />
                {t("restricted")}
              </Badge>
            )}
          </div>
          <p className="text-muted-foreground mt-1">
            {asset.asset_code} • {asset.asset_type.replace("credit_alphanum", "Alpha ")}
          </p>
          {metadata?.description && (
            <p className="text-muted-foreground mt-2 line-clamp-2 text-sm">
              {metadata.description}
            </p>
          )}
          {metadata?.orgName && (
            <div className="text-muted-foreground mt-2 flex items-center gap-2 text-sm">
              <Building2 className="size-4" />
              <span>
                {t("issuer")} {metadata.orgName}
              </span>
            </div>
          )}
        </div>
      </div>

      <AssetSummary asset={asset as AssetRecordExtended} />

      <Tabs defaultValue="stats" className="w-full">
        <TabsList>
          <TabsTrigger value="stats">{t("statistics")}</TabsTrigger>
          <TabsTrigger value="holders">
            <Users className="mr-1.5 size-3.5" />
            Holders
          </TabsTrigger>
          <TabsTrigger value="trades">
            <BarChart3 className="mr-1.5 size-3.5" />
            Trades
          </TabsTrigger>
          <TabsTrigger value="orderbook">
            <BookOpen className="mr-1.5 size-3.5" />
            Orderbook
          </TabsTrigger>
          <TabsTrigger value="market">
            <Star className="mr-1.5 size-3.5" />
            Market
          </TabsTrigger>
          <TabsTrigger value="flags">{t("flags")}</TabsTrigger>
        </TabsList>
        <TabsContent value="stats" className="mt-4">
          <AssetStats asset={asset as AssetRecordExtended} />
        </TabsContent>
        <TabsContent value="holders" className="mt-4">
          <AssetHolders code={asset.asset_code} issuer={asset.asset_issuer} />
        </TabsContent>
        <TabsContent value="trades" className="mt-4">
          <AssetTrades code={asset.asset_code} issuer={asset.asset_issuer} />
        </TabsContent>
        <TabsContent value="orderbook" className="mt-4">
          <AssetOrderbook code={asset.asset_code} issuer={asset.asset_issuer} />
        </TabsContent>
        <TabsContent value="market" className="mt-4">
          <AssetMarketData code={asset.asset_code} issuer={asset.asset_issuer} />
        </TabsContent>
        <TabsContent value="flags" className="mt-4">
          <AssetFlags asset={asset as AssetRecordExtended} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
