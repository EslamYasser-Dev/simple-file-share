import Link from "next/link";
import type { ReactNode } from "react";
import { APP_URL, GITHUB_URL } from "../lib/site";
import { getDict, lp, type Locale } from "../lib/i18n";
import { LanguageSwitcher } from "./language-switcher";

export function Background({ light = false }: { light?: boolean }) {
  return (
    <div
      aria-hidden
      className="pointer-events-none fixed inset-0 -z-10 bg-void"
    >
      <div
        className="absolute inset-0 opacity-[0.035]"
        style={{
          backgroundImage: light
            ? "radial-gradient(rgba(20,22,27,0.9) 0.5px, transparent 0.5px)"
            : "radial-gradient(rgba(242,239,233,0.9) 0.5px, transparent 0.5px)",
          backgroundSize: "3px 3px",
        }}
      />
    </div>
  );
}

export function Header({ locale }: { locale: Locale }) {
  const d = getDict(locale);
  return (
    <header className="sticky top-0 z-40 border-b border-paper/10 bg-void/95">
      <div className="mx-auto flex h-14 w-full max-w-5xl items-center justify-between px-5">
        <Link href={lp(locale, "/")} className="flex items-center gap-3">
          <span className="flex h-7 w-7 items-center justify-center border border-paper/25 bg-accent-dim">
            <svg viewBox="0 0 24 24" className="h-4 w-4" aria-hidden>
              <path
                fill="currentColor"
                className="text-accent"
                d="M3 7.5A1.5 1.5 0 0 1 4.5 6h4.2l2 2H19a1.5 1.5 0 0 1 1.5 1.5v8A1.5 1.5 0 0 1 19 19H4.5A1.5 1.5 0 0 1 3 17.5z"
              />
            </svg>
          </span>
          <span className="font-display text-base font-semibold tracking-tight text-paper">
            FileShare
          </span>
        </Link>
        <nav className="hidden items-center gap-6 font-mono text-[11px] uppercase tracking-wider text-muted md:flex">
          <Link
            href={lp(locale, "/#features")}
            className="transition hover:text-paper"
          >
            {d.nav.features}
          </Link>
          <Link
            href={lp(locale, "/#compare")}
            className="transition hover:text-paper"
          >
            {d.nav.compare}
          </Link>
          <Link
            href={lp(locale, "/getting-started/")}
            className="transition hover:text-paper"
          >
            {d.nav.install}
          </Link>
          <a
            href={GITHUB_URL}
            target="_blank"
            rel="noreferrer"
            className="transition hover:text-paper"
          >
            GitHub
          </a>
        </nav>
        <div className="flex items-center gap-3">
          <LanguageSwitcher />
          <a
            href={GITHUB_URL}
            target="_blank"
            rel="noreferrer"
            className="hidden border border-paper/20 px-3 py-1.5 font-mono text-[11px] uppercase tracking-wider text-paper/80 transition hover:border-paper/40 hover:text-paper sm:inline-flex"
          >
            {d.nav.source}
          </a>
          <a
            href={APP_URL}
            className="inline-flex border border-accent/50 bg-accent-dim px-3 py-1.5 font-mono text-[11px] uppercase tracking-wider text-accent transition hover:bg-accent/20"
          >
            {d.nav.openApp}
          </a>
        </div>
      </div>
    </header>
  );
}

export function Footer({ locale }: { locale: Locale }) {
  const d = getDict(locale);
  return (
    <footer className="border-t border-paper/10">
      <div className="mx-auto flex w-full max-w-5xl flex-col items-start justify-between gap-4 px-5 py-8 sm:flex-row sm:items-center">
        <p className="font-mono text-[11px] uppercase tracking-wider text-muted">
          © {new Date().getFullYear()} FileShare · {d.footer.tagline}
        </p>
        <div className="flex flex-wrap gap-5 font-mono text-[11px] uppercase tracking-wider text-muted">
          <Link
            href={lp(locale, "/#features")}
            className="transition hover:text-paper"
          >
            {d.footer.features}
          </Link>
          <Link
            href={lp(locale, "/getting-started/")}
            className="transition hover:text-paper"
          >
            {d.footer.install}
          </Link>
          <a
            href={GITHUB_URL}
            target="_blank"
            rel="noreferrer"
            className="transition hover:text-paper"
          >
            GitHub
          </a>
          <a href={APP_URL} className="transition hover:text-paper">
            {d.footer.app}
          </a>
        </div>
      </div>
    </footer>
  );
}

