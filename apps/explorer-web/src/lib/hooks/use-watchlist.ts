"use client";

import { useState, useCallback, useEffect } from "react";
import type { WatchlistItem } from "@/types";
import { toast } from "sonner";

const STORAGE_KEY = "stellar-explorer-watchlist";

export function useWatchlist() {
  const [items, setItems] = useState<WatchlistItem[]>([]);
  const [isHydrated, setIsHydrated] = useState(false);

  // Load from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setItems(JSON.parse(stored));
      }
    } catch {
      // Ignore parse errors
    }
    setIsHydrated(true);
  }, []);

  // Save to localStorage whenever items change (debounced to avoid rapid writes)
  useEffect(() => {
    if (!isHydrated) return;
    const timer = setTimeout(() => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
    }, 500);
    return () => clearTimeout(timer);
  }, [items, isHydrated]);

  const add = useCallback(
    (item: Omit<WatchlistItem, "addedAt">) => {
      const exists = items.some((i) => i.type === item.type && i.id === item.id);
      if (exists) {
        toast.error("Already in watchlist");
        return false;
      }

      setItems((prev) => [...prev, { ...item, addedAt: Date.now() }]);
      toast.success("Added to watchlist");
      return true;
    },
    [items]
  );

  const remove = useCallback((id: string) => {
    setItems((prev) => prev.filter((item) => item.id !== id));
    toast.success("Removed from watchlist");
  }, []);

  const update = useCallback((id: string, updates: Partial<WatchlistItem>) => {
    setItems((prev) => prev.map((item) => (item.id === id ? { ...item, ...updates } : item)));
  }, []);

  const has = useCallback((id: string) => items.some((item) => item.id === id), [items]);

  const clear = useCallback(() => {
    setItems([]);
    toast.success("Watchlist cleared");
  }, []);

  return {
    items,
    add,
    remove,
    update,
    has,
    clear,
    isHydrated,
  };
}
