"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { useAccountDataEntries } from "@/lib/hooks";
import { Database } from "lucide-react";
import { useTranslations } from "next-intl";

export function AccountDataEntries({ accountId }: { accountId: string }) {
  const { data, isLoading, error, refetch } = useAccountDataEntries(accountId);
  const t = useTranslations("account");

  if (isLoading) return <LoadingCard rows={3} />;
  if (error)
    return <ErrorState title={t("failedToLoadData")} message={error.message} onRetry={refetch} />;
  if (!data?.length)
    return <EmptyState title={t("noDataEntries")} description={t("noDataEntriesDesc")} />;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Database className="size-4" />
          {t("dataEntries", { count: data.length })}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {data.map(({ key, value, raw }: { key: string; value: string; raw: string }) => (
            <div key={key} className="bg-card/50 rounded-lg p-3">
              <div className="text-muted-foreground mb-1 text-xs font-medium tracking-wide uppercase">
                {key}
              </div>
              <div className="font-mono text-sm break-all">
                {value || <span className="text-muted-foreground italic">({raw})</span>}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
