"use client";

import { useState, useMemo } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  glossaryTermsMeta,
  glossaryCategoryMeta,
  getTermIdsByCategory,
  getTermIdsByLevel,
  getAllTermIds,
  type GlossaryEntry,
  type GlossaryLevel,
} from "@/lib/glossary";
import {
  BookOpen,
  Search,
  ExternalLink,
  Globe,
  User,
  ArrowLeftRight,
  Layers,
  Coins,
  TrendingUp,
  Code,
  Puzzle,
  Terminal,
  ChevronDown,
  ChevronUp,
  GraduationCap,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Link } from "@/i18n/navigation";

const iconMap: Record<string, React.ComponentType<{ className?: string }>> = {
  BookOpen,
  Globe,
  User,
  ArrowLeftRight,
  Layers,
  Coins,
  TrendingUp,
  Code,
  Puzzle,
  Terminal,
};

const levelColors: Record<GlossaryLevel, string> = {
  beginner: "bg-green-500/10 text-green-500 border-green-500/20",
  intermediate: "bg-amber-500/10 text-amber-500 border-amber-500/20",
  advanced: "bg-purple-500/10 text-purple-500 border-purple-500/20",
};

// Hook to get translated glossary entries
function useTranslatedGlossary() {
  const tGlossary = useTranslations("glossary");

  const getEntry = (id: string): GlossaryEntry | null => {
    const meta = glossaryTermsMeta[id];
    if (!meta) return null;

    return {
      ...meta,
      term: tGlossary(`terms.${id}.term`),
      shortDefinition: tGlossary(`terms.${id}.short`),
      fullDefinition: tGlossary(`terms.${id}.full`),
    };
  };

  const getAllEntries = (): GlossaryEntry[] => {
    return getAllTermIds()
      .map(getEntry)
      .filter((entry): entry is GlossaryEntry => entry !== null);
  };

  const getEntriesByCategory = (category: string): GlossaryEntry[] => {
    return getTermIdsByCategory(category)
      .map(getEntry)
      .filter((entry): entry is GlossaryEntry => entry !== null);
  };

  const getEntriesByLevel = (level: GlossaryLevel): GlossaryEntry[] => {
    return getTermIdsByLevel(level)
      .map(getEntry)
      .filter((entry): entry is GlossaryEntry => entry !== null);
  };

  const searchEntries = (query: string): GlossaryEntry[] => {
    const lowerQuery = query.toLowerCase();
    return getAllEntries().filter(
      (entry) =>
        entry.term.toLowerCase().includes(lowerQuery) ||
        entry.shortDefinition.toLowerCase().includes(lowerQuery)
    );
  };

  const getCategoryLabel = (categoryId: string): string => {
    return tGlossary(`categories.${categoryId}`);
  };

  return {
    getEntry,
    getAllEntries,
    getEntriesByCategory,
    getEntriesByLevel,
    searchEntries,
    getCategoryLabel,
  };
}

