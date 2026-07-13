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
  Menu,
  GraduationCap,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import Image from "next/image";
import { useState } from "react";
import { Separator } from "@/components/ui/separator";

export function MobileNav() {
  const pathname = usePathname();
  const t = useTranslations("navigation");
  const tMobile = useTranslations("mobileNav");
  const [open, setOpen] = useState(false);

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
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="md:hidden">
          <Menu className="size-5" />
          <span className="sr-only">{t("toggleMenu")}</span>
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-72 p-0">
        <SheetHeader className="border-border border-b p-4">
          <SheetTitle className="flex items-center gap-3">
            <Image
              src="/stellar-explorer.png"
              alt="Stellar Explorer"
              width={32}
              height={32}
              className="rounded-lg"
            />
            <span>{tMobile("stellarExplorer")}</span>
          </SheetTitle>
        </SheetHeader>

        <nav className="flex flex-col gap-1 p-4">
          {navItems.map((item) => {
            const isActive =
              pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setOpen(false)}
                className={cn(
                  "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors",
                  "hover:bg-accent hover:text-accent-foreground",
                  isActive ? "bg-accent text-accent-foreground" : "text-muted-foreground"
                )}
              >
                <item.icon className={cn("size-5", isActive && "text-primary")} />
                <span>{item.label}</span>
              </Link>
            );
          })}

          <Separator className="my-2" />

          {bottomItems.map((item) => {
            const isActive = pathname === item.href;

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setOpen(false)}
                className={cn(
                  "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors",
                  "hover:bg-accent hover:text-accent-foreground",
                  isActive ? "bg-accent text-accent-foreground" : "text-muted-foreground"
                )}
              >
                <item.icon className={cn("size-5", isActive && "text-primary")} />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>
      </SheetContent>
    </Sheet>
  );
}
