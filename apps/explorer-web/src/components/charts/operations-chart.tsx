"use client";

import { BarChart3 } from "lucide-react";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";
import { ChartWrapper } from "./chart-wrapper";
import { useOpsPerLedgerChartData } from "@/lib/hooks/use-chart-data";
import { useTranslations } from "next-intl";
import { chartColors, chartConfig, chartAxisStyle, chartGridStyle } from "./chart-config";

interface TooltipProps {
  active?: boolean;
  payload?: Array<{ value: number; payload: { ledger: number } }>;
  ledgerLabel?: string;
  operationsLabel?: string;
}

function CustomTooltip({
  active,
  payload,
  ledgerLabel = "Ledger",
  operationsLabel = "Operations",
}: TooltipProps) {
  if (!active || !payload?.length) return null;

  const { ledger } = payload[0].payload;

  return (
    <div className="bg-popover border-border rounded-lg border px-3 py-2 shadow-lg">
      <p className="text-foreground text-sm font-medium">
        {ledgerLabel} #{Number(ledger).toLocaleString()}
      </p>
      <p className="text-muted-foreground text-xs">
        {operationsLabel}: {Number(payload[0].value).toLocaleString()}
      </p>
    </div>
  );
}

export default function OperationsChart() {
  const { data, avgOps, isLoading } = useOpsPerLedgerChartData();
  const t = useTranslations("charts");

  const hasData = data.length > 1;

  return (
    <ChartWrapper
      title={t("opsPerLedger")}
      subtitle={hasData ? t("avgOps", { count: avgOps }) : t("collectingData")}
      icon={BarChart3}
      loading={isLoading}
    >
      {!hasData ? (
        <div className="flex h-[140px] items-center justify-center">
          <p className="text-muted-foreground text-sm">{t("chartWillPopulate")}</p>
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={chartConfig.mobileHeight}>
          <BarChart data={data} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
            <CartesianGrid {...chartGridStyle} vertical={false} />
            <XAxis
              dataKey="ledger"
              {...chartAxisStyle}
              tickFormatter={(v: number) => `#${(v % 100000).toString()}`}
            />
            <YAxis {...chartAxisStyle} />
            <Tooltip
              cursor={{ fill: "var(--muted)", opacity: 0.4 }}
              content={
                <CustomTooltip
                  ledgerLabel={t("ledgerLabel")}
                  operationsLabel={t("operationsLabel")}
                />
              }
            />
            <Bar dataKey="ops" fill={chartColors.primary} radius={[4, 4, 0, 0]} opacity={0.85} />
          </BarChart>
        </ResponsiveContainer>
      )}
    </ChartWrapper>
  );
}
