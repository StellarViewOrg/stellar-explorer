"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { useAccountOffers } from "@/lib/hooks";
import { formatNumber } from "@/lib/utils";
import { ArrowLeftRight } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function AccountOffers({ accountId }: { accountId: string }) {
  const { data, isLoading, error, refetch } = useAccountOffers(accountId);
  const t = useTranslations("account");

  if (isLoading) return <LoadingCard rows={4} />;
  if (error)
    return <ErrorState title={t("failedToLoadOffers")} message={error.message} onRetry={refetch} />;
  if (!data?.records?.length)
    return <EmptyState title={t("noOpenOffers")} description={t("noOpenOffersDesc")} />;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <ArrowLeftRight className="size-4" />
          {t("openOffers", { count: data.records.length })}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {data.records.map((offer: Horizon.ServerApi.OfferRecord) => {
            const selling =
              offer.selling.asset_type === "native"
                ? "XLM"
                : `${offer.selling.asset_code}/${offer.selling.asset_issuer?.slice(0, 4)}…`;
            const buying =
              offer.buying.asset_type === "native"
                ? "XLM"
                : `${offer.buying.asset_code}/${offer.buying.asset_issuer?.slice(0, 4)}…`;

            return (
              <div
                key={offer.id}
                className="bg-card/50 flex flex-col gap-1.5 rounded-lg p-3 sm:flex-row sm:items-center sm:justify-between"
              >
                <div className="flex items-center gap-2 text-sm">
                  <Badge variant="outline">{t("sellAsset", { asset: selling })}</Badge>
                  <span className="text-muted-foreground">→</span>
                  <Badge variant="outline">{t("buyAsset", { asset: buying })}</Badge>
                </div>
                <div className="text-muted-foreground text-xs">
                  {t("amount")}:{" "}
                  <span className="text-foreground font-medium">{formatNumber(offer.amount)}</span>
                  {" · "}
                  {t("price")}: <span className="text-foreground font-medium">{offer.price}</span>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
