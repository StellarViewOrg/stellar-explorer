import type { NextConfig } from "next";
import path from "node:path";
import createNextIntlPlugin from "next-intl/plugin";
import withBundleAnalyzer from "@next/bundle-analyzer";

const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");
const withAnalyzer = withBundleAnalyzer({ enabled: process.env.ANALYZE === "true" });

const LOCALES = ["en", "es", "pt", "fr", "de", "zh", "ja", "ko", "it"];

const nextConfig: NextConfig = {
  async redirects() {
    return LOCALES.flatMap((locale) => [
      {
        source: `/${locale}/public`,
        destination: `/${locale}/mainnet`,
        permanent: true,
      },
      {
        source: `/${locale}/public/:path*`,
        destination: `/${locale}/mainnet/:path*`,
        permanent: true,
      },
    ]);
  },
  turbopack: {
    // In the monorepo, Turbopack needs the workspace root to resolve hoisted deps.
    root: path.join(__dirname, "../.."),
  },
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "**",
      },
    ],
  },
};

export default withAnalyzer(withNextIntl(nextConfig));
