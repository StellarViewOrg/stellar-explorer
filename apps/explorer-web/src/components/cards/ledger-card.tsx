"use client";

import { memo } from "react";
import { Link } from "@/i18n/navigation";
import { Card, CardContent } from "@/components/ui/card";
import { TimeAgo } from "@/components/common/time-ago";
import { formatLedgerSequence } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { Layers } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

interface LedgerCardProps {
  ledger: Horizon.ServerApi.LedgerRecord;
  className?: string;
}

export const LedgerCard = memo(function LedgerCard({ ledger, className }: LedgerCardProps) {
  const t = useTranslations("cards.ledger");

  return (
    <Link href={`/ledger/${ledger.sequence}`}>
      <Card variant="elevated" interactive className={cn("group border-0 py-0", className)}>
        <CardContent className="flex items-center justify-between p-4">
          <div className="flex items-center gap-3">
            <div className="bg-chart-1/10 relative flex size-10 items-center justify-center rounded-xl">
              <div className="bg-chart-1/20 absolute inset-0 rounded-xl opacity-0 blur-md transition-opacity duration-300 group-hover:opacity-60" />
              <Layers className="text-chart-1 relative size-5" />
            </div>
            <div>
              <span className="text-sm font-medium tabular-nums">
                #{formatLedgerSequence(ledger.sequence)}
              </span>
              <div className="text-muted-foreground mt-0.5 flex items-center gap-2 text-xs">
                <span>
                  {ledger.successful_transaction_count} {t("txs")}
                </span>
                <span className="opacity-30">|</span>
                <span>
                  {ledger.operation_count} {t("ops")}
                </span>
              </div>
            </div>
          </div>
          <TimeAgo timestamp={ledger.closed_at} className="shrink-0 text-xs" />
        </CardContent>
      </Card>
    </Link>
  );
});

export function LedgerCardSkeleton({ className }: { className?: string }) {
  return (
    <Card
      variant="elevated"
      className={cn("border-0 py-0", className)}
      role="status"
      aria-busy="true"
    >
      <CardContent className="flex items-center justify-between p-4">
        <div className="flex items-center gap-3">
          <div className="bg-muted/50 size-10 animate-pulse rounded-xl" />
          <div className="space-y-2">
            <div className="bg-muted/50 h-4 w-24 animate-pulse rounded" />
            <div className="bg-muted/50 h-3 w-20 animate-pulse rounded" />
          </div>
        </div>
        <div className="bg-muted/50 h-3 w-12 animate-pulse rounded" />
      </CardContent>
    </Card>
  );
}
