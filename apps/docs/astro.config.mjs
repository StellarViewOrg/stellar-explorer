import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import rehypeExternalLinks from "rehype-external-links";

const site = "https://docs.stellar-explorer.acachete.xyz";
const base = "/";
const githubUrl = "https://github.com/salazarsebas/stellar-explorer";

export default defineConfig({
  site,
  base,
  markdown: {
    rehypePlugins: [
      [
        rehypeExternalLinks,
        { target: "_blank", rel: ["noopener", "noreferrer"] },
      ],
    ],
  },
  integrations: [
    starlight({
      title: "Stellar View | Stellar Explorer",
      favicon: "/favicon.png",
      customCss: ["./src/styles/custom.css"],
      disable404Route: true,
      defaultLocale: "root",
      locales: {
        root: { label: "English", lang: "en" },
        es: { label: "Español", lang: "es" },
      },
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: githubUrl,
        },
      ],
      sidebar: [
        { slug: "getting-started" },
        {
          label: "Guides",
          translations: { es: "Guías" },
          items: [
            { slug: "guides/exploring-transactions" },
            { slug: "guides/exploring-accounts" },
            { slug: "guides/exploring-assets" },
            { slug: "guides/exploring-contracts" },
            { slug: "guides/exploring-ledgers" },
            { slug: "guides/using-watchlist" },
            { slug: "guides/network-switching" },
          ],
        },
        {
          label: "Architecture",
          translations: { es: "Arquitectura" },
          items: [
            { slug: "architecture/overview" },
            { slug: "architecture/data-flow" },
            { slug: "architecture/routing" },
            { slug: "architecture/providers" },
            { slug: "architecture/indexer" },
            { slug: "architecture/tui" },
            { slug: "architecture/tui-product-roadmap" },
            { slug: "architecture/tui-technical-roadmap" },
          ],
        },
        {
          label: "Development",
          translations: { es: "Desarrollo" },
          items: [
            { slug: "development/setup" },
            { slug: "development/project-structure" },
            { slug: "development/testing" },
            { slug: "development/i18n" },
          ],
        },
      ],
    }),
  ],
});