export function SiteShell({
  children,
  light = false,
  locale = "en",
}: {
  children: ReactNode;
  light?: boolean;
  locale?: Locale;
}) {
  return (
    <div className={light ? "light-site min-h-screen bg-void" : undefined}>
      <Background light={light} />
      <Header locale={locale} />
      {children}
      <Footer locale={locale} />
    </div>
  );
}

export function DocsShell({
  eyebrow,
  title,
  description,
  children,
  aside,
  locale = "en",
}: {
  eyebrow: string;
  title: string;
  description: string;
  children: ReactNode;
  aside?: ReactNode;
  locale?: Locale;
}) {
  const d = getDict(locale);
  return (
    <SiteShell locale={locale}>
      <main className="mx-auto w-full max-w-5xl px-5 pb-24 pt-10">
        <nav
          aria-label={d.docs.breadcrumb}
          className="mb-8 font-mono text-[11px] uppercase tracking-wider text-muted"
        >
          <Link href={lp(locale, "/")} className="transition hover:text-paper">
            {d.nav.home}
          </Link>
          <span className="mx-2 text-paper/30">/</span>
          <span className="text-paper/70">{eyebrow}</span>
        </nav>
        <header className="mb-10 max-w-3xl border-b border-paper/10 pb-8">
          <p className="eyebrow">{eyebrow}</p>
          <h1 className="mt-3 font-display text-3xl font-semibold tracking-tight text-paper sm:text-4xl">
            {title}
          </h1>
          <p className="mt-4 text-base leading-relaxed text-muted">
            {description}
          </p>
        </header>
        <div className={aside ? "grid gap-10 lg:grid-cols-[1fr_240px]" : undefined}>
          <div className="docs min-w-0">{children}</div>
          {aside ? <aside className="lg:pt-1">{aside}</aside> : null}
        </div>
      </main>
    </SiteShell>
  );
}

export function CodeBlock({ code, label }: { code: string; label?: string }) {
  return (
    <div className="my-5 overflow-hidden border border-paper/10 bg-black/40">
      {label ? (
        <div className="border-b border-paper/10 bg-paper/[0.03] px-4 py-2 font-mono text-[10px] uppercase tracking-wider text-muted">
          {label}
        </div>
      ) : null}
      <pre className="overflow-x-auto px-4 py-4 text-[13px] leading-relaxed text-paper/85">
        <code>{code}</code>
      </pre>
    </div>
  );
}

export function Step({
  n,
  title,
  children,
  locale = "en",
}: {
  n: string;
  title: string;
  children: ReactNode;
  locale?: Locale;
}) {
  const d = getDict(locale);
  return (
    <section className="panel mb-5 p-6">
      <div className="mb-3 flex items-center gap-3">
        <span className="font-mono text-[11px] uppercase tracking-wider text-accent">
          {d.docs.step} {n}
        </span>
        <h2 className="font-display text-lg font-semibold text-paper">
          {title}
        </h2>
      </div>
      <div className="space-y-3 text-sm leading-relaxed text-muted">
        {children}
      </div>
    </section>
  );
}

export function Callout({
  tone = "info",
  children,
}: {
  tone?: "info" | "warn";
  children: ReactNode;
}) {
  const styles =
    tone === "warn"
      ? "border-amber-500/35 bg-amber-500/10 text-amber-100"
      : "border-accent/35 bg-accent-dim text-paper/90";
  return (
    <div className={`my-5 border px-4 py-3 text-sm ${styles}`}>{children}</div>
  );
}

export function ProseH2({ children }: { children: ReactNode }) {
  return (
    <h2 className="mt-10 mb-4 font-display text-2xl font-semibold text-paper first:mt-0">
      {children}
    </h2>
  );
}

export function ProseH3({ children }: { children: ReactNode }) {
  return (
    <h3 className="mt-8 mb-3 font-display text-lg font-semibold text-paper">
      {children}
    </h3>
  );
}

export function P({ children }: { children: ReactNode }) {
  return <p className="my-3 text-[15px] leading-relaxed text-muted">{children}</p>;
}

export function UL({ items }: { items: ReactNode[] }) {
  return (
    <ul className="my-4 space-y-2.5 text-[15px] leading-relaxed text-muted">
      {items.map((item, i) => (
        <li key={i} className="flex gap-2.5">
          <span className="mt-2 h-1 w-1 shrink-0 bg-accent" />
          <span>{item}</span>
        </li>
      ))}
    </ul>
  );
}
