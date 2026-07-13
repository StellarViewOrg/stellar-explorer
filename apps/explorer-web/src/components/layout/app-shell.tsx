import { Header } from "./header";
import { Dock } from "./dock";
import { PageTransition } from "./page-transition";
import type { ReactNode } from "react";

interface AppShellProps {
  children: ReactNode;
}

export function AppShell({ children }: AppShellProps) {
  return (
    <div className="bg-background min-h-screen">
      <Dock />
      <Header />
      <div className="pt-16 md:pl-[72px]">
        <main>
          <div className="container mx-auto min-h-full max-w-7xl p-4 md:p-6 lg:p-8">
            <PageTransition>{children}</PageTransition>
          </div>
        </main>
      </div>
    </div>
  );
}
