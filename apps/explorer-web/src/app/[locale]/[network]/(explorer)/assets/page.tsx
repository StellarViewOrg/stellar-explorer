"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/layout/page-header";
import { AssetLogo } from "@/components/common/asset-logo";
import { AssetTable, AssetTableSkeleton, type AssetData } from "@/components/assets";
import { ErrorState } from "@/components/common/error-state";
import { useNetwork } from "@/lib/providers";
import { NetworkBadge } from "@/components/common/network-badge";
import { useTopAssets, useAssetsList } from "@/lib/hooks";
import { POPULAR_ASSETS } from "@/lib/constants";
import { Coins, Search, TrendingUp, Star, ArrowRight } from "lucide-react";
import { useRouter } from "@/i18n/navigation";
import { Link } from "@/i18n/navigation";

// Native XLM asset for display
const XLM_ASSET = {
  code: "XLM",
  issuer: "native",
  name: "Stellar Lumens",
  description: "Native Asset",
  href: "/asset/XLM-native",
};

// Known popular assets with metadata
const POPULAR_ASSETS_META: Record<string, { name: string; description: string; tomlUrl?: string }> =
  {
    USDC: {
      name: "USD Coin",
      description: "Circle",
      tomlUrl: "https://www.centre.io/.well-known/stellar.toml",
    },
    yXLM: {
      name: "Ultra Stellar",
      description: "Liquid Staking",
      tomlUrl: "https://ultrastellar.com/.well-known/stellar.toml",
    },
    AQUA: {
      name: "Aquarius",
      description: "Liquidity Rewards",
      tomlUrl: "https://aqua.network/.well-known/stellar.toml",
    },
    SHX: { name: "Stronghold", description: "Asset Platform" },
    EURC: { name: "Euro Coin", description: "Circle" },
    BTC: { name: "Bitcoin", description: "Wrapped BTC" },
  };

function TopAssetsSection() {
  const t = useTranslations("assets");
  const { data: topAssets, isLoading, error, refetch } = useTopAssets();

  if (error) {
    return <ErrorState title={t("failedToLoad")} message={error.message} onRetry={refetch} />;
  }

  if (isLoading) {
    return <AssetTableSkeleton rows={6} />;
  }

  // Transform data for table
  const assetsData: AssetData[] = (topAssets || [])
    .filter((a): a is NonNullable<typeof a> => a !== null)
    .map((asset) => ({
      code: asset.code,
      issuer: asset.issuer,
      assetType: asset.assetType,
      numAccounts: asset.numAccounts,
      amount: asset.amount,
      volume24h: asset.volume24h,
      priceChange24h: asset.priceChange24h,
      currentPrice: asset.currentPrice,
      flags: asset.flags,
    }));

  return <AssetTable assets={assetsData} showRank={true} title={t("topAssets")} />;
}

function extractCursor(href?: string): string | undefined {
  if (!href) return undefined;
  try {
    return new URL(href).searchParams.get("cursor") ?? undefined;
  } catch {
    return undefined;
  }
}

function AllAssetsList() {
  const t = useTranslations("assets");
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [history, setHistory] = useState<string[]>([]);
  const { data, isLoading, error, refetch } = useAssetsList(cursor);

  if (error) {
    return <ErrorState title={t("failedToLoad")} message={error.message} onRetry={refetch} />;
  }

  if (isLoading) {
    return <AssetTableSkeleton rows={10} />;
  }

  const assetsData: AssetData[] = (data?.records || []).map((record) => ({
    code: record.asset_code,
    issuer: record.asset_issuer,
    assetType: record.asset_type,
    numAccounts: record.accounts.authorized + record.accounts.authorized_to_maintain_liabilities,
    amount: parseFloat(record.balances.authorized),
    flags: record.flags,
  }));

  const nextCursor = extractCursor(
    (data as { _links?: { next?: { href?: string } } })?._links?.next?.href
  );

  const handleNext = () => {
    if (!nextCursor) return;
    setHistory((h) => [...h, cursor ?? ""]);
    setCursor(nextCursor);
  };

  const handlePrev = () => {
    const prev = [...history];
    const last = prev.pop();
    setHistory(prev);
    setCursor(last === "" ? undefined : last);
  };

  return (
    <div className="space-y-4">
      <AssetTable assets={assetsData} showRank={false} title={t("allAssets")} />
      <div className="flex items-center justify-between">
        <Button variant="outline" size="sm" onClick={handlePrev} disabled={history.length === 0}>
          ← Previous
        </Button>
        <span className="text-muted-foreground text-sm">Page {history.length + 1}</span>
        <Button variant="outline" size="sm" onClick={handleNext} disabled={!nextCursor}>
          Next →
        </Button>
      </div>
    </div>
  );
}

