import { ImageResponse } from "next/og";

export const alt = "Stellar Ledger";
export const size = {
  width: 1200,
  height: 630,
};
export const contentType = "image/png";

export default async function Image({ params }: { params: { sequence: string; locale: string } }) {
  const sequence = params.sequence;
  const locale = params.locale || "en";

  // Format sequence with commas
  const formattedSequence = parseInt(sequence, 10).toLocaleString();

  const title = locale === "es" ? "Libro Mayor" : "Ledger";
  const subtitle = locale === "es" ? "Red Stellar" : "Stellar Network";

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
            "radial-gradient(circle at 20% 30%, rgba(59, 130, 246, 0.15) 0%, transparent 50%), radial-gradient(circle at 80% 70%, rgba(6, 182, 212, 0.1) 0%, transparent 50%)",
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
            background: "linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%)",
            boxShadow: "0 20px 40px rgba(59, 130, 246, 0.3)",
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
            <rect width="18" height="18" x="3" y="3" rx="2" />
            <path d="M3 9h18" />
            <path d="M3 15h18" />
            <path d="M9 3v18" />
            <path d="M15 3v18" />
          </svg>
        </div>

        {/* Badge */}
        <div
          style={{
            display: "flex",
            padding: "8px 20px",
            background: "rgba(59, 130, 246, 0.2)",
            borderRadius: 24,
            color: "#60a5fa",
            fontSize: 18,
            fontWeight: 600,
          }}
        >
          {title}
        </div>

        {/* Ledger Sequence */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 8,
            fontSize: 56,
            fontWeight: 700,
            color: "white",
            letterSpacing: "-0.02em",
          }}
        >
          <span style={{ color: "rgba(255, 255, 255, 0.4)" }}>#</span>
          {formattedSequence}
        </div>

        {/* Subtitle */}
        <div
          style={{
            display: "flex",
            fontSize: 24,
            color: "rgba(255, 255, 255, 0.6)",
            marginTop: 8,
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
          background: "linear-gradient(90deg, #3b82f6 0%, #06b6d4 100%)",
        }}
      />
    </div>,
    {
      ...size,
    }
  );
}
