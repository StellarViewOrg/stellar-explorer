"use client";

import { memo, useMemo } from "react";
import { Link } from "@/i18n/navigation";
import { Card, CardContent } from "@/components/ui/card";
import { HashDisplay } from "@/components/common/hash-display";
import { TimeAgo } from "@/components/common/time-ago";
import { cn } from "@/lib/utils";
import { CheckCircle2, XCircle, ChevronRight, ArrowRight } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";
import { decodeTransactionEnvelope } from "@/lib/utils/xdr-decoder";
import { useNetwork } from "@/lib/providers";

interface TransactionCardProps {
  transaction: Horizon.ServerApi.TransactionRecord;
  className?: string;
  animationDelay?: number;
}

export const TransactionCard = memo(function TransactionCard({
  transaction,
  className,
  animationDelay,
}: TransactionCardProps) {
  const isSuccess = transaction.successful;
  const t = useTranslations("cards.transaction");
  const tOps = useTranslations("operations");
  const { network } = useNetwork();

  const decoded = useMemo(() => {
    return decodeTransactionEnvelope(transaction.envelope_xdr, network);
  }, [transaction, network]);

  const operationLabel = useMemo(() => {
    if (!decoded) return t("unknownOp");
    const key = decoded.operationType as Parameters<typeof tOps>[0];
    try {
      return tOps(key);
    } catch {
      return t("unknownOp");
    }
  }, [decoded, t, tOps]);

  return (
    <Link href={`/tx/${transaction.hash}`}>
      <Card
        variant="elevated"
        interactive
        className={cn("group animate-fade-in-up border-0 py-0", className)}
        style={animationDelay ? { animationDelay: `${animationDelay}ms` } : undefined}
      >
        <CardContent className="flex items-center justify-between p-4">
          <div className="flex min-w-0 items-center gap-4">
            {/* Status icon with glow */}
            <div
              className={cn(
                "relative flex size-10 shrink-0 items-center justify-center rounded-xl transition-all duration-300",
                isSuccess ? "bg-success/10 text-success" : "bg-destructive/10 text-destructive"
              )}
            >
              {/* Glow effect */}
              <div
                className={cn(
                  "absolute inset-0 rounded-xl opacity-0 blur-md transition-opacity duration-300 group-hover:opacity-60",
                  isSuccess ? "bg-success/30" : "bg-destructive/30"
                )}
              />
              {isSuccess ? (
                <CheckCircle2 className="relative size-5" />
              ) : (
                <XCircle className="relative size-5" />
              )}
            </div>

            {/* Transaction info */}
            <div className="min-w-0 flex-1 space-y-1">
              {/* Row 1: Operation type + amount/asset */}
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm font-medium">{operationLabel}</span>
                {decoded?.amount && decoded?.asset && (
                  <span className="text-muted-foreground shrink-0 text-sm font-medium">
                    {decoded.amount} {decoded.asset}
                  </span>
                )}
              </div>

              {/* Row 2: Source → Destination */}
              <div className="text-muted-foreground flex items-center gap-1 text-xs">
                <HashDisplay
                  hash={decoded?.sourceAccount ?? transaction.source_account}
                  truncate
                  startLength={4}
                  endLength={4}
                  copyable={false}
                  className="text-xs"
                />
                {decoded?.destination && (
                  <>
                    <ArrowRight className="size-3 shrink-0" />
                    <HashDisplay
                      hash={decoded.destination}
                      truncate
                      startLength={4}
                      endLength={4}
                      copyable={false}
                      className="text-xs"
                    />
                  </>
                )}
              </div>

              {/* Row 3: Hash, op count, timestamp, memo */}
              <div className="text-muted-foreground flex flex-wrap items-center gap-2 text-xs">
                <HashDisplay
                  hash={transaction.hash}
                  truncate
                  startLength={6}
                  endLength={4}
                  copyable={false}
                  className="font-mono text-xs"
                />
                <span className="opacity-30">|</span>
                <span>
                  {transaction.operation_count}{" "}
                  {transaction.operation_count !== 1 ? t("operations") : t("operation")}
                </span>
                <span className="opacity-30">|</span>
                <TimeAgo timestamp={transaction.created_at} />
                {decoded?.memo && decoded.memoType !== "none" && (
                  <>
                    <span className="opacity-30">|</span>
                    <span>
                      {t("memo")}: &ldquo;{decoded.memo}&rdquo;
                    </span>
                  </>
                )}
              </div>
            </div>
          </div>

          {/* Arrow indicator */}
          <ChevronRight className="text-muted-foreground ml-2 size-5 shrink-0 -translate-x-2 opacity-0 transition-all duration-300 group-hover:translate-x-0 group-hover:opacity-100" />
        </CardContent>
      </Card>
    </Link>
  );
});

interface TransactionCardSkeletonProps {
  className?: string;
}

export function TransactionCardSkeleton({ className }: TransactionCardSkeletonProps) {
  return (
    <Card
      variant="elevated"
      className={cn("border-0 py-0", className)}
      role="status"
      aria-busy="true"
    >
      <CardContent className="flex items-center justify-between p-4">
        <div className="flex items-center gap-4">
          <div className="bg-muted/50 size-10 animate-pulse rounded-xl" />
          <div className="space-y-2">
            <div className="bg-muted/50 h-4 w-32 animate-pulse rounded" />
            <div className="bg-muted/50 h-3 w-24 animate-pulse rounded" />
          </div>
        </div>
        <div className="bg-muted/50 h-4 w-4 animate-pulse rounded" />
      </CardContent>
    </Card>
  );
}
