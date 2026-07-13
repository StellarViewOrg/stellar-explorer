"use client";

import { Link } from "@/i18n/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { StatusBadge } from "@/components/common/status-badge";
import { HashDisplay } from "@/components/common/hash-display";
import { formatDateTime, formatLedgerSequence, stroopsToXLM } from "@/lib/utils";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function TransactionSummary({
  transaction,
}: {
  transaction: Horizon.ServerApi.TransactionRecord;
}) {
  const t = useTranslations("transaction");

  return (
    <Card variant="elevated" className="animate-fade-in-up border-0">
      <CardHeader>
        <CardTitle className="text-base">{t("details")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("status")}</span>
              <StatusBadge status={transaction.successful ? "success" : "failed"} />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("ledger")}</span>
              <Link
                href={`/ledger/${transaction.ledger_attr}`}
                className="text-primary text-sm font-medium hover:underline"
              >
                #{formatLedgerSequence(transaction.ledger_attr)}
              </Link>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("timestamp")}</span>
              <span className="text-sm">{formatDateTime(transaction.created_at)}</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("operationsCount")}</span>
              <span className="text-sm font-medium">{transaction.operation_count}</span>
            </div>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("feePaid")}</span>
              <span className="font-mono text-sm">{stroopsToXLM(transaction.fee_charged)} XLM</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("maxFee")}</span>
              <span className="font-mono text-sm">{stroopsToXLM(transaction.max_fee)} XLM</span>
            </div>
            <Separator />
            <div className="flex items-start justify-between">
              <span className="text-muted-foreground text-sm">{t("source")}</span>
              <HashDisplay
                hash={transaction.source_account}
                truncate
                linkTo={`/account/${transaction.source_account}`}
                className="text-sm"
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("memo")}</span>
              <span className="font-mono text-sm">{transaction.memo || "-"}</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
