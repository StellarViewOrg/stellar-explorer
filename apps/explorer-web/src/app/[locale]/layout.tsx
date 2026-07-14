import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, setRequestLocale, getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import { Analytics } from "@vercel/analytics/next";
import Script from "next/script";
import { Providers } from "@/lib/providers";
import { locales, type Locale } from "@/i18n/config";
import "../globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "metadata" });

  const title = t("title");
  const brand = t("brand");
  const description = t("description");

  return {
    metadataBase: new URL(
      process.env.NEXT_PUBLIC_BASE_URL || "https://stellar-explorer.acachete.xyz"
    ),
    title: {
      default: title,
      template: `%s | ${brand}`,
    },
    description: description,
    keywords: [
      "StellarView Explorer",
      "StellarView",
      "Stellar",
      "blockchain explorer",
      "block explorer",
      "XLM",
      "Lumens",
      "Soroban",
      "Soroban smart contracts",
      "Stellar Network",
      "Stellar blockchain",
      "Stellar transactions",
      "Stellar mainnet",
      "Stellar testnet",
      "crypto explorer",
      "DeFi",
    ],
    authors: [{ name: "StellarView" }],
    icons: {
      icon: "/stellar-explorer.png",
      apple: "/stellar-explorer.png",
    },
    openGraph: {
      type: "website",
      locale: locale === "es" ? "es_ES" : "en_US",
      siteName: title,
      title: title,
      description: description,
      images: [{ url: "/stellar-explorer.png", width: 1024, height: 1024, alt: title }],
    },
    twitter: {
      card: "summary_large_image",
      title: title,
      description: description,
      images: ["/stellar-explorer.png"],
    },
    robots: {
      index: true,
      follow: true,
    },
  };
}

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;

  if (!locales.includes(locale as Locale)) {
    notFound();
  }

  setRequestLocale(locale);
  const messages = await getMessages();

  return (
    <html lang={locale} suppressHydrationWarning>
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <NextIntlClientProvider locale={locale} messages={messages}>
          <Providers>{children}</Providers>
        </NextIntlClientProvider>
        <Analytics />
        <Script
          defer
          src="https://cloud.umami.is/script.js"
          data-website-id="381d4776-b0ac-43c4-a5d2-eacb5d0beb9b"
          strategy="afterInteractive"
        />
      </body>
    </html>
  );
}
