"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { HashDisplay } from "@/components/common/hash-display";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { useTransactionEffects } from "@/lib/hooks";
import { formatNumber } from "@/lib/utils";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

function getEffectColorClass(effectType: string): string {
  if (
    effectType.includes("created") ||
    effectType.includes("credited") ||
    effectType.includes("claimed")
  ) {
    return "bg-success/5";
  }
  if (
    effectType.includes("removed") ||
    effectType.includes("debited") ||
    effectType.includes("clawed")
  ) {
    return "bg-destructive/5";
  }
  if (
    effectType.includes("updated") ||
    effectType.includes("trade") ||
    effectType.includes("changed")
  ) {
    return "bg-primary/5";
  }
  return "bg-card/50";
}

export function TransactionEffects({ hash }: { hash: string }) {
  const { data, isLoading, error, refetch } = useTransactionEffects(hash);
  const t = useTranslations("transaction");

  if (isLoading) {
    return <LoadingCard rows={5} />;
  }

  if (error) {
    return (
      <ErrorState title={t("failedToLoadEffects")} message={error.message} onRetry={refetch} />
    );
  }

  if (!data?.records?.length) {
    return (
      <Card>
        <CardContent className="text-muted-foreground py-8 text-center">
          {t("noEffects")}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">
          {t("effects")} ({data.records.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {data.records.map((effect: Horizon.ServerApi.EffectRecord) => {
            const eff = effect as unknown as Record<string, unknown>;
            const effectType = eff.type as string;
            const colorClass = getEffectColorClass(effectType);

            return (
              <div
                key={effect.id}
                className={`flex items-center justify-between rounded-lg p-3 ${colorClass}`}
              >
                <div className="flex items-center gap-3">
                  <Badge variant="outline" className="capitalize">
                    {effectType.replace(/_/g, " ")}
                  </Badge>
                  {typeof eff.account === "string" && (
                    <HashDisplay
                      hash={eff.account}
                      truncate
                      startLength={6}
                      endLength={4}
                      linkTo={`/account/${eff.account}`}
                      className="text-sm"
                    />
                  )}
                  {typeof eff.sold_asset_code === "string" && (
                    <span className="text-muted-foreground text-xs">
                      {formatNumber(eff.sold_amount as string)}{" "}
                      {eff.sold_asset_type === "native" ? "XLM" : eff.sold_asset_code}
                      {" → "}
                      {formatNumber(eff.bought_amount as string)}{" "}
                      {eff.bought_asset_type === "native"
                        ? "XLM"
                        : (eff.bought_asset_code as string)}
                    </span>
                  )}
                  {typeof eff.weight === "number" && (
                    <span className="text-muted-foreground text-xs">weight: {eff.weight}</span>
                  )}
                </div>
                {typeof eff.amount === "string" && (
                  <span className="font-mono text-sm">
                    {formatNumber(eff.amount)}{" "}
                    {(eff.asset_type as string) === "native"
                      ? "XLM"
                      : (eff.asset_code as string) || ""}
                  </span>
                )}
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
