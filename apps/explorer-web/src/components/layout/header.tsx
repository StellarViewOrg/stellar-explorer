"use client";

import { Link } from "@/i18n/navigation";
import Image from "next/image";
import { RefreshCw, Moon, Sun } from "lucide-react";
import { Button } from "@/components/ui/button";
import { NetworkSelector } from "./network-selector";
import { MobileNav } from "./mobile-nav";
import { GlobalSearch } from "@/components/search/global-search";
import { LocaleSwitcher } from "@/components/common/locale-switcher";
import { ViewModesToggle } from "@/components/common/view-modes-toggle";
import { useTheme } from "@/lib/providers";
import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useTranslations } from "next-intl";

export function Header() {
  const { setTheme, resolvedTheme } = useTheme();
  const queryClient = useQueryClient();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const t = useTranslations("header");

  const handleRefresh = async () => {
    setIsRefreshing(true);
    await queryClient.invalidateQueries();
    setTimeout(() => setIsRefreshing(false), 1000);
  };

  const toggleTheme = () => {
    setTheme(resolvedTheme === "dark" ? "light" : "dark");
  };

  return (
    <header className="glass-effect border-border/40 fixed top-0 right-0 left-0 z-50 h-16 border-b md:left-[72px]">
      <div className="flex h-full items-center gap-4 px-4 md:px-6 lg:px-8">
        {/* Mobile menu */}
        <MobileNav />

        {/* Logo (mobile only) */}
        <Link href="/" className="group flex items-center gap-2.5 font-semibold md:hidden">
          <Image
            src="/stellar-explorer.png"
            alt="StellarView Explorer"
            width={32}
            height={32}
            className="rounded-lg"
          />
        </Link>

        {/* Spacer for mobile to push actions to right */}
        <div className="flex-1 md:hidden" />

        {/* Global Search */}
        <GlobalSearch className="md:flex-1" />

        {/* Right side actions */}
        <div className="flex shrink-0 items-center gap-1">
          {/* Network Selector */}
          <NetworkSelector />

          {/* View Modes (Dev + Analytics) */}
          <ViewModesToggle />

          {/* Refresh (hidden on mobile) */}
          <Button
            variant="ghost"
            size="icon"
            onClick={handleRefresh}
            disabled={isRefreshing}
            className="hidden transition-colors hover:bg-white/10 md:inline-flex"
          >
            <RefreshCw className={cn("size-4", isRefreshing && "animate-spin")} />
            <span className="sr-only">{t("refreshData")}</span>
          </Button>

          {/* Theme toggle */}
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleTheme}
            className="transition-colors hover:bg-white/10"
          >
            <span className="relative size-4">
              <Sun
                className={cn(
                  "absolute inset-0 size-4 transition-all duration-300",
                  resolvedTheme === "dark"
                    ? "scale-100 rotate-0 opacity-100"
                    : "scale-75 rotate-90 opacity-0"
                )}
              />
              <Moon
                className={cn(
                  "absolute inset-0 size-4 transition-all duration-300",
                  resolvedTheme !== "dark"
                    ? "scale-100 rotate-0 opacity-100"
                    : "scale-75 -rotate-90 opacity-0"
                )}
              />
            </span>
            <span className="sr-only">{t("toggleTheme")}</span>
          </Button>

          {/* Language switcher */}
          <LocaleSwitcher />
        </div>
      </div>
    </header>
  );
}
