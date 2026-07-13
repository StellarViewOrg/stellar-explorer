"use client";

import { notFound } from "next/navigation";
import { useTranslations } from "next-intl";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageHeader } from "@/components/layout/page-header";
import { CrossNetworkBanner } from "@/components/common/cross-network-banner";
import { CopyContextButton } from "@/components/common/copy-context-button";
import { ContractEventDetails, ContractTransactions, TokenGallery } from "@/components/contracts";
import { useContractInfo } from "@/lib/hooks";
import { isValidContractId } from "@/lib/utils";
import { Activity, Database, Code, ArrowRightLeft, Zap, LayoutGrid } from "lucide-react";
import { ContractSummary } from "./sections/contract-summary";
import { ContractInvocations } from "./sections/contract-invocations";
import { ContractStorage } from "./sections/contract-storage";
import { ContractCode } from "./sections/contract-code";

interface ContractContentProps {
  id: string;
}

export function ContractContent({ id }: ContractContentProps) {
  const t = useTranslations("contract");
  const tCommon = useTranslations("common");
  const { error: contractError } = useContractInfo(id);

  if (!isValidContractId(id)) {
    return notFound();
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("title")}
        hash={id}
        backHref="/"
        backLabel={tCommon("home")}
        showQr={false}
        actions={<CopyContextButton type="contract" data={{ id }} />}
      />

      <CrossNetworkBanner entityType="contract" entityId={id} enabled={!!contractError} />

      <ContractSummary contractId={id} />

      <Tabs defaultValue="events" className="w-full">
        <TabsList className="flex-wrap">
          <TabsTrigger value="events">
            <Activity className="mr-2 size-4" />
            {t("events")}
          </TabsTrigger>
          <TabsTrigger value="invocations">
            <Zap className="mr-2 size-4" />
            Invocations
          </TabsTrigger>
          <TabsTrigger value="transactions">
            <ArrowRightLeft className="mr-2 size-4" />
            {t("contractTransactions")}
          </TabsTrigger>
          <TabsTrigger value="storage">
            <Database className="mr-2 size-4" />
            {t("storage")}
          </TabsTrigger>
          <TabsTrigger value="gallery">
            <LayoutGrid className="mr-2 size-4" />
            Gallery
          </TabsTrigger>
          <TabsTrigger value="code">
            <Code className="mr-2 size-4" />
            {t("code")}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="events" className="mt-4">
          <ContractEventDetails contractId={id} live />
        </TabsContent>
        <TabsContent value="invocations" className="mt-4">
          <ContractInvocations contractId={id} />
        </TabsContent>
        <TabsContent value="transactions" className="mt-4">
          <ContractTransactions contractId={id} />
        </TabsContent>
        <TabsContent value="storage" className="mt-4">
          <ContractStorage contractId={id} />
        </TabsContent>
        <TabsContent value="gallery" className="mt-4">
          <TokenGallery contractId={id} />
        </TabsContent>
        <TabsContent value="code" className="mt-4">
          <ContractCode contractId={id} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
