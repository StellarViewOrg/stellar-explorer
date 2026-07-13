import { Metadata } from "next";
import { AssetContent } from "./asset-content";
import { parseAssetSlug } from "@/lib/utils";
import { JsonLd } from "@/components/common/json-ld";

type Props = {
  params: Promise<{ slug: string; network: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const parsed = parseAssetSlug(slug);

  if (!parsed) {
    return {
      title: "Asset Not Found",
    };
  }

  const { code, issuer } = parsed;
  const isNative = issuer === "native";
  const shortIssuer = isNative ? "Native" : `${issuer.slice(0, 4)}...${issuer.slice(-4)}`;

  return {
    title: isNative ? "XLM - Stellar Lumens" : `${code} Asset`,
    description: isNative
      ? "XLM (Stellar Lumens) is the native asset of the Stellar network."
      : `View details of ${code} asset on Stellar. Issuer: ${shortIssuer}. Explore supply, holders, and flags.`,
    openGraph: {
      title: isNative ? "Stellar Lumens (XLM)" : `${code} - Stellar Asset`,
      description: isNative
        ? "Native asset of the Stellar network"
        : `View ${code} asset details on Stellar Explorer`,
      type: "website",
    },
    twitter: {
      card: "summary",
      title: isNative ? "Stellar Lumens (XLM)" : `${code} - Stellar Asset`,
      description: isNative
        ? "Native asset of the Stellar network"
        : `View ${code} asset details on Stellar Explorer`,
    },
  };
}

export default async function AssetPage({ params }: Props) {
  const { slug, network } = await params;
  const parsed = parseAssetSlug(slug);

  const isNative = parsed?.issuer === "native";
  const code = parsed?.code ?? slug;
  const issuer = parsed?.issuer ?? "";

  const jsonLd = isNative
    ? {
        "@context": "https://schema.org",
        "@type": "Currency",
        name: "Stellar Lumens",
        alternateName: "XLM",
        description: "XLM is the native asset of the Stellar blockchain network.",
        url: "https://stellar.org",
      }
    : {
        "@context": "https://schema.org",
        "@type": "FinancialProduct",
        name: `${code} — Stellar Asset`,
        description: `${code} is a tokenized asset issued on the Stellar ${network} network by ${issuer.slice(0, 6)}…${issuer.slice(-6)}.`,
        provider: {
          "@type": "Organization",
          identifier: issuer,
          name: `Issuer ${issuer.slice(0, 6)}…${issuer.slice(-6)}`,
        },
        identifier: slug,
      };

  return (
    <>
      <JsonLd data={jsonLd} />
      <AssetContent slug={slug} />
    </>
  );
}
