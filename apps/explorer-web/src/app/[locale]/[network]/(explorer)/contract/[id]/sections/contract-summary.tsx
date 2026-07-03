"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Hash, Box, Coins } from "lucide-react";
import { HashDisplay } from "@/components/common/hash-display";
import { ContractVerification } from "@/components/contracts";
import { useContractBalance, useContractCode } from "@/lib/hooks";
import { useTranslations } from "next-intl";

export function ContractSummary({ contractId }: { contractId: string }) {
  const t = useTranslations("contract");
  const { data: balanceData } = useContractBalance(contractId);
  const { data: codeData } = useContractCode(contractId);
  const isSac = codeData?.type === "sac";

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between text-base">
          {t("information")}
          {isSac && (
            <Badge className="bg-chart-4/15 text-chart-4 border-chart-4/25">
              <Box className="mr-1.5 size-3" />
              {t("stellarAssetContract")}
            </Badge>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          {/* Left column */}
          <div className="space-y-4">
            <div className="flex items-start justify-between">
              <span className="text-muted-foreground flex items-center gap-2 text-sm">
                <Hash className="size-4" />
                {t("contractId")}
              </span>
              <HashDisplay
                hash={contractId}
                truncate
                startLength={12}
                endLength={8}
                className="text-sm"
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("type")}</span>
              <Badge variant="secondary">{t("sorobanContract")}</Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground text-sm">{t("status")}</span>
              <Badge className="bg-success/15 text-success border-success/25">{t("active")}</Badge>
            </div>
            {balanceData?.exists && parseFloat(balanceData.balance) > 0 && (
              <>
                <Separator />
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground flex items-center gap-2 text-sm">
                    <Coins className="size-4" />
                    XLM Balance
                  </span>
                  <span className="font-mono text-sm font-medium">
                    {parseFloat(balanceData.balance).toLocaleString(undefined, {
                      maximumFractionDigits: 7,
                    })}{" "}
                    XLM
                  </span>
                </div>
              </>
            )}
          </div>

          {/* Right column - Enhanced Verification Status */}
          <div className="space-y-4">
            <ContractVerification contractId={contractId} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
