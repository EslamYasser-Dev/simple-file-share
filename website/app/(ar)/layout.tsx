import type { Metadata } from "next";
import "../globals.css";

export const metadata: Metadata = {
  title: {
    default: "FileShare — مشاركة ملفات مستضافة ذاتيًا",
    template: "%s · FileShare",
  },
  description:
    "مساحة عمل ملفات مفتوحة المصدر بلغة Go واحدة: رفع قابل للاستئناف، نقل في الشبكة المحلية، أدوار وحصص، S3 أو قرص محلي، بالعربية والإنجليزية.",
  openGraph: {
    title: "FileShare — مشاركة ملفات مستضافة ذاتيًا",
    description:
      "رفع قابل للاستئناف، نقل مباشر في الشبكة المحلية، RBAC، S3، gRPC، وواجهة عربية/إنجليزية — مفتوح المصدر.",
    type: "website",
    locale: "ar_AR",
  },
};

export default function ArabicRootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ar" dir="rtl">
      <body className="min-h-screen bg-void font-sans text-paper antialiased">
        {children}
      </body>
    </html>
  );
}
