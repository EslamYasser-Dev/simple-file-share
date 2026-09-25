import type { Metadata } from "next";
import "../globals.css";

export const metadata: Metadata = {
  title: {
    default: "FileShare — self-hosted file sharing",
    template: "%s · FileShare",
  },
  description:
    "Open-source file workspace in one Go binary: resumable uploads, LAN P2P, roles, quotas, S3 or local disk, English and Arabic.",
  icons: [{ rel: "icon", url: "/favicon.svg", type: "image/svg+xml" }],
  openGraph: {
    title: "FileShare — self-hosted file sharing",
    description:
      "Resumable uploads, LAN P2P, RBAC, S3, gRPC, and EN/AR UI — open source, one binary.",
    type: "website",
    locale: "en_US",
  },
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" dir="ltr">
      <body className="min-h-screen bg-void font-sans text-paper antialiased">
        {children}
      </body>
    </html>
  );
}
