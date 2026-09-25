"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

/** EN / AR toggle that mirrors the current path between locales. */
export function LanguageSwitcher() {
  const pathname = usePathname() ?? "/";
  const isAr = pathname === "/ar" || pathname.startsWith("/ar/");
  const base = isAr ? pathname.replace(/^\/ar(?=\/|$)/, "") || "/" : pathname;
  const enHref = base;
  const arHref = base === "/" ? "/ar" : `/ar${base}`;

  const linkCls = (active: boolean) =>
    active
      ? "text-accent"
      : "text-muted transition hover:text-paper";

  return (
    <div
      className="flex items-center gap-1.5 font-mono text-[11px] tracking-wider"
      role="group"
      aria-label={isAr ? "اللغة" : "Language"}
    >
      <Link
        href={enHref}
        lang="en"
        hrefLang="en"
        aria-current={isAr ? undefined : "true"}
        aria-label={isAr ? "التبديل إلى الإنجليزية" : "Switch to English"}
        className={linkCls(!isAr)}
      >
        EN
      </Link>
      <span aria-hidden className="text-paper/30">
        /
      </span>
      <Link
        href={arHref}
        lang="ar"
        hrefLang="ar"
        aria-current={isAr ? "true" : undefined}
        aria-label={isAr ? "التبديل إلى العربية" : "Switch to Arabic"}
        className={linkCls(isAr)}
      >
        AR
      </Link>
    </div>
  );
}
