"use client";

import type { ReactNode } from "react";
import { QueryProvider } from "./query-provider";
import { NetworkProvider } from "./network-provider";
import { ThemeProvider, useTheme } from "./theme-provider";
import { DeveloperModeProvider } from "./developer-mode-provider";
import { AnalyticsModeProvider } from "./analytics-mode-provider";
import { Toaster } from "sonner";

function ThemedToaster() {
  const { resolvedTheme } = useTheme();
  return (
    <Toaster
      position="bottom-right"
      theme={resolvedTheme}
      toastOptions={{
        style: {
          background: "var(--card)",
          border: "1px solid var(--border)",
          color: "var(--foreground)",
        },
      }}
    />
  );
}

export function Providers({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider defaultTheme="dark">
      <QueryProvider>
        <NetworkProvider>
          <DeveloperModeProvider>
            <AnalyticsModeProvider>
              {children}
              <ThemedToaster />
            </AnalyticsModeProvider>
          </DeveloperModeProvider>
        </NetworkProvider>
      </QueryProvider>
    </ThemeProvider>
  );
}

export { useNetwork } from "./network-provider";
export { useTheme } from "./theme-provider";
export { useDeveloperMode } from "./developer-mode-provider";
export { useAnalyticsMode } from "./analytics-mode-provider";
