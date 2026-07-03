"use client";

import { Badge } from "@/components/ui/badge";
import { CheckCircle2, XCircle } from "lucide-react";
import { useTranslations } from "next-intl";

export function FlagBadge({
  enabled,
  label,
  enabledIcon: EnabledIcon = CheckCircle2,
  disabledIcon: DisabledIcon = XCircle,
}: {
  enabled: boolean;
  label: string;
  enabledIcon?: typeof CheckCircle2;
  disabledIcon?: typeof XCircle;
}) {
  const t = useTranslations("assetDetails");

  return (
    <div className="bg-card/50 flex items-center justify-between rounded-lg p-3">
      <span className="text-sm">{label}</span>
      <Badge
        variant="outline"
        className={
          enabled
            ? "bg-success/15 text-success border-success/25"
            : "bg-muted text-muted-foreground"
        }
      >
        {enabled ? (
          <EnabledIcon className="mr-1 size-3" />
        ) : (
          <DisabledIcon className="mr-1 size-3" />
        )}
        {enabled ? t("enabled") : t("disabled")}
      </Badge>
    </div>
  );
}
