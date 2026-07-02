"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { HashDisplay } from "@/components/common/hash-display";
import { formatNumber } from "@/lib/utils";
import { Coins } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function AccountBalances({ account }: { account: Horizon.ServerApi.AccountRecord }) {
  const t = useTranslations("account");

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">
          {t("balances")} ({account.balances.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {account.balances.map((balance) => {
            const isNative = balance.asset_type === "native";
            const assetCode = isNative
              ? "XLM"
              : (balance as Horizon.HorizonApi.BalanceLineAsset).asset_code;
            const issuer = isNative
              ? null
              : (balance as Horizon.HorizonApi.BalanceLineAsset).asset_issuer;

            return (
              <div
                key={`${balance.asset_type}-${assetCode}-${issuer || "native"}`}
                className="bg-card/50 flex items-center justify-between rounded-lg p-3"
              >
                <div className="flex items-center gap-3">
                  <div className="bg-chart-3/10 flex size-8 items-center justify-center rounded-md">
                    <Coins className="text-chart-3 size-4" />
                  </div>
                  <div>
                    <span className="font-medium">{assetCode}</span>
                    {issuer && (
                      <div className="text-muted-foreground text-xs">
                        <HashDisplay
                          hash={issuer}
                          truncate
                          startLength={4}
                          endLength={4}
                          copyable={false}
                        />
                      </div>
                    )}
                  </div>
                </div>
                <span className="font-mono tabular-nums">{formatNumber(balance.balance)}</span>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
