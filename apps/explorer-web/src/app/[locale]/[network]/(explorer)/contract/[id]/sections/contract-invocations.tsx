"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { HashDisplay } from "@/components/common/hash-display";
import { TimeAgo } from "@/components/common/time-ago";
import { StatusBadge } from "@/components/common/status-badge";
import { useContractInvocations } from "@/lib/hooks";
import { Zap } from "lucide-react";
import type { Horizon } from "@stellar/stellar-sdk";

export function ContractInvocations({ contractId }: { contractId: string }) {
  const { data, isLoading, error, refetch } = useContractInvocations(contractId);

  if (isLoading) return <LoadingCard rows={5} />;
  if (error)
    return (
      <ErrorState title="Failed to load invocations" message={error.message} onRetry={refetch} />
    );

  const invocations = data?.records ?? [];

  if (invocations.length === 0) {
    return (
      <EmptyState
        title="No invocations found"
        description="No InvokeHostFunction operations found for this contract in recent history."
        icon="search"
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Zap className="size-4" />
          Contract Invocations
          <Badge variant="secondary" className="ml-auto">
            {invocations.length}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <div className="divide-border divide-y">
          {invocations.map((op) => {
            const operation = op as Horizon.ServerApi.InvokeHostFunctionOperationRecord;
            return (
              <div
                key={operation.id}
                className="hover:bg-muted/30 flex items-center justify-between px-6 py-3 transition-colors"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <div
                    className={`flex size-8 shrink-0 items-center justify-center rounded-lg ${operation.transaction_successful ? "bg-success/10" : "bg-destructive/10"}`}
                  >
                    <Zap
                      className={`size-4 ${operation.transaction_successful ? "text-success" : "text-destructive"}`}
                    />
                  </div>
                  <div className="min-w-0 space-y-0.5">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">
                        {operation.function?.includes("HostFunctionTypeUploadContractWasm")
                          ? "Upload WASM"
                          : operation.function?.includes("HostFunctionTypeCreateContract")
                            ? "Create Contract"
                            : "Invoke Contract"}
                      </span>
                      <StatusBadge
                        status={operation.transaction_successful ? "success" : "failed"}
                        showLabel={false}
                      />
                    </div>
                    <div className="text-muted-foreground flex items-center gap-2 text-xs">
                      <span>from</span>
                      <HashDisplay
                        hash={operation.source_account}
                        truncate
                        startLength={4}
                        endLength={4}
                        copyable={false}
                        linkTo={`/account/${operation.source_account}`}
                        className="text-xs"
                      />
                    </div>
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-3">
                  <TimeAgo
                    timestamp={operation.created_at}
                    className="text-muted-foreground text-xs"
                  />
                  <HashDisplay
                    hash={operation.transaction_hash}
                    truncate
                    startLength={4}
                    endLength={4}
                    copyable={false}
                    linkTo={`/tx/${operation.transaction_hash}`}
                    className="font-mono text-xs opacity-60"
                  />
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
