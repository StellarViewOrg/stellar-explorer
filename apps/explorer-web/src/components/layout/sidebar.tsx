"use client";

import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/navigation";
import {
  Home,
  Layers,
  ArrowRightLeft,
  Users,
  Coins,
  FileCode,
  BarChart3,
  Star,
  ChevronLeft,
  ChevronRight,
  GraduationCap,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import Image from "next/image";
import { useState } from "react";

export function Sidebar() {
  const pathname = usePathname();
  const t = useTranslations("navigation");
  const [collapsed, setCollapsed] = useState(false);

  const navItems = [
    { href: "/", icon: Home, label: t("overview") },
    { href: "/ledgers", icon: Layers, label: t("ledgers") },
    { href: "/transactions", icon: ArrowRightLeft, label: t("transactions") },
    { href: "/accounts", icon: Users, label: t("accounts") },
    { href: "/assets", icon: Coins, label: t("assets") },
    { href: "/contracts", icon: FileCode, label: t("contracts") },
    { href: "/analytics", icon: BarChart3, label: t("analytics") },
  ];

  const bottomItems = [
    { href: "/learn", icon: GraduationCap, label: t("learn") },
    { href: "/watchlist", icon: Star, label: t("watchlist") },
  ];

  return (
    <aside
      className={cn(
        "border-border bg-sidebar/50 flex h-screen flex-col border-r backdrop-blur-sm transition-all duration-300",
        collapsed ? "w-16" : "w-60"
      )}
    >
      {/* Logo */}
      <div className="border-border/50 flex h-16 items-center border-b px-4">
        <Link href="/" className="group flex items-center gap-3">
          <Image
            src="/stellar-explorer.png"
            alt="StellarView Explorer"
            width={36}
            height={36}
            className="shrink-0 rounded-lg transition-transform duration-300 group-hover:scale-110"
          />
          {!collapsed && <span className="text-foreground text-lg font-bold">Stellar</span>}
        </Link>
      </div>

      {/* Navigation */}
      <ScrollArea className="flex-1 py-4">
        <TooltipProvider delayDuration={0}>
          <nav className="flex flex-col gap-1 px-3">
            {navItems.map((item) => {
              const isActive =
                pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));

              const link = (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-200",
                    "hover:bg-white/5 hover:pl-4",
                    isActive
                      ? "from-primary/15 text-foreground bg-gradient-to-r to-transparent"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  {/* Active indicator bar */}
                  {isActive && (
                    <div className="from-primary to-primary/50 absolute top-1/2 left-0 h-6 w-0.5 -translate-y-1/2 rounded-full bg-gradient-to-b" />
                  )}
                  <item.icon
                    className={cn(
                      "size-[18px] shrink-0 transition-colors duration-200",
                      isActive
                        ? "text-primary"
                        : "text-muted-foreground group-hover:text-foreground"
                    )}
                  />
                  {!collapsed && (
                    <span className="transition-colors duration-200">{item.label}</span>
                  )}
                </Link>
              );

              if (collapsed) {
                return (
                  <Tooltip key={item.href}>
                    <TooltipTrigger asChild>{link}</TooltipTrigger>
                    <TooltipContent side="right" className="font-medium">
                      {item.label}
                    </TooltipContent>
                  </Tooltip>
                );
              }

              return link;
            })}
          </nav>
        </TooltipProvider>
      </ScrollArea>

      {/* Bottom section */}
      <div className="border-border/50 border-t p-3">
        <TooltipProvider delayDuration={0}>
          <nav className="flex flex-col gap-1">
            {bottomItems.map((item) => {
              const isActive = pathname === item.href;

              const link = (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-200",
                    "hover:bg-white/5 hover:pl-4",
                    isActive
                      ? "from-warning/15 text-foreground bg-gradient-to-r to-transparent"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  {isActive && (
                    <div className="from-warning to-warning/50 absolute top-1/2 left-0 h-6 w-0.5 -translate-y-1/2 rounded-full bg-gradient-to-b" />
                  )}
                  <item.icon
                    className={cn(
                      "size-[18px] shrink-0 transition-colors duration-200",
                      isActive
                        ? "text-warning"
                        : "text-muted-foreground group-hover:text-foreground"
                    )}
                  />
                  {!collapsed && <span>{item.label}</span>}
                </Link>
              );

              if (collapsed) {
                return (
                  <Tooltip key={item.href}>
                    <TooltipTrigger asChild>{link}</TooltipTrigger>
                    <TooltipContent side="right" className="font-medium">
                      {item.label}
                    </TooltipContent>
                  </Tooltip>
                );
              }

              return link;
            })}
          </nav>
        </TooltipProvider>

        <Separator className="my-3 opacity-50" />

        {/* Collapse toggle */}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setCollapsed(!collapsed)}
          className={cn(
            "w-full justify-center transition-all duration-200 hover:bg-white/5",
            !collapsed && "justify-start px-3"
          )}
        >
          {collapsed ? (
            <ChevronRight className="size-4" />
          ) : (
            <>
              <ChevronLeft className="size-4" />
              <span className="text-muted-foreground ml-2">{t("collapse")}</span>
            </>
          )}
        </Button>
      </div>
    </aside>
  );
}
