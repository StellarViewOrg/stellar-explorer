"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { HashDisplay } from "@/components/common/hash-display";
import { Key } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function AccountSigners({ account }: { account: Horizon.ServerApi.AccountRecord }) {
  const t = useTranslations("account");

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">
          {t("signers")} ({account.signers.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {account.signers.map((signer) => (
            <div
              key={signer.key}
              className="bg-card/50 flex items-center justify-between rounded-lg p-3"
            >
              <div className="flex items-center gap-3">
                <div className="bg-primary/10 flex size-8 items-center justify-center rounded-md">
                  <Key className="text-primary size-4" />
                </div>
                <div>
                  <HashDisplay
                    hash={signer.key}
                    truncate
                    linkTo={
                      signer.type === "ed25519_public_key" ? `/account/${signer.key}` : undefined
                    }
                    className="text-sm"
                  />
                  <span className="text-muted-foreground text-xs capitalize">
                    {signer.type.replace(/_/g, " ")}
                  </span>
                </div>
              </div>
              <Badge variant="outline">{t("weight", { weight: signer.weight })}</Badge>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
