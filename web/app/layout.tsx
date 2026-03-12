import Link from "next/link";
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "AgentFence Admin",
  description: "Dashboard for audit events, approvals, and policy status.",
};

const navItems = [
  { href: "/", label: "Dashboard" },
  { href: "/audit", label: "Audit" },
  { href: "/approvals", label: "Approvals" },
];

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <div className="app-shell">
          <header className="topbar">
            <div>
              <p className="eyebrow">AgentFence</p>
              <h1>Admin Console</h1>
            </div>
            <nav className="nav">
              {navItems.map((item) => (
                <Link key={item.href} href={item.href}>
                  {item.label}
                </Link>
              ))}
            </nav>
          </header>
          <main className="content">{children}</main>
        </div>
      </body>
    </html>
  );
}
