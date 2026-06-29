import { notFound } from "next/navigation";
import type { ReactNode } from "react";
import { VALID_NETWORKS } from "@/lib/constants";

export function generateStaticParams() {
  return VALID_NETWORKS.map((network) => ({ network }));
}

export default async function NetworkLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ network: string }>;
}) {
  const { network } = await params;

  if (!(VALID_NETWORKS as readonly string[]).includes(network)) {
    notFound();
  }

  return children;
}
