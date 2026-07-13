"use client";

import { TransactionCard, TransactionCardSkeleton } from "@/components/cards/transaction-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { useAccountTransactions } from "@/lib/hooks";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function AccountTransactions({ accountId }: { accountId: string }) {
  const { data, isLoading, error, refetch } = useAccountTransactions(accountId);
  const t = useTranslations("account");

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
