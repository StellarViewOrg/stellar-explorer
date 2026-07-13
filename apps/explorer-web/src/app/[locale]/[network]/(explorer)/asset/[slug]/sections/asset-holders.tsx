"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Users } from "lucide-react";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { HashDisplay } from "@/components/common/hash-display";
import { useAssetAccounts } from "@/lib/hooks";
import { formatNumber } from "@/lib/utils";
import type { Horizon } from "@stellar/stellar-sdk";

export function AssetHolders({ code, issuer }: { code: string; issuer: string }) {
  const { data, isLoading, error } = useAssetAccounts(code, issuer);

  if (isLoading) return <LoadingCard rows={5} />;
  if (error) return <ErrorState title="Failed to load holders" message={error.message} />;
  if (!data?.records?.length)
    return <EmptyState title="No holders found" description="This asset has no trustlines yet." />;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Users className="size-4" />
          Top Holders
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {data.records.map((account: Horizon.ServerApi.AccountRecord) => {
            const balance = account.balances.find(
              (b) =>
                b.asset_type !== "native" &&
                (b as Horizon.HorizonApi.BalanceLineAsset).asset_code === code &&
                (b as Horizon.HorizonApi.BalanceLineAsset).asset_issuer === issuer
            ) as Horizon.HorizonApi.BalanceLineAsset | undefined;

            return (
              <div
                key={account.id}
                className="bg-card/50 flex items-center justify-between rounded-lg p-3"
              >
                <HashDisplay
                  hash={account.id}
                  truncate
                  linkTo={`/account/${account.id}`}
                  className="text-sm"
                />
                <span className="font-mono text-sm tabular-nums">
                  {balance ? formatNumber(balance.balance) : "—"} {code}
                </span>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
