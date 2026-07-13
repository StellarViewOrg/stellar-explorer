"use client";

import { createNavigation } from "next-intl/navigation";
import { locales, defaultLocale } from "./config";
import { useParams } from "next/navigation";
import { forwardRef, type ComponentProps } from "react";

const VALID_NETWORKS = ["mainnet", "testnet", "futurenet"] as const;

const nav = createNavigation({ locales, defaultLocale });

// Re-export unchanged utilities
export const { redirect, getPathname } = nav;

/**
 * Read the current network from URL params.
 * Falls back to "mainnet" if not present or invalid (e.g. in not-found pages).
 */
export function useNetworkPrefix(): string {
  const params = useParams();
  const network = params?.network as string | undefined;
  if (network && (VALID_NETWORKS as readonly string[]).includes(network)) {
    return `/${network}`;
  }
  return "/mainnet";
}

/**
 * Link component that auto-prefixes href with the current network segment.
 * Drop-in replacement for next-intl's Link --- existing href values like
 * "/tx/abc" become "/public/tx/abc" automatically.
 */
export const Link = forwardRef<HTMLAnchorElement, ComponentProps<typeof nav.Link>>(
  function NetworkLink({ href, ...props }, ref) {
    const prefix = useNetworkPrefix();

    const prefixedHref = typeof href === "string" ? `${prefix}${href}` : href;

    return <nav.Link ref={ref} href={prefixedHref} suppressHydrationWarning {...props} />;
  }
);

/**
 * useRouter that auto-prefixes push/replace paths with the current network.
 */
export function useRouter() {
  const router = nav.useRouter();
  const prefix = useNetworkPrefix();

  return {
    ...router,
    push(path: string, options?: Parameters<typeof router.push>[1]) {
      router.push(`${prefix}${path}`, options);
    },
    replace(path: string, options?: Parameters<typeof router.replace>[1]) {
      router.replace(`${prefix}${path}`, options);
    },
  };
}

/**
 * usePathname that strips the /{network} prefix so consumers see paths
 * like "/tx/abc" regardless of network.
 */
export function usePathname(): string {
  const pathname = nav.usePathname();
  for (const net of VALID_NETWORKS) {
    if (pathname === `/${net}`) return "/";
    if (pathname.startsWith(`/${net}/`)) return pathname.slice(`/${net}`.length);
  }
  return pathname;
}
