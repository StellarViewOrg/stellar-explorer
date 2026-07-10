"use client";

import { Suspense, useEffect, useRef, useState } from "react";
import {
  Layers,
  Activity,
  Wallet,
  ArrowRight,
  FileCode,
  ArrowRightLeft,
  Users,
  Coins,
} from "lucide-react";
import { TransactionCard, TransactionCardSkeleton } from "@/components/cards/transaction-card";
import { LedgerCard, LedgerCardSkeleton } from "@/components/cards/ledger-card";
import { LiveIndicator } from "@/components/common/live-indicator";
import {
  useLatestLedger,
  useRecentTransactions,
  useFeeStats,
  useLedgerStream,
  useTransactionStream,
} from "@/lib/hooks";
import { formatLedgerSequence, stroopsToXLM } from "@/lib/utils";
import { useNetwork } from "@/lib/providers";
import { NetworkBadge } from "@/components/common/network-badge";
import { Link } from "@/i18n/navigation";
import { Button } from "@/components/ui/button";
import { GlobalSearch } from "@/components/search/global-search";
import { NetworkGlobe } from "@/components/common/network-globe";
import { useTranslations } from "next-intl";
import { DashboardCharts } from "@/components/charts";

/* ─── Animated counter ─── */
function AnimatedNumber({
  value,
  prefix = "",
  suffix = "",
}: {
  value: string | number;
  prefix?: string;
  suffix?: string;
}) {
  const [display, setDisplay] = useState("–");
  const prevRef = useRef<number | null>(null);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    if (value === "-" || value === "–") return;
    const numVal = typeof value === "string" ? parseFloat(value.replace(/,/g, "")) : value;
    if (isNaN(numVal)) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setDisplay(String(value));
      return;
    }

    if (prevRef.current === null) {
      setDisplay(String(value));
      prevRef.current = numVal;
      return;
    }

    const start = prevRef.current;
    const diff = numVal - start;
    const duration = 280;
    const startTime = performance.now();

    if (rafRef.current) cancelAnimationFrame(rafRef.current);

    const tick = (now: number) => {
      const elapsed = now - startTime;
      const progress = Math.min(elapsed / duration, 1);
      const ease = 1 - Math.pow(1 - progress, 3);
      const current = Math.round(start + diff * ease);
      setDisplay(current.toLocaleString());
      if (progress < 1) {
        rafRef.current = requestAnimationFrame(tick);
      } else {
        setDisplay(String(value));
      }
    };

    rafRef.current = requestAnimationFrame(tick);
    prevRef.current = numVal;
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [value]);

  return (
    <>
      {prefix}
      {display}
      {suffix}
    </>
  );
}

