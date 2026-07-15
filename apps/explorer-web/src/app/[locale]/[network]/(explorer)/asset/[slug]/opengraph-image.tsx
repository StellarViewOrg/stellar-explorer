import { ImageResponse } from "next/og";
import { parseAssetSlug } from "@/lib/utils";

export const alt = "Stellar Asset";
export const size = {
  width: 1200,
  height: 630,
};
export const contentType = "image/png";

export default async function Image({ params }: { params: { slug: string; locale: string } }) {
  const slug = params.slug;
  const locale = params.locale || "en";

  const parsed = parseAssetSlug(slug);
  const code = parsed?.code || "Asset";
  const issuer = parsed?.issuer || "";
  const isNative = issuer === "native";
  const shortIssuer = isNative ? "" : `${issuer.slice(0, 6)}...${issuer.slice(-6)}`;

  const title = locale === "es" ? "Activo" : "Asset";
  const subtitle = locale === "es" ? "Red Stellar" : "Stellar Network";
  const nativeLabel = locale === "es" ? "Activo Nativo" : "Native Asset";

  return new ImageResponse(
    <div
      style={{
        height: "100%",
        width: "100%",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        background: "linear-gradient(135deg, #0a0a0f 0%, #1a1a2e 50%, #0f0f1a 100%)",
        fontFamily: "system-ui, sans-serif",
      }}
    >
      {/* Background pattern */}
      <div
        style={{
          position: "absolute",
          inset: 0,
          backgroundImage:
            "radial-gradient(circle at 20% 30%, rgba(234, 179, 8, 0.15) 0%, transparent 50%), radial-gradient(circle at 80% 70%, rgba(249, 115, 22, 0.1) 0%, transparent 50%)",
        }}
      />

      {/* Content */}
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          gap: 20,
        }}
      >
        {/* Icon */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: 100,
            height: 100,
            borderRadius: 20,
            background: isNative
              ? "linear-gradient(135deg, #eab308 0%, #f59e0b 100%)"
              : "linear-gradient(135deg, #f97316 0%, #ea580c 100%)",
            boxShadow: isNative
              ? "0 20px 40px rgba(234, 179, 8, 0.3)"
              : "0 20px 40px rgba(249, 115, 22, 0.3)",
          }}
        >
          <svg
            width="48"
            height="48"
            viewBox="0 0 24 24"
            fill="none"
            stroke="white"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="12" cy="12" r="10" />
            <path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8" />
            <path d="M12 18V6" />
          </svg>
        </div>

        {/* Badge */}
        <div
          style={{
            display: "flex",
            padding: "8px 20px",
            background: isNative ? "rgba(234, 179, 8, 0.2)" : "rgba(249, 115, 22, 0.2)",
            borderRadius: 24,
            color: isNative ? "#fbbf24" : "#fb923c",
            fontSize: 18,
            fontWeight: 600,
          }}
        >
          {title}
        </div>

        {/* Asset Code */}
        <div
          style={{
            display: "flex",
            fontSize: 64,
            fontWeight: 700,
            color: "white",
            letterSpacing: "-0.02em",
          }}
        >
          {code}
        </div>

        {/* Issuer or Native label */}
        <div
          style={{
            display: "flex",
            fontSize: 20,
            color: "rgba(255, 255, 255, 0.6)",
            fontFamily: "monospace",
            background: "rgba(255, 255, 255, 0.05)",
            padding: "12px 24px",
            borderRadius: 8,
            border: "1px solid rgba(255, 255, 255, 0.1)",
          }}
        >
          {isNative ? nativeLabel : shortIssuer}
        </div>

        {/* Subtitle */}
        <div
          style={{
            display: "flex",
            fontSize: 24,
            color: "rgba(255, 255, 255, 0.5)",
            marginTop: 4,
          }}
        >
          {subtitle}
        </div>
      </div>

      {/* StellarView Explorer branding */}
      <div
        style={{
          position: "absolute",
          bottom: 40,
          display: "flex",
          alignItems: "center",
          gap: 12,
        }}
      >
        <div
          style={{
            width: 32,
            height: 32,
            borderRadius: 8,
            background: "linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="white"
            strokeWidth="2"
          >
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
          </svg>
        </div>
        <span style={{ color: "rgba(255, 255, 255, 0.5)", fontSize: 18 }}>
          StellarView Explorer
        </span>
      </div>

      {/* Bottom gradient */}
      <div
        style={{
          position: "absolute",
          bottom: 0,
          left: 0,
          right: 0,
          height: 4,
          background: isNative
            ? "linear-gradient(90deg, #eab308 0%, #f59e0b 100%)"
            : "linear-gradient(90deg, #f97316 0%, #ea580c 100%)",
        }}
      />
    </div>,
    {
      ...size,
    }
  );
}