function GlossaryCard({ entry }: { entry: GlossaryEntry }) {
  const [isExpanded, setIsExpanded] = useState(false);
  const t = useTranslations("learn");
  const { getEntry } = useTranslatedGlossary();

  return (
    <Card id={entry.id} className="group hover:border-primary/20 scroll-mt-20 transition-colors">
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-base font-semibold">{entry.term}</CardTitle>
          <Badge variant="outline" className={cn("shrink-0 text-[10px]", levelColors[entry.level])}>
            {t(entry.level)}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-muted-foreground text-sm">
          {isExpanded ? entry.fullDefinition : entry.shortDefinition}
        </p>

        <div className="flex items-center justify-between pt-1">
          <Button
            variant="ghost"
            size="sm"
            className="text-muted-foreground hover:text-foreground h-7 text-xs"
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {isExpanded ? (
              <>
                <ChevronUp className="mr-1 size-3" />
                {t("showLess")}
              </>
            ) : (
              <>
                <ChevronDown className="mr-1 size-3" />
                {t("readMore")}
              </>
            )}
          </Button>

          {entry.learnMoreUrl && (
            <Button variant="ghost" size="sm" className="h-7 text-xs" asChild>
              <a href={entry.learnMoreUrl} target="_blank" rel="noopener noreferrer">
                <ExternalLink className="mr-1 size-3" />
                {t("officialDocs")}
              </a>
            </Button>
          )}
        </div>

        {isExpanded && entry.relatedTerms && entry.relatedTerms.length > 0 && (
          <>
            <Separator />
            <div className="flex flex-wrap gap-1.5">
              <span className="text-muted-foreground mr-1 text-xs">{t("related")}:</span>
              {entry.relatedTerms.map((termId) => {
                const relatedTerm = getEntry(termId);
                if (!relatedTerm) return null;
                return (
                  <Link key={termId} href={`#${termId}`}>
                    <Badge
                      variant="secondary"
                      className="hover:bg-secondary/80 cursor-pointer text-[10px]"
                    >
                      {relatedTerm.term}
                    </Badge>
                  </Link>
                );
              })}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}

function CategorySection({ categoryId }: { categoryId: string }) {
  const { getEntriesByCategory, getCategoryLabel } = useTranslatedGlossary();
  const t = useTranslations("learn");
  const category = glossaryCategoryMeta.find((c) => c.id === categoryId);
  const terms = getEntriesByCategory(categoryId);
  const Icon = category ? iconMap[category.icon] : BookOpen;
  const categoryLabel = getCategoryLabel(categoryId);

  if (!category || terms.length === 0) return null;

  return (
    <section className="space-y-4">
      <div className="flex items-center gap-2">
        {Icon && <Icon className="text-primary size-5" />}
        <h2 className="text-lg font-semibold">{categoryLabel}</h2>
        <Badge variant="secondary" className="text-xs">
          {terms.length} {t("totalTerms")}
        </Badge>
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        {terms.map((term) => (
          <GlossaryCard key={term.id} entry={term} />
        ))}
      </div>
    </section>
  );
}

function LevelSection({ level }: { level: GlossaryLevel }) {
  const { getEntriesByLevel } = useTranslatedGlossary();
  const terms = getEntriesByLevel(level);

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {terms.map((term) => (
        <GlossaryCard key={term.id} entry={term} />
      ))}
    </div>
  );
}

export default function LearnPage() {
  const t = useTranslations("learn");
  const { getAllEntries, getEntriesByLevel, searchEntries } = useTranslatedGlossary();
  const [searchQuery, setSearchQuery] = useState("");
  const [viewMode, setViewMode] = useState<"category" | "level">("category");

  const searchResults = useMemo(() => {
    if (!searchQuery.trim()) return null;
    return searchEntries(searchQuery);
  }, [searchQuery, searchEntries]);

  const allTerms = getAllEntries();
  const beginnerCount = getEntriesByLevel("beginner").length;
  const intermediateCount = getEntriesByLevel("intermediate").length;
  const advancedCount = getEntriesByLevel("advanced").length;

  return (
    <div className="space-y-8">
      {/* Header */}
      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <div className="bg-primary/10 rounded-xl p-2.5">
            <GraduationCap className="text-primary size-6" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">{t("title")}</h1>
            <p className="text-muted-foreground">{t("subtitle")}</p>
          </div>
        </div>

        {/* Stats */}
        <div className="flex flex-wrap gap-3">
          <Badge variant="outline" className="gap-1.5">
            <BookOpen className="size-3" />
            {allTerms.length} {t("totalTerms")}
          </Badge>
          <Badge variant="outline" className={cn("gap-1.5", levelColors.beginner)}>
            {beginnerCount} {t("beginnerTerms")}
          </Badge>
          <Badge variant="outline" className={cn("gap-1.5", levelColors.intermediate)}>
            {intermediateCount} {t("intermediateTerms")}
          </Badge>
          <Badge variant="outline" className={cn("gap-1.5", levelColors.advanced)}>
            {advancedCount} {t("advancedTerms")}
          </Badge>
        </div>
      </section>

      {/* Search */}
      <Card>
        <CardContent className="p-4">
          <div className="relative">
            <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
            <Input
              type="text"
              placeholder={t("searchPlaceholder")}
              aria-label={t("searchPlaceholder")}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        </CardContent>
      </Card>

      {/* Search Results */}
      {searchResults !== null ? (
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">
              {t("searchResults")} ({searchResults.length})
            </h2>
            <Button variant="ghost" size="sm" onClick={() => setSearchQuery("")}>
              {t("clearSearch")}
            </Button>
          </div>
          {searchResults.length > 0 ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {searchResults.map((term) => (
                <GlossaryCard key={term.id} entry={term} />
              ))}
            </div>
          ) : (
            <Card>
              <CardContent className="text-muted-foreground py-8 text-center">
                {t("noResults")}
              </CardContent>
            </Card>
          )}
        </section>
      ) : (
        <>
          {/* View Toggle */}
          <Tabs value={viewMode} onValueChange={(v) => setViewMode(v as "category" | "level")}>
            <TabsList>
              <TabsTrigger value="category">{t("byCategory")}</TabsTrigger>
              <TabsTrigger value="level">{t("byLevel")}</TabsTrigger>
            </TabsList>

            <TabsContent value="category" className="mt-6 space-y-8">
              {glossaryCategoryMeta.map((category) => (
                <CategorySection key={category.id} categoryId={category.id} />
              ))}
            </TabsContent>

            <TabsContent value="level" className="mt-6 space-y-8">
              <section className="space-y-4">
                <div className="flex items-center gap-2">
                  <Badge className={cn("gap-1.5", levelColors.beginner)}>{t("beginner")}</Badge>
                  <span className="text-muted-foreground text-sm">{t("beginnerDescription")}</span>
                </div>
                <LevelSection level="beginner" />
              </section>

              <Separator />

              <section className="space-y-4">
                <div className="flex items-center gap-2">
                  <Badge className={cn("gap-1.5", levelColors.intermediate)}>
                    {t("intermediate")}
                  </Badge>
                  <span className="text-muted-foreground text-sm">
                    {t("intermediateDescription")}
                  </span>
                </div>
                <LevelSection level="intermediate" />
              </section>

              <Separator />

              <section className="space-y-4">
                <div className="flex items-center gap-2">
                  <Badge className={cn("gap-1.5", levelColors.advanced)}>{t("advanced")}</Badge>
                  <span className="text-muted-foreground text-sm">{t("advancedDescription")}</span>
                </div>
                <LevelSection level="advanced" />
              </section>
            </TabsContent>
          </Tabs>
        </>
      )}

      {/* Learning Resources */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("moreResources")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <a
              href="https://stellar.org/learn"
              target="_blank"
              rel="noopener noreferrer"
              className="group hover:bg-muted/50 flex items-center gap-3 rounded-lg border p-3 transition-colors"
            >
              <div className="bg-primary/10 rounded-lg p-2">
                <BookOpen className="text-primary size-4" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-sm font-medium">{t("stellarLearn")}</div>
                <div className="text-muted-foreground truncate text-xs">stellar.org/learn</div>
              </div>
              <ExternalLink className="text-muted-foreground group-hover:text-foreground size-4" />
            </a>

            <a
              href="https://developers.stellar.org"
              target="_blank"
              rel="noopener noreferrer"
              className="group hover:bg-muted/50 flex items-center gap-3 rounded-lg border p-3 transition-colors"
            >
              <div className="bg-chart-1/10 rounded-lg p-2">
                <Code className="text-chart-1 size-4" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-sm font-medium">{t("developerDocs")}</div>
                <div className="text-muted-foreground truncate text-xs">developers.stellar.org</div>
              </div>
              <ExternalLink className="text-muted-foreground group-hover:text-foreground size-4" />
            </a>

            <a
              href="https://soroban.stellar.org"
              target="_blank"
              rel="noopener noreferrer"
              className="group hover:bg-muted/50 flex items-center gap-3 rounded-lg border p-3 transition-colors"
            >
              <div className="bg-chart-2/10 rounded-lg p-2">
                <Terminal className="text-chart-2 size-4" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-sm font-medium">{t("sorobanDocs")}</div>
                <div className="text-muted-foreground truncate text-xs">soroban.stellar.org</div>
              </div>
              <ExternalLink className="text-muted-foreground group-hover:text-foreground size-4" />
            </a>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
