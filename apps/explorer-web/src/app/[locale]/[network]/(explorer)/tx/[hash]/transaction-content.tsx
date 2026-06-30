"use client";

import { notFound } from "next/navigation";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageHeader } from "@/components/layout/page-header";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { CrossNetworkBanner } from "@/components/common/cross-network-banner";
import { DeveloperPanel } from "@/components/common/developer-panel";
import { StatusBadge } from "@/components/common/status-badge";
import { Breadcrumbs } from "@/components/common/breadcrumbs";
import { useTransaction } from "@/lib/hooks";
import { useNetwork } from "@/lib/providers";
import { NETWORKS } from "@/lib/constants";
import { truncateHash } from "@/lib/utils";
import { useTranslations } from "next-intl";
import { TransactionSummary } from "./sections/transaction-summary";
import { OperationsTimeline } from "./sections/operations-timeline";
import { TransactionEffects } from "./sections/transaction-effects";
import { RawData } from "./sections/raw-data";

interface TransactionContentProps {
  hash: string;
}

export function TransactionContent({ hash }: TransactionContentProps) {
  const { data: transaction, isLoading, error, refetch } = useTransaction(hash);
  const { network } = useNetwork();
  const networkConfig = NETWORKS[network];
  const t = useTranslations("transaction");
  const tNav = useTranslations("navigation");
  const tCommon = useTranslations("common");

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="space-y-2">
          <div className="bg-muted h-8 w-48 animate-pulse rounded" />
          <div className="bg-muted h-4 w-96 animate-pulse rounded" />
        </div>
        <LoadingCard rows={6} />
        <LoadingCard rows={5} />
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <PageHeader
          title={t("title")}
          subtitle={t("notFound")}
          backHref="/"
          backLabel={tCommon("home")}
        />
        <CrossNetworkBanner entityType="transaction" entityId={hash} enabled />
        <ErrorState title={t("notFound")} message={t("notFoundMessage")} onRetry={refetch} />
      </div>
    );
  }

  if (!transaction) {
    return notFound();
  }

  return (
    <div className="space-y-6">
      <Breadcrumbs
        items={[
          { label: tNav("transactions"), href: "/transactions" },
          { label: truncateHash(transaction.hash, 8, 8), href: `/tx/${transaction.hash}` },
        ]}
      />

      <PageHeader
        title={t("title")}
        hash={transaction.hash}
        backHref="/transactions"
        backLabel={tNav("transactions")}
        showQr={false}
        badge={<StatusBadge status={transaction.successful ? "success" : "failed"} />}
      />

      <TransactionSummary transaction={transaction} />

      <DeveloperPanel
        xdrData={{
          envelope: transaction.envelope_xdr,
          result: transaction.result_xdr,
        }}
        additionalEndpoints={[
          {
            label: "Transaction API",
            url: `${networkConfig.horizonUrl}/transactions/${transaction.hash}`,
          },
          {
            label: "Operations API",
            url: `${networkConfig.horizonUrl}/transactions/${transaction.hash}/operations`,
          },
        ]}
        internalIds={[
          { label: "Ledger Sequence", value: transaction.ledger_attr },
          { label: "Source Account Sequence", value: transaction.source_account_sequence },
          { label: "Paging Token", value: transaction.paging_token },
        ]}
      />

      <Tabs defaultValue="operations" className="w-full">
        <TabsList>
          <TabsTrigger value="operations">{t("operations")}</TabsTrigger>
          <TabsTrigger value="effects">{t("effects")}</TabsTrigger>
          <TabsTrigger value="raw">{t("rawData")}</TabsTrigger>
        </TabsList>
        <TabsContent value="operations" className="mt-4">
          <OperationsTimeline hash={hash} />
        </TabsContent>
        <TabsContent value="effects" className="mt-4">
          <TransactionEffects hash={hash} />
        </TabsContent>
        <TabsContent value="raw" className="mt-4">
          <RawData transaction={transaction} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
