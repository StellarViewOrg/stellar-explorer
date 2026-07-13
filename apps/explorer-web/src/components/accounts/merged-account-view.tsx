"use client";

import { GitMerge, ArrowRight, ExternalLink } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { HashDisplay } from "@/components/common/hash-display";
import { TimeAgo } from "@/components/common/time-ago";
import { PageHeader } from "@/components/layout/page-header";
import { Breadcrumbs } from "@/components/common/breadcrumbs";
import { Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";
export type AccountMergeOp = {
  type: "account_merge";
  account: string;
  into: string;
  created_at: string;
  transaction_hash: string;
  id: string;
};

interface MergedAccountViewProps {
  id: string;
  mergeOp: AccountMergeOp;
}

export function MergedAccountView({ id, mergeOp }: MergedAccountViewProps) {
  const t = useTranslations("account");
  const tNav = useTranslations("navigation");

  return (
    <div className="space-y-6">
      <Breadcrumbs
        items={[
          { label: tNav("accounts"), href: "/accounts" },
          { label: `${id.slice(0, 6)}...${id.slice(-6)}`, href: `/account/${id}` },
        ]}
      />

      <PageHeader
        title={t("title")}
        hash={id}
        backHref="/accounts"
        backLabel={tNav("accounts")}
        showQr
      />

      <div className="flex items-center gap-3">
        <Badge variant="destructive" className="gap-1.5 text-sm">
          <GitMerge className="size-3.5" />
          {t("merged")}
        </Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <GitMerge className="text-destructive size-4" />
            {t("merged")}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-muted-foreground text-sm">{t("mergedDescription")}</p>

          <Separator />

          <div className="flex items-center justify-between">
            <span className="text-muted-foreground text-sm">{t("mergedInto")}</span>
            <div className="flex items-center gap-2">
              <HashDisplay
                hash={mergeOp.into}
                truncate
                startLength={6}
                endLength={6}
                linkTo={`/account/${mergeOp.into}`}
                className="text-sm"
              />
              <ArrowRight className="text-muted-foreground size-3.5" />
            </div>
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <span className="text-muted-foreground text-sm">{t("mergeDate")}</span>
            <TimeAgo timestamp={mergeOp.created_at} className="text-sm" />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <span className="text-muted-foreground text-sm">{t("viewMergeTransaction")}</span>
            <Button variant="outline" size="sm" asChild>
              <Link href={`/tx/${mergeOp.transaction_hash}`}>
                <ExternalLink className="mr-1.5 size-3.5" />
                {mergeOp.transaction_hash.slice(0, 8)}…{mergeOp.transaction_hash.slice(-6)}
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
