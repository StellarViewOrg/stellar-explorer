"use client";

import { useEffect } from "react";
import { AlertTriangle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function ExplorerError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center gap-4 text-center">
      <div className="bg-destructive/15 rounded-full p-4">
        <AlertTriangle className="text-destructive size-8" />
      </div>
      <div>
        <h2 className="text-foreground mb-1 text-lg font-semibold">Something went wrong</h2>
        <p className="text-muted-foreground max-w-sm text-sm">
          {error.message || "An unexpected error occurred while loading this page."}
        </p>
      </div>
      <Button variant="outline" onClick={reset}>
        <RefreshCw className="mr-2 size-4" />
        Try again
      </Button>
    </div>
  );
}
