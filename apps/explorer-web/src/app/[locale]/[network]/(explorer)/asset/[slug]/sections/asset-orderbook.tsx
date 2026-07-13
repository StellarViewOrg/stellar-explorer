"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { useAssetOrderbook } from "@/lib/hooks";
import { formatCompactNumber } from "@/lib/utils";

export function AssetOrderbook({ code, issuer }: { code: string; issuer: string }) {
  const { data, isLoading, error } = useAssetOrderbook(code, issuer);

  if (isLoading) return <LoadingCard rows={5} />;
  if (error) return <ErrorState title="Failed to load orderbook" message={error.message} />;
  if (!data?.bids?.length && !data?.asks?.length)
    return <EmptyState title="Empty orderbook" description="No open orders for this asset." />;

  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
      {/* Summary */}
      {data.midPrice !== null && (
        <div className="col-span-full grid grid-cols-3 gap-4">
          <Card>
            <CardContent className="p-4 text-center">
              <div className="text-xl font-semibold tabular-nums">
                {data.bestBid?.toFixed(7) ?? "—"}
              </div>
              <div className="text-muted-foreground mt-1 text-xs">Best Bid</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <div className="text-xl font-semibold tabular-nums">
                {data.midPrice?.toFixed(7) ?? "—"}
              </div>
              <div className="text-muted-foreground mt-1 text-xs">Mid Price</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4 text-center">
              <div className="text-xl font-semibold tabular-nums">
                {data.bestAsk?.toFixed(7) ?? "—"}
              </div>
              <div className="text-muted-foreground mt-1 text-xs">Best Ask</div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Bids */}
      <Card>
        <CardHeader>
          <CardTitle className="text-success flex items-center gap-2 text-base">Bids</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-1">
            <div className="text-muted-foreground grid grid-cols-2 text-xs">
              <span>Price (XLM)</span>
              <span className="text-right">Amount</span>
            </div>
            {data.bids.map((bid, i) => (
              <div key={i} className="text-success/90 grid grid-cols-2 text-sm tabular-nums">
                <span>{parseFloat(bid.price).toFixed(7)}</span>
                <span className="text-right">{formatCompactNumber(parseFloat(bid.amount))}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Asks */}
      <Card>
        <CardHeader>
          <CardTitle className="text-destructive flex items-center gap-2 text-base">Asks</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-1">
            <div className="text-muted-foreground grid grid-cols-2 text-xs">
              <span>Price (XLM)</span>
              <span className="text-right">Amount</span>
            </div>
            {data.asks.map((ask, i) => (
              <div key={i} className="text-destructive/90 grid grid-cols-2 text-sm tabular-nums">
                <span>{parseFloat(ask.price).toFixed(7)}</span>
                <span className="text-right">{formatCompactNumber(parseFloat(ask.amount))}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {data.spread !== null && (
        <div className="col-span-full">
          <p className="text-muted-foreground text-center text-sm">
            Spread: <span className="text-foreground font-medium">{data.spread.toFixed(4)}%</span>
          </p>
        </div>
      )}
    </div>
  );
}
