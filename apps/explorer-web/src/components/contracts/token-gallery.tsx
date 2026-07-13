"use client";

import { useMemo, useState } from "react";
import { Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";
import { useContractEvents } from "@/lib/hooks";
import { LoadingCard } from "@/components/common/loading-card";
import { EmptyState } from "@/components/common/empty-state";
import { TimeAgo } from "@/components/common/time-ago";
import { truncateHash } from "@/lib/utils/format";
import { ArrowRightLeft, Flame, Coins, Zap, ExternalLink } from "lucide-react";

interface TokenGalleryProps {
  contractId: string;
}

interface GalleryItem {
  id: string;
  pattern: "transfer" | "mint" | "burn" | "approve" | "event";
  txHash: string;
  ledger: number;
  timestamp: string;
  from?: string;
  to?: string;
  amount?: string;
  tokenId?: string;
  label: string;
  color1: string;
  color2: string;
  color3: string;
}

const PATTERN_CONFIG = {
  transfer: { icon: ArrowRightLeft, label: "Transfer", hue: 215 },
  mint: { icon: Coins, label: "Mint", hue: 145 },
  burn: { icon: Flame, label: "Burn", hue: 20 },
  approve: { icon: Zap, label: "Approve", hue: 290 },
  event: { icon: Zap, label: "Event", hue: 185 },
} as const;

function hashToColors(hash: string): [string, string, string] {
  const h1 = parseInt(hash.slice(0, 2), 16) / 255;
  const h2 = parseInt(hash.slice(2, 4), 16) / 255;
  const h3 = parseInt(hash.slice(4, 6), 16) / 255;
  const toOklch = (l: number, base: number) =>
    `oklch(${0.4 + l * 0.3} ${0.18 + base * 0.16} ${Math.round(base * 360)})`;
  return [toOklch(h1, h1), toOklch(h2, h2), toOklch(h3, h3)];
}

function detectPattern(topics: Array<{ type: string; value: unknown }>): string {
  const first = topics[0]?.value;
  if (typeof first === "string") {
    const lower = first.toLowerCase();
    if (lower.includes("transfer")) return "transfer";
    if (lower.includes("mint")) return "mint";
    if (lower.includes("burn")) return "burn";
    if (lower.includes("approv")) return "approve";
  }
  return "event";
}

function extractField(
  topics: Array<{ type: string; value: unknown }>,
  value: { type: string; value: unknown } | null,
  index: number
): string | undefined {
  const raw = topics[index]?.value ?? (index === 0 ? value?.value : undefined);
  if (typeof raw === "string" && raw.length > 0) return raw;
  if (typeof raw === "bigint" || typeof raw === "number") return String(raw);
  return undefined;
}

function TokenCard({ item }: { item: GalleryItem }) {
  const config = PATTERN_CONFIG[item.pattern];
  const Icon = config.icon;
  const [flipped, setFlipped] = useState(false);

  return (
    <div
      className="card-3d-scene aspect-[3/4] w-full cursor-pointer"
      onClick={() => setFlipped((f) => !f)}
    >
      <div
        className="card-3d rounded-2xl"
        style={flipped ? { transform: "rotateY(180deg)" } : undefined}
      >
        {/* ── Front face ── */}
        <div className="card-face border-border/20 border bg-[var(--card)]">
          {/* Generative gradient art */}
          <div
            className="relative h-[65%] overflow-hidden"
            style={{
              background: `radial-gradient(ellipse at 20% 30%, ${item.color1} 0%, transparent 60%),
                           radial-gradient(ellipse at 80% 70%, ${item.color2} 0%, transparent 55%),
                           radial-gradient(ellipse at 50% 100%, ${item.color3} 0%, transparent 50%),
                           var(--background)`,
            }}
          >
            {/* Abstract grid overlay */}
            <div
              className="absolute inset-0 opacity-10"
              style={{
                backgroundImage: `linear-gradient(${item.color1} 1px, transparent 1px),
                                  linear-gradient(90deg, ${item.color2} 1px, transparent 1px)`,
                backgroundSize: "24px 24px",
              }}
            />
            {/* Center icon */}
            <div className="absolute inset-0 flex items-center justify-center">
              <div
                className="flex size-14 items-center justify-center rounded-2xl"
                style={{ background: `${item.color1}33`, border: `1px solid ${item.color1}66` }}
              >
                <Icon className="size-7 text-white/80" />
              </div>
            </div>
            {/* Token ID badge */}
            {item.tokenId && (
              <div className="absolute top-3 right-3 rounded-lg bg-black/50 px-2 py-1 font-mono text-[10px] text-white/70 backdrop-blur-sm">
                #{item.tokenId.slice(0, 8)}
              </div>
            )}
          </div>

          {/* Metadata strip */}
          <div className="border-border/20 border-t p-3">
            <div className="mb-1 flex items-center justify-between">
              <span
                className="rounded-md px-1.5 py-0.5 text-[10px] font-semibold tracking-wider uppercase"
                style={{ background: `${item.color1}22`, color: item.color1 }}
              >
                {config.label}
              </span>
              <TimeAgo timestamp={item.timestamp} className="text-muted-foreground text-[10px]" />
            </div>
            <p className="text-foreground/70 truncate font-mono text-xs">
              {truncateHash(item.txHash, 5, 4)}
            </p>
          </div>
        </div>

        {/* ── Back face ── */}
        <div
          className="card-face card-face-back border-border/20 flex flex-col justify-between border p-4"
          style={{
            background: `linear-gradient(160deg, var(--card) 0%, oklch(from ${item.color1} l c h / 0.08) 100%)`,
          }}
        >
          <div>
            <div className="mb-3 flex items-center gap-2">
              <Icon className="size-4" style={{ color: item.color1 }} />
              <span className="text-sm font-semibold">{config.label}</span>
            </div>

            <div className="space-y-2 text-xs">
              {item.from && (
                <div>
                  <p className="text-muted-foreground mb-0.5">From</p>
                  <p className="font-mono opacity-80">{truncateHash(item.from, 5, 4)}</p>
                </div>
              )}
              {item.to && (
                <div>
                  <p className="text-muted-foreground mb-0.5">To</p>
                  <p className="font-mono opacity-80">{truncateHash(item.to, 5, 4)}</p>
                </div>
              )}
              {item.amount && (
                <div>
                  <p className="text-muted-foreground mb-0.5">Amount</p>
                  <p className="font-mono font-medium">{item.amount}</p>
                </div>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <div className="border-border/30 border-t pt-2">
              <p className="text-muted-foreground text-[10px]">Ledger #{item.ledger}</p>
            </div>
            <Link
              href={`/tx/${item.txHash}`}
              className="text-primary flex items-center gap-1 text-xs font-medium hover:underline"
              onClick={(e) => e.stopPropagation()}
            >
              View transaction
              <ExternalLink className="size-3" />
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}

export function TokenGallery({ contractId }: TokenGalleryProps) {
  const { data, isLoading } = useContractEvents(contractId);
  const t = useTranslations("contract.tokenGallery");

  const items = useMemo<GalleryItem[]>(() => {
    if (!data?.events) return [];

    return data.events
      .filter((e) => e.decodedTopics.length > 0)
      .slice(0, 24)
      .map((event): GalleryItem => {
        const pattern = detectPattern(
          event.decodedTopics as Array<{ type: string; value: unknown }>
        ) as GalleryItem["pattern"];

        const hash = event.txHash ?? event.id ?? "000000";
        const [c1, c2, c3] = hashToColors(hash.replace(/^0x/, "").padEnd(6, "0"));

        const topics = event.decodedTopics as Array<{ type: string; value: unknown }>;
        const val = event.decodedValue as { type: string; value: unknown } | null;

        const from = extractField(topics, val, 1);
        const to = extractField(topics, val, 2);
        const amount = extractField(topics, val, 3) ?? (val?.value ? String(val.value) : undefined);

        return {
          id: event.id ?? `event-${hash}-${event.ledger ?? 0}`,
          pattern,
          txHash: event.txHash ?? "",
          ledger: event.ledger ?? 0,
          timestamp: event.ledgerClosedAt ?? new Date().toISOString(),
          from,
          to,
          amount,
          tokenId: hash.slice(0, 16),
          label: PATTERN_CONFIG[pattern].label,
          color1: c1,
          color2: c2,
          color3: c3,
        };
      });
  }, [data]);

  if (isLoading) return <LoadingCard rows={4} />;

  if (items.length === 0) {
    return <EmptyState title={t("noActivity")} description={t("noActivityDesc")} icon="search" />;
  }

  return (
    <div>
      <p className="text-muted-foreground mb-4 text-sm">
        {t("recentEvents", { count: items.length })}
      </p>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {items.map((item) => (
          <TokenCard key={item.id} item={item} />
        ))}
      </div>
    </div>
  );
}
