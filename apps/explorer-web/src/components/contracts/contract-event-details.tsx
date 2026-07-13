"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { HashDisplay } from "@/components/common/hash-display";
import { LoadingCard } from "@/components/common/loading-card";
import { ErrorState } from "@/components/common/error-state";
import { EmptyState } from "@/components/common/empty-state";
import { useContractEvents } from "@/lib/hooks";
import {
  detectEventPattern,
  categorizeEventType,
  getEventCategoryColor,
  formatScValDisplay,
  isAddress,
  getAddressLink,
} from "@/lib/utils/soroban-helpers";
import type { DecodedScVal } from "@/lib/utils/soroban-helpers";
import { Activity, ArrowRightLeft, Flame, Coins, ShieldCheck, Zap } from "lucide-react";

interface DecodedEvent {
  id?: string;
  type?: string;
  ledger?: number;
  txHash?: string;
  decodedTopics: DecodedScVal[];
  decodedValue: DecodedScVal | null;
}

function getPatternIcon(pattern: string) {
  switch (pattern) {
    case "transfer":
      return <ArrowRightLeft className="size-4" />;
    case "mint":
      return <Coins className="size-4" />;
    case "burn":
      return <Flame className="size-4" />;
    case "approve":
      return <ShieldCheck className="size-4" />;
    default:
      return <Zap className="size-4" />;
  }
}

function TopicValue({ topic }: { topic: DecodedScVal }) {
  const display = formatScValDisplay(topic.type, topic.value);
  const addressValue = isAddress(topic.value);

  if (addressValue && typeof topic.value === "string") {
    return (
      <HashDisplay
        hash={topic.value}
        truncate
        startLength={6}
        endLength={4}
        linkTo={getAddressLink(topic.value)}
        className="text-xs"
      />
    );
  }

  return (
    <code className="bg-muted rounded px-1.5 py-0.5 font-mono text-xs">
      {display.length > 40 ? display.slice(0, 37) + "..." : display}
    </code>
  );
}

function EventRow({ event }: { event: DecodedEvent }) {
  const t = useTranslations("contract");
  const category = categorizeEventType(event.decodedTopics);
  const colors = getEventCategoryColor(category);
  const { pattern, summary } = detectEventPattern(event.decodedTopics);

  return (
    <div className="bg-card/50 space-y-3 rounded-lg border p-4">
      {/* Header row */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-3">
          <div className={`flex size-8 items-center justify-center rounded-md ${colors.bg}`}>
            <span className={colors.text}>{getPatternIcon(pattern)}</span>
          </div>
          <div className="min-w-0">
            <p className="truncate text-sm font-medium">{summary}</p>
            <p className="text-muted-foreground text-xs">
              {t("ledger")} {event.ledger}
            </p>
          </div>
        </div>
        <Badge variant="outline" className={`shrink-0 ${colors.text} ${colors.border}`}>
          {pattern !== "unknown"
            ? t(
                `event${pattern.charAt(0).toUpperCase() + pattern.slice(1)}` as
                  | "eventTransfer"
                  | "eventMint"
                  | "eventBurn"
                  | "eventApprove"
              )
            : t("contractEvent")}
        </Badge>
      </div>

      {/* TX Hash */}
      {event.txHash && (
        <div className="flex items-center gap-2">
          <span className="text-muted-foreground text-xs">{t("eventTxHash")}:</span>
          <HashDisplay
            hash={event.txHash}
            truncate
            startLength={8}
            endLength={6}
            linkTo={`tx/${event.txHash}`}
            className="text-xs"
          />
        </div>
      )}

      {/* Topics */}
      {event.decodedTopics.length > 0 && (
        <div className="space-y-1.5">
          <span className="text-muted-foreground text-xs font-medium">{t("eventTopics")}</span>
          <div className="flex flex-wrap gap-1.5">
            {event.decodedTopics.map((topic, i) => (
              <div key={i} className="flex items-center gap-1">
                <Badge variant="secondary" className="text-[10px]">
                  {topic.type.replace("scv", "")}
                </Badge>
                <TopicValue topic={topic} />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Value */}
      {event.decodedValue && (
        <div className="space-y-1.5">
          <span className="text-muted-foreground text-xs font-medium">{t("eventValue")}</span>
          <div className="flex items-center gap-1">
            <Badge variant="secondary" className="text-[10px]">
              {event.decodedValue.type.replace("scv", "")}
            </Badge>
            <TopicValue topic={event.decodedValue} />
          </div>
        </div>
      )}
    </div>
  );
}

export function ContractEventDetails({
  contractId,
  live = false,
}: {
  contractId: string;
  live?: boolean;
}) {
  const t = useTranslations("contract");
  const { data, isLoading, error, refetch } = useContractEvents(contractId, undefined, live);

  if (isLoading) {
    return <LoadingCard rows={5} />;
  }

  if (error) {
    return <ErrorState title={t("failedToLoadEvents")} message={error.message} onRetry={refetch} />;
  }

  if (!data?.events?.length) {
    return (
      <EmptyState title={t("noEvents")} description={t("noEventsDescription")} icon="search" />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Activity className="size-4" />
          {t("recentEvents")} ({data.events.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {data.events.map((event: DecodedEvent, index: number) => (
            <EventRow key={event.id || index} event={event} />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
