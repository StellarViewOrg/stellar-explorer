"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { useContractCode } from "@/lib/hooks";
import { formatCompactNumber } from "@/lib/utils";
import { useTranslations } from "next-intl";
import { FileCode, Box, Check, Copy } from "lucide-react";

function formatHexCode(hex: string, fullView: boolean): string[] {
  const lines: string[] = [];
  const charsPerLine = 64;
  const maxLines = fullView ? Infinity : 20;

  for (let i = 0; i < hex.length && lines.length < maxLines; i += charsPerLine) {
    const line = hex.slice(i, i + charsPerLine);
    const formatted = line.match(/.{1,8}/g)?.join(" ") || line;
    lines.push(formatted);
  }

  return lines;
}

export function ContractCode({ contractId }: { contractId: string }) {
  const t = useTranslations("contract");
  const { data, isLoading, error, refetch } = useContractCode(contractId);
  const [copied, setCopied] = useState<"hash" | "code" | null>(null);
  const [showFullCode, setShowFullCode] = useState(false);

  const handleCopy = async (text: string, type: "hash" | "code") => {
    await navigator.clipboard.writeText(text);
    setCopied(type);
    setTimeout(() => setCopied(null), 2000);
  };

  if (isLoading) return <LoadingCard rows={5} />;

  if (error) {
    return <ErrorState title={t("failedToLoadCode")} message={error.message} onRetry={refetch} />;
  }

  if (!data) {
    return <EmptyState title={t("noCode")} description={t("noCodeDescription")} icon="file" />;
  }

  if (data.type === "sac") {
    return (
      <Card>
        <CardContent className="py-8">
          <div className="flex flex-col items-center justify-center gap-4 text-center">
            <div className="bg-chart-4/10 flex size-16 items-center justify-center rounded-full">
              <Box className="text-chart-4 size-8" />
            </div>
            <div>
              <h3 className="text-lg font-medium">{t("stellarAssetContract")}</h3>
              <p className="text-muted-foreground mt-1 text-sm">{t("sacDescription")}</p>
            </div>
            <Badge variant="secondary">{t("nativeContract")}</Badge>
          </div>
        </CardContent>
      </Card>
    );
  }

  const codeLines = formatHexCode(data.wasmCodeHex || "", showFullCode);
  const totalLines = Math.ceil((data.wasmCodeHex?.length || 0) / 64);
  const hasMoreLines = totalLines > 20 && !showFullCode;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <FileCode className="size-4" />
          {t("wasmBytecode")}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* WASM Info */}
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <span className="text-muted-foreground text-sm">{t("wasmHash")}</span>
            <div className="flex items-center gap-2">
              <code className="bg-muted truncate rounded px-2 py-1 font-mono text-xs">
                {data.wasmHash?.slice(0, 16)}...{data.wasmHash?.slice(-16)}
              </code>
              <Button
                variant="ghost"
                size="icon"
                className="size-7 shrink-0"
                onClick={() => handleCopy(data.wasmHash || "", "hash")}
              >
                {copied === "hash" ? (
                  <Check className="size-3.5 text-green-500" />
                ) : (
                  <Copy className="size-3.5" />
                )}
              </Button>
            </div>
          </div>
          <div className="space-y-2">
            <span className="text-muted-foreground text-sm">{t("codeSize")}</span>
            <div className="flex items-center gap-2">
              <Badge variant="outline" className="font-mono">
                {formatCompactNumber(data.codeSize)} {t("bytes")}
              </Badge>
              <span className="text-muted-foreground text-xs">
                ({(data.codeSize / 1024).toFixed(2)} KB)
              </span>
            </div>
          </div>
        </div>

        <Separator />

        {/* Bytecode viewer */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground text-sm">{t("hexBytecode")}</span>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1.5 text-xs"
              onClick={() => handleCopy(data.wasmCodeHex || "", "code")}
            >
              {copied === "code" ? (
                <>
                  <Check className="size-3.5 text-green-500" />
                  {t("copied")}
                </>
              ) : (
                <>
                  <Copy className="size-3.5" />
                  {t("copyCode")}
                </>
              )}
            </Button>
          </div>
          <div className="bg-muted/50 relative rounded-lg border">
            <div className="overflow-x-auto">
              <pre className="p-4 font-mono text-xs leading-relaxed">
                {codeLines.map((line, index) => (
                  <div key={index} className="flex">
                    <span className="text-muted-foreground mr-4 w-8 text-right select-none">
                      {String((index + 1) * 32).padStart(6, "0")}
                    </span>
                    <span className="text-foreground">{line}</span>
                  </div>
                ))}
              </pre>
            </div>
            {hasMoreLines && (
              <div className="from-muted/80 to-muted absolute inset-x-0 bottom-0 flex h-20 items-end justify-center bg-gradient-to-t pb-4">
                <Button variant="secondary" size="sm" onClick={() => setShowFullCode(true)}>
                  {t("showFullCode")} ({totalLines - 20} {t("moreLines")})
                </Button>
              </div>
            )}
          </div>
          {showFullCode && totalLines > 20 && (
            <Button
              variant="ghost"
              size="sm"
              className="w-full text-xs"
              onClick={() => setShowFullCode(false)}
            >
              {t("collapseCode")}
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
