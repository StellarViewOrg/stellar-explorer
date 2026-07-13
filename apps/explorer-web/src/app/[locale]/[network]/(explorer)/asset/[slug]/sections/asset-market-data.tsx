"use client";

import { Card, CardContent } from "@/components/ui/card";
import { LoadingCard } from "@/components/common/loading-card";
import { useAssetAnalytics } from "@/lib/hooks";
import { formatCompactNumber } from "@/lib/utils";
import { Star, ExternalLink } from "lucide-react";

export function AssetMarketData({ code, issuer }: { code: string; issuer: string }) {
  const { data, isLoading } = useAssetAnalytics(code, issuer);

  if (isLoading) return <LoadingCard rows={4} />;

  if (!data || !data.available) {
    return (
      <Card>
        <CardContent className="p-6 text-center">
          <Star className="text-muted-foreground mx-auto mb-3 size-8" />
          <p className="text-muted-foreground text-sm">
            Historical market analytics are only available on the Public network via{" "}
            <a
              href={`https://stellar.expert/explorer/public/asset/${code}-${issuer}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary hover:underline"
            >
              Stellar Expert
              <ExternalLink className="ml-1 inline size-3" />
            </a>
          </p>
        </CardContent>
      </Card>
    );
  }

  const d = data.data;

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        {d.payments_amount !== undefined && (
          <Card>
            <CardContent className="p-4 text-center">
              <div className="text-2xl font-semibold tabular-nums">
                {formatCompactNumber(d.payments_amount)}
              </div>
              <div className="text-muted-foreground mt-1 text-xs">Total Payment Volume</div>
            </CardContent>
          </Card>
        )}
        {d.payments !== undefined && (
          <Card>
            <CardContent className="p-4 text-center">
              <div className="text-2xl font-semibold tabular-nums">
                {formatCompactNumber(d.payments)}
              </div>
              <div className="text-muted-foreground mt-1 text-xs">Total Payments</div>
            </CardContent>
          </Card>
        )}
        {d.trades !== undefined && (
          <Card>
            <CardContent className="p-4 text-center">
              <div className="text-2xl font-semibold tabular-nums">
                {formatCompactNumber(d.trades)}
              </div>
              <div className="text-muted-foreground mt-1 text-xs">Total Trades</div>
            </CardContent>
          </Card>
        )}
        {d.rating && (
          <Card>
            <CardContent className="p-4 text-center">
              <div className="text-2xl font-semibold tabular-nums">
                {d.rating.average?.toFixed(1) ?? "—"}
              </div>
              <div className="text-muted-foreground mt-1 text-xs">Trust Rating</div>
            </CardContent>
          </Card>
        )}
      </div>

      {d.created && (
        <Card>
          <CardContent className="flex items-center justify-between p-4">
            <span className="text-muted-foreground text-sm">Asset First Seen</span>
            <span className="text-sm font-medium">
              {new Date(d.created * 1000).toLocaleDateString()}
            </span>
          </CardContent>
        </Card>
      )}

      <div className="text-right">
        <a
          href={`https://stellar.expert/explorer/public/asset/${code}-${issuer}`}
          target="_blank"
          rel="noopener noreferrer"
          className="text-muted-foreground hover:text-primary inline-flex items-center gap-1 text-xs transition-colors"
        >
          View full analytics on Stellar Expert
          <ExternalLink className="size-3" />
        </a>
      </div>
    </div>
  );
}
