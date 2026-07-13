"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Shield, Lock, Unlock } from "lucide-react";
import { useTranslations } from "next-intl";
import { FlagBadge } from "./flag-badge";
import type { AssetRecordExtended } from "./types";

export function AssetFlags({ asset }: { asset: AssetRecordExtended }) {
  const flags = asset.flags;
  const t = useTranslations("assetDetails");

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Shield className="size-4" />
          {t("flags")}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        <FlagBadge
          enabled={flags.auth_required}
          label={t("authRequired")}
          enabledIcon={Lock}
          disabledIcon={Unlock}
        />
        <FlagBadge
          enabled={flags.auth_revocable}
          label={t("authRevocable")}
          enabledIcon={Lock}
          disabledIcon={Unlock}
        />
        <FlagBadge
          enabled={flags.auth_immutable}
          label={t("authImmutable")}
          enabledIcon={Lock}
          disabledIcon={Unlock}
        />
        <FlagBadge enabled={flags.auth_clawback_enabled} label={t("clawbackEnabled")} />
      </CardContent>
    </Card>
  );
}
