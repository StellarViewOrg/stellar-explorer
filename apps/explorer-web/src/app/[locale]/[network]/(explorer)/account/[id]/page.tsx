import { Metadata } from "next";
import { AccountContent } from "./account-content";
import { JsonLd } from "@/components/common/json-ld";

type Props = {
  params: Promise<{ id: string; network: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const shortId = `${id.slice(0, 6)}...${id.slice(-6)}`;

  return {
    title: `Account ${shortId}`,
    description: `View Stellar account ${shortId}. Explore balances, transactions, operations, and signers.`,
    openGraph: {
      title: `Stellar Account ${shortId}`,
      description: `View account details on Stellar Explorer`,
      type: "website",
    },
    twitter: {
      card: "summary",
      title: `Stellar Account ${shortId}`,
      description: `View account details on Stellar Explorer`,
    },
  };
}

export default async function AccountPage({ params }: Props) {
  const { id, network } = await params;

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "FinancialAccount",
    identifier: id,
    name: `Stellar Account ${id.slice(0, 6)}…${id.slice(-6)}`,
    description: `A Stellar ${network} account. Explore balances, signers, and transaction history.`,
    accountId: id,
    provider: {
      "@type": "Organization",
      name: "Stellar Network",
      url: "https://stellar.org",
    },
  };

  return (
    <>
      <JsonLd data={jsonLd} />
      <AccountContent id={id} />
    </>
  );
}
