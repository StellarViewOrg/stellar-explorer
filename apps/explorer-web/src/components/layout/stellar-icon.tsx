import { cn } from "@/lib/utils";

interface StellarIconProps {
  className?: string;
}

export function StellarIcon({ className }: StellarIconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={cn("size-6", className)}
    >
      <path
        d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function StellarLogo({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 100 100"
      className={cn("size-8", className)}
    >
      <defs>
        <linearGradient id="stellar-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="var(--primary)" />
          <stop offset="100%" stopColor="var(--chart-1)" />
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(#stellar-gradient)" />
      <g fill="white">
        <path d="M68.5 31.3L25.7 49.6l2.6 5.9 42.8-18.3L68.5 31.3z" />
        <path d="M71.1 37.2L28.3 55.5l2.6 5.9 42.8-18.3L71.1 37.2z" />
        <path d="M73.7 43.1L30.9 61.4l-5.2-2.2 2.6-5.9 38-16.3 2.6 6z" />
        <path d="M25.7 67.3l5.2 2.2 42.8-18.3-2.6-6-42.8 18.3-2.6 5.9 42.8-18.3" />
        <circle cx="28" cy="36" r="6" />
        <circle cx="72" cy="64" r="6" />
      </g>
    </svg>
  );
}
