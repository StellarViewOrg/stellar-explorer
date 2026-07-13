"use client";

import { notFound } from "next/navigation";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/layout/page-header";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { CrossNetworkBanner } from "@/components/common/cross-network-banner";
import { Breadcrumbs } from "@/components/common/breadcrumbs";
import { MergedAccountView, type AccountMergeOp } from "@/components/accounts/merged-account-view";
import { AccountStatement } from "@/components/accounts/account-statement";
import { CopyContextButton } from "@/components/common/copy-context-button";
import { useAccount, useAccountLastOperations, useWatchlist } from "@/lib/hooks";
import { truncateHash } from "@/lib/utils";
import { asRecord, isAccountMergeOp } from "@/lib/utils/horizon-types";
import { Star, StarOff, ArrowLeftRight, Database } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";
import { AccountSummary } from "./sections/account-summary";
import { AccountBalances } from "./sections/account-balances";
import { AccountTransactions } from "./sections/account-transactions";
import { AccountOperations } from "./sections/account-operations";
import { AccountSigners } from "./sections/account-signers";
import { AccountOffers } from "./sections/account-offers";
import { AccountDataEntries } from "./sections/account-data-entries";

interface AccountContentProps {
  id: string;
}

export function AccountContent({ id }: AccountContentProps) {
  const { data: account, isLoading, error, refetch } = useAccount(id);
  const { data: lastOps } = useAccountLastOperations(id, true);
  const { has, add, remove } = useWatchlist();
  const isWatched = has(id);
  const t = useTranslations("account");
  const tNav = useTranslations("navigation");
  const tCommon = useTranslations("common");

  const toggleWatchlist = () => {
    if (isWatched) {
      remove(id);
    } else {
      add({ type: "account", id });
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="space-y-2">
          <div className="bg-muted h-8 w-32 animate-pulse rounded" />
          <div className="bg-muted h-4 w-96 animate-pulse rounded" />
        </div>
        <LoadingCard rows={6} />
      </div>
    );
  }

  if (error) {
    const mergeOp = lastOps?.records?.find(
      (op): op is Horizon.ServerApi.BaseOperationRecord & AccountMergeOp =>
        isAccountMergeOp(op) && op.account === id
    );

    if (mergeOp) {
      return <MergedAccountView id={id} mergeOp={mergeOp} />;
    }

    return (
      <div className="space-y-6">
        <PageHeader
          title={t("title")}
          subtitle={t("notFound")}
          backHref="/"
          backLabel={tCommon("home")}
        />
        <CrossNetworkBanner entityType="account" entityId={id} enabled />
        <ErrorState title={t("notFound")} message={t("notFoundMessage")} onRetry={refetch} />
      </div>
    );
  }

  if (!account) {
    return notFound();
  }

  return (
    <div className="space-y-6">
      <Breadcrumbs
        items={[
          { label: tNav("accounts"), href: "/accounts" },
          { label: truncateHash(id, 6, 6), href: `/account/${id}` },
        ]}
      />

      <PageHeader
        title={t("title")}
        hash={id}
        backHref="/accounts"
        backLabel={tNav("accounts")}
        showQr
        actions={
          <div className="flex items-center gap-2">
            <CopyContextButton type="account" data={asRecord(account)} />
            <AccountStatement accountId={id} />
            <Button
              variant={isWatched ? "secondary" : "outline"}
              size="sm"
              onClick={toggleWatchlist}
              className="hover:bg-white/10"
            >
              {isWatched ? (
                <>
                  <StarOff className="mr-2 size-4" />
                  {tCommon("remove")}
                </>
              ) : (
                <>
                  <Star className="mr-2 size-4" />
                  {tCommon("watch")}
                </>
              )}
            </Button>
          </div>
        }
      />

      <AccountSummary account={account} id={id} />

      <Tabs defaultValue="activity" className="w-full">
        <TabsList>
          <TabsTrigger value="activity">{t("activity")}</TabsTrigger>
          <TabsTrigger value="operations">{t("operations")}</TabsTrigger>
          <TabsTrigger value="balances">{t("balances")}</TabsTrigger>
          <TabsTrigger value="offers">
            <ArrowLeftRight className="mr-1.5 size-3.5" />
            {t("offersTab")}
          </TabsTrigger>
          <TabsTrigger value="data">
            <Database className="mr-1.5 size-3.5" />
            {t("dataTab")}
          </TabsTrigger>
          <TabsTrigger value="signers">{t("signers")}</TabsTrigger>
        </TabsList>
        <TabsContent value="activity" className="mt-4">
          <AccountTransactions accountId={id} />
        </TabsContent>
        <TabsContent value="operations" className="mt-4">
          <AccountOperations accountId={id} />
        </TabsContent>
        <TabsContent value="balances" className="mt-4">
          <AccountBalances account={account} />
        </TabsContent>
        <TabsContent value="offers" className="mt-4">
          <AccountOffers accountId={id} />
        </TabsContent>
        <TabsContent value="data" className="mt-4">
          <AccountDataEntries accountId={id} />
        </TabsContent>
        <TabsContent value="signers" className="mt-4">
          <AccountSigners account={account} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
