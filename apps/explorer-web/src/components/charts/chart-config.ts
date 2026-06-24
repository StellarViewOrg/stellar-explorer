// Chart series colors. These reference the theme's `--chart-*` tokens so the
// data series adapt to light/dark mode and stay consistent with the header
// numbers (text-chart-1/2/3). Recharts accepts CSS `var()` in stroke/fill/stopColor.
export const chartColors = {
  primary: "var(--chart-1)", // Blue - for TPS chart and operations
  success: "var(--chart-2)", // Green - for transaction volume
  warning: "var(--chart-3)", // Cyan - for fees
  purple: "var(--chart-4)", // Purple - for contracts
  red: "var(--chart-5)", // Red - for failed transactions
  muted: "var(--muted-foreground)", // Muted elements
};

export const chartConfig = {
  height: 180,
  mobileHeight: 140,
  animationDuration: 300,
  tpsBufferSize: 30, // ~2.5 minutes at 5s intervals
  txAccumulationHours: 24,
};

// Common chart styling.
// Theme tokens are stored as raw color values (oklch), so reference them with
// `var(--token)` directly — wrapping them in `hsl(...)` produces an invalid color.
export const chartAxisStyle = {
  fontSize: 10,
  fill: "var(--muted-foreground)",
  fontFamily: "inherit",
};

export const chartGridStyle = {
  strokeDasharray: "3 3",
  stroke: "var(--border)",
  strokeOpacity: 0.5,
};

export const chartTooltipStyle = {
  backgroundColor: "var(--popover)",
  border: "1px solid var(--border)",
  borderRadius: "8px",
  padding: "8px 12px",
  boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
};
