"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { Horizon } from "@stellar/stellar-sdk";
import { useTranslations } from "next-intl";

export function RawData({ transaction }: { transaction: Horizon.ServerApi.TransactionRecord }) {
  const t = useTranslations("transaction");

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{t("rawData")}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div>
            <h4 className="mb-2 text-sm font-medium">{t("envelopeXdr")}</h4>
            <pre className="bg-muted/50 overflow-x-auto rounded-lg p-3 font-mono text-xs break-all whitespace-pre-wrap">
              {transaction.envelope_xdr}
            </pre>
          </div>
          <div>
            <h4 className="mb-2 text-sm font-medium">{t("resultXdr")}</h4>
            <pre className="bg-muted/50 overflow-x-auto rounded-lg p-3 font-mono text-xs break-all whitespace-pre-wrap">
              {transaction.result_xdr}
            </pre>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
