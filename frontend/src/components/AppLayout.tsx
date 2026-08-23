import type { ReactNode } from "react";
import { Outlet } from "react-router-dom";
import { Navbar } from "./Navbar";

interface AppLayoutProps {
  children?: ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col bg-paper text-ink selection:bg-ochre/20 selection:text-ink">
      <Navbar />
      <main className="flex-1">
        {children ? children : <Outlet />}
      </main>
    </div>
  );
}
