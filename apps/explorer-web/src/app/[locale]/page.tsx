import { redirect } from "next/navigation";
import { defaultLocale } from "@/i18n/config";

export default async function LocaleRootPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  redirect(`/${locale || defaultLocale}/mainnet`);
}
