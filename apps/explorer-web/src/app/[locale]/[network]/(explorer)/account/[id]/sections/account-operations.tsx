"use client";

import { Link } from "@/i18n/navigation";
import { Card, CardContent } from "@/components/ui/card";
import { OperationBadge } from "@/components/common/operation-badge";
import { TimeAgo } from "@/components/common/time-ago";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { OperationDetails } from "@/components/operations/operation-details";
import { OperationSummary } from "@/components/operations/operation-summary";
import { useAccountOperations, useAccountOperationsStream } from "@/lib/hooks";
import { truncateHash } from "@/lib/utils";
import { asHorizonOperation } from "@/lib/utils/horizon-types";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function AccountOperations({ accountId }: { accountId: string }) {
  const { data, isLoading, error, refetch } = useAccountOperations(accountId);
  const t = useTranslations("account");

  useAccountOperationsStream(accountId, { enabled: true });

  if (isLoading) {
    return <LoadingCard rows={5} />;
  }

  if (error) {
    return <ErrorState title={t("failedToLoadOps")} message={error.message} onRetry={refetch} />;
  }

  if (!data?.records?.length) {
    return <EmptyState title={t("noOperations")} description={t("noOperationsDesc")} />;
  }

  return (
    <Card>
      <CardContent className="p-0">
        <div className="divide-border divide-y">
          {data.records.map((op: Horizon.ServerApi.OperationRecord) => {
            const opRecord = asHorizonOperation(op);
            return (
              <div key={op.id} className="hover:bg-card-hover p-4 transition-colors">
                <div className="mb-2 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <OperationBadge type={op.type} />
                    <Link
                      href={`/tx/${op.transaction_hash}`}
                      className="hover:text-primary font-mono text-sm transition-colors"
                    >
                      {truncateHash(op.transaction_hash, 8, 4)}
                    </Link>
                  </div>
                  <TimeAgo timestamp={op.created_at} className="text-xs" />
                </div>
                <OperationSummary operation={opRecord} />
                <OperationDetails operation={opRecord} />
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
