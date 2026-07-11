"use client";

import { useState } from "react";
import { Bot, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { generateAgentContext } from "@/lib/utils/agent-context";
import { useNetwork } from "@/lib/providers";
import { cn } from "@/lib/utils";

type EntityType = "transaction" | "account" | "contract" | "asset" | "ledger";

interface CopyContextButtonProps {
  type: EntityType;
  data: Record<string, unknown>;
  className?: string;
}

export function CopyContextButton({ type, data, className }: CopyContextButtonProps) {
  const { network } = useNetwork();
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const baseUrl =
      process.env.NEXT_PUBLIC_APP_URL ??
      (typeof window !== "undefined" ? window.location.origin : "");
    const context = generateAgentContext({ type, network, data, baseUrl });
    await navigator.clipboard.writeText(context);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCopy}
            className={cn(
              "text-muted-foreground hover:text-foreground gap-1.5 transition-colors",
              copied && "text-success hover:text-success",
              className
            )}
          >
            {copied ? <Check className="size-3.5" /> : <Bot className="size-3.5" />}
            <span className="text-xs">{copied ? "Copied!" : "Copy for AI"}</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>Copy structured context for AI assistants</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
