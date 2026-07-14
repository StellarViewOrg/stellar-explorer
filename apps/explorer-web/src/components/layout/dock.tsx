"use client";

import { useRef } from "react";
import { motion, useMotionValue, useSpring, useTransform } from "motion/react";
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
  GraduationCap,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import Image from "next/image";

const ICON_SIZE = 36;
const ICON_SIZE_MAX = 56;
const MAGNIFICATION_DISTANCE = 120;

interface DockItemProps {
  href: string;
  icon: React.ElementType;
  label: string;
  isActive: boolean;
  mouseY: ReturnType<typeof useMotionValue<number>>;
  isSecondary?: boolean;
}

function DockItem({ href, icon: Icon, label, isActive, mouseY, isSecondary }: DockItemProps) {
  const ref = useRef<HTMLDivElement>(null);

  const distance = useTransform(mouseY, (val) => {
    const bounds = ref.current?.getBoundingClientRect() ?? { y: 0, height: 0 };
    return val - bounds.y - bounds.height / 2;
  });

  const sizeTransform = useTransform(
    distance,
    [-MAGNIFICATION_DISTANCE, 0, MAGNIFICATION_DISTANCE],
    [ICON_SIZE, ICON_SIZE_MAX, ICON_SIZE]
  );

  const size = useSpring(sizeTransform, { mass: 0.1, stiffness: 180, damping: 18 });

  const activeColor = isSecondary ? "var(--chart-3)" : "var(--primary)";
  const activeBg = isSecondary
    ? "oklch(from var(--chart-3) l c h / 0.12)"
    : "oklch(from var(--primary) l c h / 0.12)";

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link
          href={href}
          className="focus-visible:ring-ring flex items-center justify-center rounded-xl outline-none focus-visible:ring-2"
        >
          <motion.div
            ref={ref}
            style={{ width: size, height: size }}
            className={cn(
              "relative flex items-center justify-center rounded-xl transition-colors duration-150",
              isActive ? "" : "hover:bg-muted/50"
            )}
          >
            {/* Active background */}
            {isActive && (
              <div className="absolute inset-0 rounded-xl" style={{ background: activeBg }} />
            )}
            {/* Active indicator dot */}
            {isActive && (
              <span
                className="absolute top-1/2 -right-0.5 size-1.5 -translate-y-1/2 rounded-full"
                style={{ background: activeColor }}
              />
            )}
            <Icon
              className="relative transition-colors duration-150"
              style={{
                width: "44%",
                height: "44%",
                color: isActive ? activeColor : "var(--muted-foreground)",
              }}
            />
          </motion.div>
        </Link>
      </TooltipTrigger>
      <TooltipContent side="right" sideOffset={8} className="font-medium">
        {label}
      </TooltipContent>
    </Tooltip>
  );
}

export function Dock() {
  const pathname = usePathname();
  const t = useTranslations("navigation");
  const mouseY = useMotionValue(Infinity);

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
    { href: "/learn", icon: GraduationCap, label: t("learn"), isSecondary: true },
    { href: "/watchlist", icon: Star, label: t("watchlist"), isSecondary: true },
  ];

  const isActive = (href: string) =>
    href === "/" ? pathname === "/" : pathname === href || pathname.startsWith(href + "/");

  return (
    <TooltipProvider delayDuration={400}>
      <aside
        className="border-border/30 bg-sidebar/60 fixed top-0 left-0 z-40 hidden h-screen w-[72px] flex-col items-center border-r py-4 backdrop-blur-md md:flex"
        onMouseMove={(e) => mouseY.set(e.clientY)}
        onMouseLeave={() => mouseY.set(Infinity)}
      >
        {/* Logo */}
        <Link href="/" className="mb-6 flex items-center justify-center">
          <div className="size-9 overflow-hidden rounded-xl transition-transform duration-200 hover:scale-105">
            <Image src="/stellar-explorer.png" alt="StellarView Explorer" width={36} height={36} />
          </div>
        </Link>

        {/* Main nav */}
        <nav className="flex flex-1 flex-col items-center gap-1.5">
          {navItems.map((item) => (
            <DockItem
              key={item.href}
              href={item.href}
              icon={item.icon}
              label={item.label}
              isActive={isActive(item.href)}
              mouseY={mouseY}
            />
          ))}
        </nav>

        {/* Separator */}
        <div className="bg-border/50 my-3 h-px w-8 shrink-0" />

        {/* Bottom nav */}
        <nav className="flex flex-col items-center gap-1.5 pb-2">
          {bottomItems.map((item) => (
            <DockItem
              key={item.href}
              href={item.href}
              icon={item.icon}
              label={item.label}
              isActive={isActive(item.href)}
              mouseY={mouseY}
              isSecondary={item.isSecondary}
            />
          ))}
        </nav>
      </aside>
    </TooltipProvider>
  );
}
