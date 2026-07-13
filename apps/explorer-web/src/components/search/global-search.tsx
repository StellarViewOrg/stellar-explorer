"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/i18n/navigation";
import { Search, ArrowRightLeft, Users, FileCode, Coins, Layers, Clock } from "lucide-react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { Button } from "@/components/ui/button";
import { detectEntityType, getEntityRoute, truncateHash } from "@/lib/utils";
import { cn } from "@/lib/utils";
import type { EntityType } from "@/types";
import { useTranslations } from "next-intl";

const entityIcons: Record<EntityType, typeof Search> = {
  transaction: ArrowRightLeft,
  account: Users,
  contract: FileCode,
  asset: Coins,
  ledger: Layers,
  unknown: Search,
};

const RECENT_SEARCHES_KEY = "stellar-explorer-recent-searches";
const MAX_RECENT = 5;

interface RecentSearch {
  query: string;
  type: EntityType;
  timestamp: number;
}

interface GlobalSearchProps {
  className?: string;
  heroMode?: boolean;
}

export function GlobalSearch({ className, heroMode = false }: GlobalSearchProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [recentSearches, setRecentSearches] = useState<RecentSearch[]>([]);
  const router = useRouter();
  const t = useTranslations("search");
  const tEntity = useTranslations("entityTypes");
  const tComponents = useTranslations("components.globalSearch");

  // Load recent searches from localStorage
  useEffect(() => {
    try {
      const stored = localStorage.getItem(RECENT_SEARCHES_KEY);
      if (stored) {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setRecentSearches(JSON.parse(stored));
      }
    } catch {
      // Ignore parse errors
    }
  }, []);

  // Keyboard shortcut to open search
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };

    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  const saveRecentSearch = useCallback((search: RecentSearch) => {
    setRecentSearches((prev) => {
      // Remove duplicates and add new search at the beginning
      const filtered = prev.filter((s) => s.query !== search.query);
      const updated = [search, ...filtered].slice(0, MAX_RECENT);
      localStorage.setItem(RECENT_SEARCHES_KEY, JSON.stringify(updated));
      return updated;
    });
  }, []);

  const clearRecentSearches = useCallback(() => {
    setRecentSearches([]);
    localStorage.removeItem(RECENT_SEARCHES_KEY);
  }, []);

  const handleSearch = useCallback(
    (searchQuery: string) => {
      if (!searchQuery.trim()) return;

      const type = detectEntityType(searchQuery);
      const route = getEntityRoute(type, searchQuery);

      // Save to recent searches
      saveRecentSearch({
        query: searchQuery,
        type,
        timestamp: Date.now(),
      });

      setOpen(false);
      setQuery("");

      if (route) {
        router.push(route);
      } else {
        // Go to search results page for unknown types
        router.push(`/search?q=${encodeURIComponent(searchQuery)}`);
      }
    },
    [router, saveRecentSearch]
  );

  const detectedType = query ? detectEntityType(query) : "unknown";
  const Icon = entityIcons[detectedType];

  return (
    <>
      {/* Search trigger - icon only on mobile, full bar on md+ */}
      <Button variant="ghost" size="icon" className="md:hidden" onClick={() => setOpen(true)}>
        <Search className="size-5" />
        <span className="sr-only">{tComponents("search")}</span>
      </Button>

      {heroMode ? (
        /* Hero mode — large full-width search trigger */
        <button
          className={cn(
            "border-border/60 bg-card hover:border-primary/50 text-muted-foreground flex w-full cursor-text items-center gap-3 rounded-xl border px-4 py-3.5 text-left text-sm shadow-lg transition-all duration-200 focus:outline-none",
            className
          )}
          onClick={() => setOpen(true)}
        >
          <Search className="size-5 shrink-0" />
          <span className="flex-1">{t("inputPlaceholder")}</span>
          <kbd className="bg-muted border-border hidden items-center gap-1 rounded border px-1.5 py-0.5 font-mono text-[11px] sm:flex">
            <span>{tComponents("keyboardShortcut")}</span>K
          </kbd>
        </button>
      ) : (
        <Button
          variant="outline"
          className={cn(
            "text-muted-foreground relative hidden justify-start md:flex",
            "h-9 w-64 px-3 lg:w-80",
            className
          )}
          onClick={() => setOpen(true)}
        >
          <Search className="mr-2 size-4" />
          <span className="flex-1 truncate text-left">{t("placeholder")}</span>
          <kbd className="bg-muted text-muted-foreground hidden h-5 items-center gap-1 rounded border px-1.5 font-mono text-[10px] font-medium select-none lg:inline-flex">
            <span className="text-xs">{tComponents("keyboardShortcut")}</span>K
          </kbd>
        </Button>
      )}

      {/* Search dialog */}
      <CommandDialog open={open} onOpenChange={setOpen} title={tComponents("commandPalette")}>
        <CommandInput
          placeholder={t("inputPlaceholder")}
          value={query}
          onValueChange={setQuery}
          onKeyDown={(e) => {
            if (e.key === "Enter" && query.trim()) {
              e.preventDefault();
              handleSearch(query);
            }
          }}
        />
        <CommandList>
          <CommandEmpty>
            <div className="py-6 text-center">
              <p className="text-muted-foreground text-sm">{t("noResults")}</p>
              <p className="text-muted-foreground mt-1 text-xs">{t("noResultsHint")}</p>
            </div>
          </CommandEmpty>

          {/* Current query suggestion */}
          {query && (
            <CommandGroup heading={tComponents("search")}>
              <CommandItem value={query} onSelect={() => handleSearch(query)} className="gap-3">
                <Icon className="size-4" />
                <div className="flex-1 overflow-hidden">
                  <p className="truncate font-mono text-sm">{truncateHash(query, 16, 16)}</p>
                  <p className="text-muted-foreground text-xs">
                    {t("searchAs", { type: tEntity(detectedType) })}
                  </p>
                </div>
              </CommandItem>
            </CommandGroup>
          )}

          {/* Recent searches */}
          {!query && recentSearches.length > 0 && (
            <>
              <CommandGroup
                heading={
                  <div className="flex items-center justify-between">
                    <span>{t("recent")}</span>
                    <Button
                      variant="ghost"
                      size="xs"
                      className="h-auto px-1.5 py-0.5 text-xs"
                      onClick={(e) => {
                        e.stopPropagation();
                        clearRecentSearches();
                      }}
                    >
                      {t("clear")}
                    </Button>
                  </div>
                }
              >
                {recentSearches.map((search) => {
                  const SearchIcon = entityIcons[search.type];
                  return (
                    <CommandItem
                      key={search.query}
                      onSelect={() => handleSearch(search.query)}
                      className="gap-3"
                    >
                      <Clock className="text-muted-foreground size-4" />
                      <div className="flex-1 overflow-hidden">
                        <p className="truncate font-mono text-sm">
                          {truncateHash(search.query, 12, 12)}
                        </p>
                        <p className="text-muted-foreground text-xs">{tEntity(search.type)}</p>
                      </div>
                      <SearchIcon className="text-muted-foreground size-4" />
                    </CommandItem>
                  );
                })}
              </CommandGroup>
            </>
          )}

          {/* Quick links when empty */}
          {!query && (
            <>
              <CommandSeparator />
              <CommandGroup heading={t("quickAccess")}>
                <CommandItem
                  onSelect={() => {
                    setOpen(false);
                    router.push("/transactions");
                  }}
                >
                  <ArrowRightLeft className="mr-3 size-4" />
                  {t("recentTransactions")}
                </CommandItem>
                <CommandItem
                  onSelect={() => {
                    setOpen(false);
                    router.push("/ledgers");
                  }}
                >
                  <Layers className="mr-3 size-4" />
                  {t("latestLedgers")}
                </CommandItem>
                <CommandItem
                  onSelect={() => {
                    setOpen(false);
                    router.push("/contracts");
                  }}
                >
                  <FileCode className="mr-3 size-4" />
                  {t("smartContracts")}
                </CommandItem>
              </CommandGroup>
            </>
          )}
        </CommandList>
      </CommandDialog>
    </>
  );
}
