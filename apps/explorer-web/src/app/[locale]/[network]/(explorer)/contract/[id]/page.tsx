import { Metadata } from "next";
import { ContractContent } from "./contract-content";
import { JsonLd } from "@/components/common/json-ld";

type Props = {
  params: Promise<{ id: string; network: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const shortId = `${id.slice(0, 6)}...${id.slice(-6)}`;

  return {
    title: `Contract ${shortId}`,
    description: `View Soroban smart contract ${shortId}. Explore events, storage, and contract code on Stellar.`,
    openGraph: {
      title: `Soroban Contract ${shortId}`,
      description: `View smart contract details on StellarView Explorer`,
      type: "website",
    },
    twitter: {
      card: "summary",
      title: `Soroban Contract ${shortId}`,
      description: `View smart contract details on StellarView Explorer`,
    },
  };
}

export default async function ContractPage({ params }: Props) {
  const { id, network } = await params;

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    identifier: id,
    name: `Soroban Contract ${id.slice(0, 6)}…${id.slice(-6)}`,
    description: `A Soroban smart contract deployed on the Stellar ${network} network. Explore events, storage, and code.`,
    applicationCategory: "SmartContract",
    operatingSystem: "Stellar Network",
    author: {
      "@type": "Organization",
      name: "Stellar Network",
      url: "https://stellar.org",
    },
  };

  return (
    <>
      <JsonLd data={jsonLd} />
      <ContractContent id={id} />
    </>
  );
}