/* ─── Hero stats bar ─── */
function LiveStatsRibbon() {
  const { network } = useNetwork();
  const { data: ledger, isLoading: ledgerLoading } = useLatestLedger();
  const { data: feeStats, isLoading: feeLoading } = useFeeStats();
  const t = useTranslations("stats");
  const tNetwork = useTranslations("network");

  const avgFee = feeStats?.fee_charged?.mode ? stroopsToXLM(feeStats.fee_charged.mode) : "–";
  const networkLabel: Record<string, string> = {
    public: tNetwork("mainnet"),
    testnet: tNetwork("testnet"),
    futurenet: tNetwork("futurenet"),
  };

  const items = [
    {
      label: t("latestLedger"),
      value: ledger ? formatLedgerSequence(ledger.sequence) : "–",
      icon: Layers,
      loading: ledgerLoading,
      accent: "var(--chart-1)",
    },
    {
      label: "Transactions",
      value: ledger ? String(ledger.successful_transaction_count) : "–",
      suffix: " / ledger",
      icon: ArrowRightLeft,
      loading: ledgerLoading,
      accent: "var(--chart-2)",
    },
    {
      label: t("baseFee"),
      value: avgFee,
      suffix: " XLM",
      icon: Wallet,
      loading: feeLoading,
      accent: "var(--chart-3)",
    },
    {
      label: t("protocolVersion"),
      value: ledger?.protocol_version?.toString() ?? "–",
      icon: Activity,
      loading: ledgerLoading,
      accent: "var(--chart-4)",
    },
    {
      label: "Network",
      value: networkLabel[network] ?? network,
      icon: Activity,
      loading: false,
      accent: "var(--chart-5)",
    },
  ];

  return (
    <div className="border-border/40 border-b">
      <div className="divide-border/30 flex items-stretch divide-x overflow-x-auto">
        {items.map((item, i) => {
          const Icon = item.icon;
          return (
            <div
              key={i}
              className="flex min-w-[120px] flex-1 flex-col justify-center px-5 py-3"
              style={{ animationDelay: `${i * 60}ms` }}
            >
              <div className="text-muted-foreground mb-1 flex items-center gap-1.5 text-[11px] font-medium tracking-wider uppercase">
                <Icon className="size-3" style={{ color: item.accent }} />
                {item.label}
              </div>
              {item.loading ? (
                <div className="bg-muted/40 h-6 w-20 animate-pulse rounded" />
              ) : (
                <div className="text-foreground text-lg leading-none font-bold tabular-nums">
                  <AnimatedNumber value={item.value} suffix={item.suffix} />
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

/* ─── Recent transactions feed ─── */
function RecentTransactions() {
  const { data, isLoading } = useRecentTransactions(8);
  useTransactionStream({ enabled: true });

  if (isLoading) {
    return (
      <div className="space-y-2.5">
        {Array.from({ length: 5 }).map((_, i) => (
          <TransactionCardSkeleton key={i} />
        ))}
      </div>
    );
  }
  if (!data?.records?.length) {
    return <p className="text-muted-foreground py-8 text-center text-sm">No recent transactions</p>;
  }
  return (
    <div className="space-y-2.5">
      {data.records.slice(0, 6).map((tx, i) => (
        <TransactionCard key={tx.hash} transaction={tx} animationDelay={i * 40} />
      ))}
    </div>
  );
}

/* ─── Recent ledgers ─── */
function RecentLedgers() {
  const { data: latestLedger, isLoading } = useLatestLedger();
  if (isLoading || !latestLedger) {
    return (
      <div className="space-y-2.5">
        {Array.from({ length: 4 }).map((_, i) => (
          <LedgerCardSkeleton key={i} />
        ))}
      </div>
    );
  }
  return (
    <div className="space-y-2.5">
      <LedgerCard ledger={latestLedger} />
      <p className="text-muted-foreground pt-2 text-center text-xs">
        View all ledgers for complete history
      </p>
    </div>
  );
}

/* ─── Explore cards ─── */
const EXPLORE_ITEMS = [
  {
    href: "/transactions",
    icon: ArrowRightLeft,
    label: "Transactions",
    desc: "Live network activity",
    hue: "var(--chart-1)",
  },
  {
    href: "/ledgers",
    icon: Layers,
    label: "Ledgers",
    desc: "Browse ledger history",
    hue: "var(--chart-2)",
  },
  {
    href: "/accounts",
    icon: Users,
    label: "Accounts",
    desc: "Stellar accounts",
    hue: "var(--chart-3)",
  },
  { href: "/assets", icon: Coins, label: "Assets", desc: "Tokens & assets", hue: "var(--chart-4)" },
  {
    href: "/contracts",
    icon: FileCode,
    label: "Contracts",
    desc: "Soroban smart contracts",
    hue: "var(--chart-5)",
  },
  {
    href: "/analytics",
    icon: Activity,
    label: "Analytics",
    desc: "Network statistics",
    hue: "var(--chart-1)",
  },
];

/* ─── Home page ─── */
export default function HomePage() {
  const { isConnected, error: ledgerError } = useLedgerStream({ enabled: true });
  const t = useTranslations("home");

  const streamingStatus = isConnected ? "connected" : ledgerError ? "disconnected" : "connecting";

  return (
    <div className="space-y-0">
      {/* ── Hero ── */}
      <section className="relative -mx-4 -mt-4 overflow-hidden md:-mx-6 md:-mt-6 lg:-mx-8 lg:-mt-8">
        {/* Two-column layout: text left, globe right */}
        <div className="grid items-center gap-0 md:grid-cols-2">
          {/* Left: text + search */}
          <div className="relative z-10 flex flex-col justify-center px-4 pt-14 pb-8 md:px-6 md:py-16 lg:px-8">
            <div className="mb-5 flex items-center gap-3">
              <NetworkBadge network={useNetwork().network} />
              <LiveIndicator status={streamingStatus} />
            </div>

            <h1 className="mb-4 text-4xl leading-tight font-black tracking-tight md:text-5xl lg:text-6xl">
              {t("title")}
            </h1>

            <p className="text-muted-foreground mb-6 max-w-md text-base">{t("subtitle")}</p>

            <GlobalSearch className="w-full max-w-md" heroMode />
          </div>

          {/* Right: live globe */}
          <div className="relative flex items-center justify-center px-4 pt-2 pb-4 md:px-6 md:py-8">
            {/* Subtle glow behind globe */}
            <div
              className="pointer-events-none absolute inset-0 opacity-10 blur-3xl"
              style={{
                background:
                  "radial-gradient(ellipse at 60% 50%, var(--chart-1) 0%, transparent 65%)",
              }}
            />
            <div className="relative w-full max-w-[420px]">
              <NetworkGlobe />
              {/* "Live transactions" label */}
              <div className="border-border/40 bg-background/70 absolute bottom-4 left-1/2 flex -translate-x-1/2 items-center gap-2 rounded-full border px-3 py-1.5 backdrop-blur-sm">
                <span className="size-1.5 animate-pulse rounded-full bg-emerald-400" />
                <span className="text-muted-foreground text-[11px] font-medium">
                  Live network activity
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── Stats ribbon ── */}
      <LiveStatsRibbon />

      {/* ── Activity feed + ledgers ── */}
      <div className="mt-8 grid gap-6 md:grid-cols-5">
        {/* Transactions — wider */}
        <div className="md:col-span-3">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div
                className="flex size-7 items-center justify-center rounded-lg"
                style={{ background: "oklch(from var(--chart-1) l c h / 0.12)" }}
              >
                <ArrowRightLeft className="size-3.5" style={{ color: "var(--chart-1)" }} />
              </div>
              <h2 className="font-semibold">{t("recentTransactions")}</h2>
            </div>
            <Button variant="ghost" size="sm" asChild>
              <Link href="/transactions" className="gap-1 text-xs">
                View all <ArrowRight className="size-3" />
              </Link>
            </Button>
          </div>
          <Suspense
            fallback={
              <div className="space-y-2.5">
                {Array.from({ length: 5 }).map((_, i) => (
                  <TransactionCardSkeleton key={i} />
                ))}
              </div>
            }
          >
            <RecentTransactions />
          </Suspense>
        </div>

        {/* Ledgers — narrower */}
        <div className="md:col-span-2">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div
                className="flex size-7 items-center justify-center rounded-lg"
                style={{ background: "oklch(from var(--chart-2) l c h / 0.12)" }}
              >
                <Layers className="size-3.5" style={{ color: "var(--chart-2)" }} />
              </div>
              <h2 className="font-semibold">{t("latestLedger")}</h2>
            </div>
            <Button variant="ghost" size="sm" asChild>
              <Link href="/ledgers" className="gap-1 text-xs">
                View all <ArrowRight className="size-3" />
              </Link>
            </Button>
          </div>
          <Suspense
            fallback={
              <div className="space-y-2.5">
                {Array.from({ length: 4 }).map((_, i) => (
                  <LedgerCardSkeleton key={i} />
                ))}
              </div>
            }
          >
            <RecentLedgers />
          </Suspense>
        </div>
      </div>

      {/* ── Network charts ── */}
      <section className="mt-10">
        <div className="mb-5 flex items-center gap-2">
          <div
            className="flex size-7 items-center justify-center rounded-lg"
            style={{ background: "oklch(from var(--chart-3) l c h / 0.12)" }}
          >
            <Activity className="size-3.5" style={{ color: "var(--chart-3)" }} />
          </div>
          <h2 className="font-semibold">{t("networkActivity")}</h2>
        </div>
        <DashboardCharts />
      </section>

      {/* ── Explore grid ── */}
      <section className="mt-10">
        <h2 className="mb-5 font-semibold">{t("explore")}</h2>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-6">
          {EXPLORE_ITEMS.map((item, i) => {
            const Icon = item.icon;
            return (
              <Link key={item.href} href={item.href}>
                <div
                  className="border-border/40 hover:border-border group flex flex-col gap-3 rounded-xl border p-4 transition-all duration-200 hover:-translate-y-0.5"
                  style={{ animationDelay: `${i * 50}ms` }}
                >
                  <div
                    className="flex size-9 items-center justify-center rounded-lg transition-transform duration-200 group-hover:scale-110"
                    style={{ background: `oklch(from ${item.hue} l c h / 0.12)` }}
                  >
                    <Icon className="size-4" style={{ color: item.hue }} />
                  </div>
                  <div>
                    <p className="text-sm font-semibold">{item.label}</p>
                    <p className="text-muted-foreground text-xs">{item.desc}</p>
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      </section>
    </div>
  );
}
