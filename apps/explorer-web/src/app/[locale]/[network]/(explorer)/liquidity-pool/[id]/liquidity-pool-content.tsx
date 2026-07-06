"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { PageHeader } from "@/components/layout/page-header";
import { HashDisplay } from "@/components/common/hash-display";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { TransactionCard, TransactionCardSkeleton } from "@/components/cards/transaction-card";
import { Breadcrumbs } from "@/components/common/breadcrumbs";
import { useLiquidityPool, useLiquidityPoolTransactions } from "@/lib/hooks";
import { formatNumber, formatCompactNumber, truncateHash } from "@/lib/utils";
import { Droplets, ArrowLeftRight, Percent } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

interface LiquidityPoolContentProps {
  id: string;
}

function PoolSummary({ pool }: { pool: Horizon.ServerApi.LiquidityPoolRecord }) {
  const t = useTranslations("liquidityPool");
  const feePercent = ((pool.fee_bp ?? 0) / 100).toFixed(2);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{t("overview")}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          <div className="space-y-4">
            <div className="bg-primary/5 border-primary/10 rounded-lg border p-4">
              <div className="text-muted-foreground mb-2 flex items-center gap-2 text-sm">
                <Droplets className="size-4" />
                {t("reserves")}
              </div>
              <div className="space-y-2">
                {pool.reserves.map((r) => {
                  const isNative = r.asset === "native";
                  const [code, issuer] = isNative ? ["XLM", "native"] : r.asset.split(":");
                  return (
                    <div key={r.asset} className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary">{code}</Badge>
                        {!isNative && issuer && (
                          <HashDisplay
                            hash={issuer}
                            truncate
                            startLength={4}
                            endLength={4}
                            copyable={false}
                            linkTo={`/account/${issuer}`}
                            className="text-xs"
                          />
                        )}
                      </div>
                      <span className="font-mono text-sm tabular-nums">
                        {formatNumber(r.amount)}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="bg-card/50 rounded-lg p-4 text-center">
                <div className="flex items-center justify-center gap-1 text-2xl font-semibold">
                  <Percent className="size-4" />
                  {feePercent}
                </div>
                <div className="text-muted-foreground mt-1 text-xs">{t("fee")}</div>
              </div>
              <div className="bg-card/50 rounded-lg p-4 text-center">
                <div className="text-2xl font-semibold tabular-nums">
                  {formatCompactNumber(pool.total_trustlines ?? 0)}
                </div>
                <div className="text-muted-foreground mt-1 text-xs">{t("participants")}</div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("poolId")}</span>
              <HashDisplay hash={pool.id} truncate className="text-sm" />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("type")}</span>
              <Badge variant="outline">{pool.type}</Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("totalShares")}</span>
              <span className="font-mono text-sm tabular-nums">
                {formatNumber(pool.total_shares ?? "0")}
              </span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground flex items-center gap-1.5 text-sm">
                <ArrowLeftRight className="size-3.5" />
                {t("pair")}
              </span>
              <div className="flex items-center gap-1.5">
                {pool.reserves.map((r, i) => {
                  const code = r.asset === "native" ? "XLM" : r.asset.split(":")[0];
                  return (
                    <span key={r.asset}>
                      <Badge variant="secondary">{code}</Badge>
                      {i < pool.reserves.length - 1 && (
                        <span className="text-muted-foreground mx-1">/</span>
                      )}
                    </span>
                  );
                })}
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function PoolTransactions({ id }: { id: string }) {
  const { data, isLoading, error, refetch } = useLiquidityPoolTransactions(id);
  const t = useTranslations("liquidityPool");

  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 5 }).map((_, i) => (
          <TransactionCardSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (error) {
    return <ErrorState title={t("failedToLoadTx")} message={error.message} onRetry={refetch} />;
  }

  if (!data?.records?.length) {
    return <EmptyState title={t("noTransactions")} description={t("noTransactionsDesc")} />;
  }

  return (
    <div className="space-y-2">
      {data.records.map((tx: Horizon.ServerApi.TransactionRecord) => (
        <TransactionCard key={tx.hash} transaction={tx} />
      ))}
    </div>
  );
}

export function LiquidityPoolContent({ id }: LiquidityPoolContentProps) {
  const { data: pool, isLoading, error, refetch } = useLiquidityPool(id);
  const t = useTranslations("liquidityPool");

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="space-y-2">
          <div className="bg-muted h-8 w-48 animate-pulse rounded" />
          <div className="bg-muted h-4 w-64 animate-pulse rounded" />
        </div>
        <LoadingCard rows={6} />
      </div>
    );
  }

  if (error || !pool) {
    return (
      <div className="space-y-6">
        <PageHeader
          title={t("title")}
          subtitle={t("notFound")}
          backHref="/assets"
          backLabel={t("assets")}
        />
        <ErrorState
          title={t("poolNotFound")}
          message={t("poolNotFoundMessage")}
          onRetry={refetch}
        />
      </div>
    );
  }

  const pairLabel = pool.reserves
    .map((r) => (r.asset === "native" ? "XLM" : r.asset.split(":")[0]))
    .join(" / ");

  return (
    <div className="space-y-6">
      <Breadcrumbs
        items={[
          { label: t("assets"), href: "/assets" },
          { label: `${t("pool")} ${truncateHash(id, 6, 6)}`, href: `/liquidity-pool/${id}` },
        ]}
      />

      <PageHeader
        title={`${t("title")} · ${pairLabel}`}
        hash={id}
        backHref="/assets"
        backLabel={t("assets")}
      />

      <PoolSummary pool={pool} />

      <Tabs defaultValue="transactions" className="w-full">
        <TabsList>
          <TabsTrigger value="transactions">{t("transactions")}</TabsTrigger>
        </TabsList>
        <TabsContent value="transactions" className="mt-4">
          <PoolTransactions id={id} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