function AssetSearchForm() {
  const t = useTranslations("assets");
  const router = useRouter();
  const [assetCode, setAssetCode] = useState("");
  const [issuer, setIssuer] = useState("");
  const [error, setError] = useState("");

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!assetCode.trim()) {
      setError(t("enterCodeError"));
      return;
    }
    if (!issuer.trim() && assetCode.toUpperCase() !== "XLM") {
      setError(t("enterIssuerError"));
      return;
    }

    if (assetCode.toUpperCase() === "XLM") {
      router.push("/asset/XLM-native");
    } else {
      router.push(`/asset/${assetCode.trim()}-${issuer.trim()}`);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{t("findAsset")}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSearch} className="space-y-4">
          <div className="grid grid-cols-1 gap-2 md:grid-cols-3">
            <Input
              placeholder={t("assetCode")}
              aria-label={t("assetCode")}
              value={assetCode}
              onChange={(e) => {
                setAssetCode(e.target.value);
                setError("");
              }}
            />
            <Input
              placeholder={t("issuerAddress")}
              aria-label={t("issuerAddress")}
              value={issuer}
              onChange={(e) => {
                setIssuer(e.target.value);
                setError("");
              }}
              className="font-mono md:col-span-2"
            />
          </div>
          {error && <p className="text-destructive text-sm">{error}</p>}
          <Button type="submit">
            <Search className="mr-2 size-4" />
            {t("search")}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function QuickAccessAssets() {
  const t = useTranslations("assets");

  return (
    <Card variant="elevated" className="border-0">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-base">{t("popularAssets")}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {/* Native XLM */}
          <Link
            href={XLM_ASSET.href}
            className="bg-primary/5 border-primary/20 hover:bg-primary/10 group flex items-center gap-3 rounded-xl border p-4 transition-all"
          >
            <div className="relative">
              <AssetLogo code="XLM" size="md" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <span className="font-semibold">{XLM_ASSET.name}</span>
                <Badge variant="secondary" className="text-[10px]">
                  Native
                </Badge>
              </div>
              <p className="text-muted-foreground text-xs">{XLM_ASSET.code}</p>
            </div>
            <ArrowRight className="text-muted-foreground group-hover:text-foreground size-4 transition-colors" />
          </Link>

          {/* Top 3 popular assets */}
          {Object.entries(POPULAR_ASSETS_META)
            .slice(0, 3)
            .map(([code, meta]) => {
              const popularAsset = POPULAR_ASSETS.find((a) => a.code === code);
              const issuer = popularAsset?.issuer ?? "";

              return (
                <Link
                  key={code}
                  href={`/asset/${code}-${issuer}`}
                  className="bg-card/50 hover:bg-card-hover group flex items-center gap-3 rounded-xl p-4 transition-all hover:-translate-y-0.5"
                >
                  <div className="relative">
                    <AssetLogo code={code} issuer={issuer} tomlUrl={meta.tomlUrl} size="md" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="font-medium">{meta.name}</p>
                    <p className="text-muted-foreground text-xs">{meta.description}</p>
                  </div>
                  <ArrowRight className="text-muted-foreground group-hover:text-foreground size-4 transition-colors" />
                </Link>
              );
            })}
        </div>
      </CardContent>
    </Card>
  );
}

export default function AssetsPage() {
  const t = useTranslations("assets");
  const tCommon = useTranslations("common");
  const { network } = useNetwork();

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("title")}
        subtitle={t("subtitle")}
        backHref="/"
        backLabel={tCommon("home")}
        showCopy={false}
        badge={<NetworkBadge network={network} />}
      />

      {/* Search Form */}
      <AssetSearchForm />

      {/* Quick Access */}
      <QuickAccessAssets />

      {/* Assets Tabs */}
      <Tabs defaultValue="top" className="w-full">
        <TabsList>
          <TabsTrigger value="top" className="gap-2">
            <TrendingUp className="size-4" />
            {t("topAssets")}
          </TabsTrigger>
          <TabsTrigger value="all" className="gap-2">
            <Coins className="size-4" />
            {t("allAssets")}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="top" className="mt-4">
          <TopAssetsSection />
        </TabsContent>

        <TabsContent value="all" className="mt-4">
          <AllAssetsList />
        </TabsContent>
      </Tabs>

      {/* Info Card */}
      <Card>
        <CardContent className="py-8">
          <div className="space-y-4 text-center">
            <div className="flex justify-center">
              <div className="bg-chart-3/10 flex size-16 items-center justify-center rounded-2xl">
                <Star className="text-chart-3 size-8" />
              </div>
            </div>
            <div>
              <h3 className="text-lg font-semibold">{t("stellarAssets")}</h3>
              <p className="text-muted-foreground mx-auto mt-2 max-w-md">
                {t("assetsDescription")}
              </p>
            </div>
            <p className="text-muted-foreground text-xs">{t("assetsHint")}</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
