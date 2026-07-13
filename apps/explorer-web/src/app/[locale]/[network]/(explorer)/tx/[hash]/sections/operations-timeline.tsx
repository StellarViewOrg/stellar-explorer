"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { OperationBadge } from "@/components/common/operation-badge";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { OperationDetails } from "@/components/operations/operation-details";
import { OperationSummary } from "@/components/operations/operation-summary";
import { useTransactionOperations } from "@/lib/hooks";
import { asHorizonOperation } from "@/lib/utils/horizon-types";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function OperationsTimeline({ hash }: { hash: string }) {
  const { data, isLoading, error, refetch } = useTransactionOperations(hash);
  const t = useTranslations("transaction");

  if (isLoading) {
    return <LoadingCard rows={5} />;
  }

  if (error) {
    return <ErrorState title={t("failedToLoadOps")} message={error.message} onRetry={refetch} />;
  }

  if (!data?.records?.length) {
    return (
      <Card>
        <CardContent className="text-muted-foreground py-8 text-center">
          {t("noOperations")}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">
          {t("operations")} ({data.records.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {data.records.map((op: Horizon.ServerApi.OperationRecord, index: number) => (
            <div key={op.id} className="relative">
              {index < data.records.length - 1 && (
                <div className="bg-border absolute top-8 left-3 h-full w-0.5" />
              )}

              <div className="flex gap-4">
                <div className="bg-primary text-primary-foreground relative z-10 mt-1 flex size-6 shrink-0 items-center justify-center rounded-full text-xs font-medium">
                  {index + 1}
                </div>

                <div className="min-w-0 flex-1 pb-4">
                  <div className="mb-2 flex items-center gap-2">
                    <OperationBadge type={op.type} />
                  </div>

                  <OperationSummary operation={asHorizonOperation(op)} />
                  <OperationDetails operation={asHorizonOperation(op)} />
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
