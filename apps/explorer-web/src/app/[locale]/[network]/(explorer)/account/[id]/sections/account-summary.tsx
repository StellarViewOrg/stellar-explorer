"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { useAccountProfile } from "@/lib/hooks";
import { formatNumber } from "@/lib/utils";
import { Calendar } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function AccountSummary({
  account,
  id,
}: {
  account: Horizon.ServerApi.AccountRecord;
  id: string;
}) {
  const t = useTranslations("account");
  const { data: profile } = useAccountProfile(id);

  const xlmBalance = account.balances.find((b) => b.asset_type === "native") as
    | Horizon.HorizonApi.BalanceLineNative
    | undefined;

  const otherAssets = account.balances.filter((b) => b.asset_type !== "native");

  const createdAt =
    profile?.available && profile.data.created
      ? new Date(profile.data.created * 1000).toLocaleDateString()
      : null;

  return (
    <Card variant="elevated" className="animate-fade-in-up border-0">
      <CardHeader>
        <CardTitle className="text-base">{t("overview")}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          <div className="space-y-4">
            <div className="bg-primary/5 border-primary/10 rounded-xl border p-4">
              <span className="text-muted-foreground text-sm">{t("xlmBalance")}</span>
              <div className="mt-1 text-2xl font-semibold tabular-nums">
                {xlmBalance ? formatNumber(xlmBalance.balance) : "0"} XLM
              </div>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("otherAssets")}</span>
              <Badge variant="secondary">{otherAssets.length}</Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("subentries")}</span>
              <span className="text-sm font-medium">{account.subentry_count}</span>
            </div>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("sequence")}</span>
              <span className="font-mono text-sm">{account.sequence}</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("thresholds")}</span>
              <div className="space-x-2 text-sm">
                <span>L:{account.thresholds.low_threshold}</span>
                <span>M:{account.thresholds.med_threshold}</span>
                <span>H:{account.thresholds.high_threshold}</span>
              </div>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("signers")}</span>
              <Badge variant="secondary">{account.signers.length}</Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("homeDomain")}</span>
              <span className="text-sm">{account.home_domain || "-"}</span>
            </div>
            {createdAt && (
              <>
                <Separator />
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground flex items-center gap-1.5 text-sm">
                    <Calendar className="size-3.5" />
                    {t("createdAt")}
                  </span>
                  <span className="text-sm">{createdAt}</span>
                </div>
              </>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
