"use client";

import { memo } from "react";
import { Link } from "@/i18n/navigation";
import { Card, CardContent } from "@/components/ui/card";
import { HashDisplay } from "@/components/common/hash-display";
import { formatBalance } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { Users } from "lucide-react";

interface AccountCardProps {
  accountId: string;
  balance?: string;
  className?: string;
}

export const AccountCard = memo(function AccountCard({
  accountId,
  balance,
  className,
}: AccountCardProps) {
  return (
    <Link href={`/account/${accountId}`}>
      <Card variant="elevated" interactive className={cn("group border-0 py-0", className)}>
        <CardContent className="flex items-center justify-between p-4">
          <div className="flex min-w-0 items-center gap-3">
            <div className="bg-chart-2/10 relative flex size-10 items-center justify-center rounded-xl">
              <div className="bg-chart-2/20 absolute inset-0 rounded-xl opacity-0 blur-md transition-opacity duration-300 group-hover:opacity-60" />
              <Users className="text-chart-2 relative size-5" />
            </div>
            <HashDisplay
              hash={accountId}
              truncate
              startLength={8}
              endLength={4}
              copyable={false}
              className="text-sm font-medium"
            />
          </div>
          {balance && (
            <span className="text-muted-foreground shrink-0 text-sm tabular-nums">
              {formatBalance(balance)}
            </span>
          )}
        </CardContent>
      </Card>
    </Link>
  );
});

export function AccountCardSkeleton({ className }: { className?: string }) {
  return (
    <Card variant="elevated" className={cn("border-0 py-0", className)}>
      <CardContent className="flex items-center justify-between p-4">
        <div className="flex items-center gap-3">
          <div className="bg-muted/50 size-10 animate-pulse rounded-xl" />
          <div className="bg-muted/50 h-4 w-28 animate-pulse rounded" />
        </div>
        <div className="bg-muted/50 h-4 w-20 animate-pulse rounded" />
      </CardContent>
    </Card>
  );
}
