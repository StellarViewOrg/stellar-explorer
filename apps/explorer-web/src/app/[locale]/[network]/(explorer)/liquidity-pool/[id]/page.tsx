import { Metadata } from "next";
import { LiquidityPoolContent } from "./liquidity-pool-content";

type Props = {
  params: Promise<{ id: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const shortId = `${id.slice(0, 8)}…${id.slice(-8)}`;

  return {
    title: `Liquidity Pool ${shortId}`,
    description: `View Stellar liquidity pool ${shortId}. Explore reserves, shares, and transaction history.`,
  };
}

export default async function LiquidityPoolPage({ params }: Props) {
  const { id } = await params;
  return <LiquidityPoolContent id={id} />;
}
