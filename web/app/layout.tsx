import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "AgentFence",
  description: "Admin UI for secure MCP gateway operations.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
