import { Metadata } from "next";
import { notFound } from "next/navigation";
import { TransactionContent } from "./transaction-content";
import { JsonLd } from "@/components/common/json-ld";

type Props = {
  params: Promise<{ hash: string; network: string; locale: string }>;
};

function isValidTransactionHash(hash: string): boolean {
  return hash.length === 64 && /^[a-f0-9]+$/i.test(hash);
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { hash } = await params;
  const shortHash = `${hash.slice(0, 8)}...${hash.slice(-8)}`;

  return {
    title: `Transaction ${shortHash}`,
    description: `View details of Stellar transaction ${shortHash}. Explore operations, effects, and raw XDR data.`,
    openGraph: {
      title: `Stellar Transaction ${shortHash}`,
      description: `View transaction details on StellarView Explorer`,
      type: "website",
    },
    twitter: {
      card: "summary",
      title: `Stellar Transaction ${shortHash}`,
      description: `View transaction details on StellarView Explorer`,
    },
  };
}

export default async function TransactionPage({ params }: Props) {
  const { hash, network } = await params;

  if (!isValidTransactionHash(hash)) {
    return notFound();
  }

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "DigitalDocument",
    identifier: hash,
    name: `Stellar Transaction ${hash.slice(0, 8)}…${hash.slice(-8)}`,
    description: `A transaction recorded on the Stellar ${network} network.`,
    encodingFormat: "application/json+stellar-xdr",
    about: {
      "@type": "Thing",
      name: "Stellar Blockchain Transaction",
      description: "A cryptographically signed operation on the Stellar network",
    },
  };

  return (
    <>
      <JsonLd data={jsonLd} />
      <TransactionContent hash={hash} />
    </>
  );
}
