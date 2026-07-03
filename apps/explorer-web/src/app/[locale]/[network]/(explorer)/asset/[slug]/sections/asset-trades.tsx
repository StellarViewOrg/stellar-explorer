"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { useAssetTrades } from "@/lib/hooks";
import { formatCompactNumber } from "@/lib/utils";

export function AssetTrades({ code, issuer }: { code: string; issuer: string }) {
  const { data, isLoading, error } = useAssetTrades(code, issuer);

  if (isLoading) return <LoadingCard rows={4} />;
  if (error) return <ErrorState title="Failed to load trade data" message={error.message} />;
  if (!data?.records?.length)
    return (
      <EmptyState
        title="No trade data"
        description="No trades found for this asset against XLM in the past 24h."
      />
    );

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-semibold tabular-nums">
              {formatCompactNumber(data.volume24h)}
            </div>
            <div className="text-muted-foreground mt-1 text-xs">24h Volume (XLM)</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-semibold tabular-nums">{data.tradeCount}</div>
            <div className="text-muted-foreground mt-1 text-xs">Trades</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 text-center">
            <div
              className={`text-2xl font-semibold tabular-nums ${
                data.priceChange24h >= 0 ? "text-success" : "text-destructive"
              }`}
            >
              {data.priceChange24h >= 0 ? "+" : ""}
              {data.priceChange24h.toFixed(2)}%
            </div>
            <div className="text-muted-foreground mt-1 text-xs">24h Change</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-semibold tabular-nums">
              {data.close24h?.toFixed(7) ?? "—"}
            </div>
            <div className="text-muted-foreground mt-1 text-xs">Price (XLM)</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Hourly Volume (24h)</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-1">
            {data.records.map((record) => {
              const vol = parseFloat(record.base_volume);
              const maxVol = Math.max(...data.records.map((r) => parseFloat(r.base_volume)));
              const pct = maxVol > 0 ? (vol / maxVol) * 100 : 0;
              const hour = new Date(Number(record.timestamp)).toLocaleTimeString([], {
                hour: "2-digit",
                minute: "2-digit",
              });
              return (
                <div key={record.timestamp} className="flex items-center gap-2 text-xs">
                  <span className="text-muted-foreground w-12 shrink-0">{hour}</span>
                  <div className="bg-muted h-4 flex-1 rounded-sm">
                    <div
                      className="bg-primary h-full rounded-sm transition-all"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <span className="text-muted-foreground w-20 text-right tabular-nums">
                    {formatCompactNumber(vol)}
                  </span>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
