import { Metadata } from "next";
import { notFound } from "next/navigation";
import { LedgerContent } from "./ledger-content";

type Props = {
  params: Promise<{ sequence: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { sequence } = await params;
  const sequenceNum = parseInt(sequence, 10);

  if (isNaN(sequenceNum) || sequenceNum <= 0) {
    return {
      title: "Ledger Not Found",
    };
  }

  const formattedSequence = sequenceNum.toLocaleString("en-US");

  return {
    title: `Ledger #${formattedSequence}`,
    description: `View details of Stellar ledger #${formattedSequence}. Explore transactions, operations, and protocol information.`,
    openGraph: {
      title: `Stellar Ledger #${formattedSequence}`,
      description: `View ledger details on StellarView Explorer`,
      type: "website",
    },
    twitter: {
      card: "summary",
      title: `Stellar Ledger #${formattedSequence}`,
      description: `View ledger details on StellarView Explorer`,
    },
  };
}

export default async function LedgerPage({ params }: Props) {
  const { sequence } = await params;
  const sequenceNum = parseInt(sequence, 10);

  if (isNaN(sequenceNum) || sequenceNum <= 0) {
    return notFound();
  }

  return <LedgerContent sequence={sequenceNum} />;
}
