"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { useContractStorage } from "@/lib/hooks";
import { formatCompactNumber } from "@/lib/utils";
import { useTranslations } from "next-intl";
import { Database, Key, Clock, HardDrive, ChevronDown, ChevronUp, Check, Copy } from "lucide-react";

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return "null";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "bigint") return value.toString();
  if (typeof value === "boolean") return value ? "true" : "false";
  if (value instanceof Uint8Array)
    return Array.from(value)
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");
  if (Array.isArray(value)) return JSON.stringify(value, null, 2);
  if (typeof value === "object") return JSON.stringify(value, null, 2);
  return String(value);
}

function getTypeColor(type: string): string {
  switch (type) {
    case "scvSymbol":
    case "scvString":
      return "text-chart-2";
    case "scvI128":
    case "scvI64":
    case "scvI32":
    case "scvU128":
    case "scvU64":
    case "scvU32":
      return "text-chart-1";
    case "scvBool":
      return "text-chart-4";
    case "scvAddress":
      return "text-chart-3";
    case "scvBytes":
      return "text-chart-5";
    default:
      return "text-muted-foreground";
  }
}

export function ContractStorage({ contractId }: { contractId: string }) {
  const t = useTranslations("contract");
  const { data, isLoading, error, refetch } = useContractStorage(contractId);
  const [expandedEntries, setExpandedEntries] = useState<Set<number>>(new Set());
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const toggleExpanded = (index: number) => {
    const newExpanded = new Set(expandedEntries);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedEntries(newExpanded);
  };

  const handleCopyValue = async (value: string, index: number) => {
    await navigator.clipboard.writeText(value);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  if (isLoading) return <LoadingCard rows={5} />;

  if (error) {
    return (
      <ErrorState title={t("failedToLoadStorage")} message={error.message} onRetry={refetch} />
    );
  }

  if (!data || data.totalEntries === 0) {
    return (
      <Card>
        <CardContent className="py-8">
          <div className="flex flex-col items-center justify-center gap-4 text-center">
            <div className="bg-muted flex size-16 items-center justify-center rounded-full">
              <Database className="text-muted-foreground size-8" />
            </div>
            <div>
              <h3 className="text-lg font-medium">{t("noStorage")}</h3>
              <p className="text-muted-foreground mt-1 text-sm">{t("noStorageDescription")}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <HardDrive className="size-4" />
            {t("instanceStorage")} ({data.totalEntries})
          </CardTitle>
          {data.ttlLedgers !== null && (
            <div className="flex items-center gap-2">
              <Clock className="text-muted-foreground size-4" />
              <span className="text-muted-foreground text-sm">
                {t("ttl")}: {formatCompactNumber(data.ttlLedgers)} {t("ledgers")}
              </span>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {data.instanceStorage.map((entry, index) => {
          const isExpanded = expandedEntries.has(index);
          const keyStr = formatValue(entry.key.value);
          const valueStr = formatValue(entry.value.value);
          const isLongValue = valueStr.length > 100 || valueStr.includes("\n");

          return (
            <div key={index} className="bg-muted/30 rounded-lg border p-3">
              {/* Key row */}
              <div className="flex items-start gap-3">
                <div className="bg-chart-1/10 flex size-8 shrink-0 items-center justify-center rounded-md">
                  <Key className="text-chart-1 size-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-muted-foreground text-xs">{t("key")}</span>
                    <Badge
                      variant="outline"
                      className={`text-[10px] ${getTypeColor(entry.key.type)}`}
                    >
                      {entry.key.type.replace("scv", "")}
                    </Badge>
                  </div>
                  <code className="mt-1 block truncate font-mono text-sm">{keyStr}</code>
                </div>
              </div>

              <Separator className="my-3" />

              {/* Value row */}
              <div className="flex items-start gap-3">
                <div className="bg-chart-2/10 flex size-8 shrink-0 items-center justify-center rounded-md">
                  <Database className="text-chart-2 size-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className="text-muted-foreground text-xs">{t("value")}</span>
                      <Badge
                        variant="outline"
                        className={`text-[10px] ${getTypeColor(entry.value.type)}`}
                      >
                        {entry.value.type.replace("scv", "")}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-6"
                        onClick={() => handleCopyValue(valueStr, index)}
                      >
                        {copiedIndex === index ? (
                          <Check className="size-3 text-green-500" />
                        ) : (
                          <Copy className="size-3" />
                        )}
                      </Button>
                      {isLongValue && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-6"
                          onClick={() => toggleExpanded(index)}
                        >
                          {isExpanded ? (
                            <ChevronUp className="size-3" />
                          ) : (
                            <ChevronDown className="size-3" />
                          )}
                        </Button>
                      )}
                    </div>
                  </div>
                  {isLongValue && !isExpanded ? (
                    <code className="mt-1 block truncate font-mono text-sm">
                      {valueStr.slice(0, 100)}...
                    </code>
                  ) : (
                    <pre className="bg-muted/50 mt-2 overflow-x-auto rounded p-2 font-mono text-xs">
                      {valueStr}
                    </pre>
                  )}
                </div>
              </div>

              {/* Durability badge */}
              <div className="mt-3 flex justify-end">
                <Badge variant="secondary" className="text-[10px]">
                  {entry.durability}
                </Badge>
              </div>
            </div>
          );
        })}

        {/* TTL Info */}
        {data.liveUntilLedger && (
          <div className="bg-muted/30 flex items-center justify-between rounded-lg border p-3">
            <div className="flex items-center gap-2">
              <Clock className="text-muted-foreground size-4" />
              <span className="text-muted-foreground text-sm">{t("liveUntilLedger")}</span>
            </div>
            <span className="font-mono text-sm">#{formatCompactNumber(data.liveUntilLedger)}</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
