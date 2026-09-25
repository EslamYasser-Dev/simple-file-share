import Link from "next/link";
import type { ReactNode } from "react";
import { SiteShell } from "./site";
import { APP_URL, GITHUB_URL } from "../lib/site";
import { getDict, lp, type Dict, type Locale } from "../lib/i18n";

const paths = {
  folder:
    "M3 7.5A1.5 1.5 0 0 1 4.5 6h4.2l2 2H19a1.5 1.5 0 0 1 1.5 1.5v8A1.5 1.5 0 0 1 19 19H4.5A1.5 1.5 0 0 1 3 17.5z",
  file: "M6 3.5h7l5 5V20a.5.5 0 0 1-.5.5h-11A.5.5 0 0 1 6 20zM13 3.5V9h5",
  clock: "M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18zM12 7.5V12l3 2",
  share:
    "M9.5 14.5l5-5M11 6.5l1.2-1.2a4 4 0 0 1 5.7 5.7L16.7 12M13 17.5l-1.2 1.2a4 4 0 0 1-5.7-5.7L7.3 12",
  trash:
    "M4.5 7h15M9.5 7V5.2a.7.7 0 0 1 .7-.7h3.6a.7.7 0 0 1 .7.7V7m-9 0 .8 12a1 1 0 0 0 1 .9h7.4a1 1 0 0 0 1-.9l.8-12",
  p2p: "M5 12.5a9.5 9.5 0 0 1 14 0M8.5 15.5a5.5 5.5 0 0 1 7 0M12 19h.01",
  search:
    "M11 18.5a7.5 7.5 0 1 0 0-15 7.5 7.5 0 0 0 0 15zM20.5 20.5l-4.2-4.2",
  pause: "M9.5 6.5v11M14.5 6.5v11",
  check: "M5 12.5l4.5 4.5L19 7.5",
  plus: "M12 6v12M6 12h12",
  laptop: "M5 6.5A1.5 1.5 0 0 1 6.5 5h11A1.5 1.5 0 0 1 19 6.5v9H5zM3 18.5h18",
  phone:
    "M8 3.5h8a1 1 0 0 1 1 1v15a1 1 0 0 1-1 1H8a1 1 0 0 1-1-1v-15a1 1 0 0 1 1-1zM11 17.5h2",
  arrow: "M5 12h14m0 0-5-5m5 5-5 5",
} as const;

type IconPath = (typeof paths)[keyof typeof paths];

function Ico({ d, className }: { d: IconPath; className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden
    >
      <path d={d} />
    </svg>
  );
}

function GitHubIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor" aria-hidden>
      <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
    </svg>
  );
}

const btnPrimary =
  "inline-flex items-center justify-center bg-accent px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-[#0b50d4]";
const btnSecondary =
  "inline-flex items-center justify-center border border-paper/25 px-5 py-2.5 text-sm font-medium text-paper transition hover:border-paper/50";
const linkMuted =
  "inline-flex items-center gap-2 px-1 py-2.5 text-sm text-muted underline-offset-4 transition hover:text-paper hover:underline";

function Chip({
  children,
  tone = "neutral",
}: {
  children: ReactNode;
  tone?: "neutral" | "accent";
}) {
  const styles =
    tone === "accent"
      ? "border-accent/30 bg-accent-dim text-accent"
      : "border-paper/15 bg-paper/[0.05] text-muted";
  return (
    <span
      className={`inline-flex shrink-0 items-center gap-1 rounded border px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wider ${styles}`}
    >
      {children}
    </span>
  );
}

function Hero({ locale, d }: { locale: Locale; d: Dict }) {
  return (
    <section className="mx-auto w-full max-w-5xl px-5 pb-24 pt-14 text-center sm:pt-20">
      <p className="eyebrow">{d.hero.eyebrow}</p>
      <h1 className="mx-auto mt-4 max-w-2xl font-display text-4xl font-bold leading-[1.08] tracking-tight text-paper sm:text-[3.25rem]">
        {d.hero.titleA}
        <br />
        {d.hero.titleB}
      </h1>
      <p className="mx-auto mt-5 max-w-xl text-base leading-relaxed text-muted">
        {d.hero.sub}
      </p>
      <div className="mt-7 flex flex-wrap items-center justify-center gap-3">
        <a href={APP_URL} className={btnPrimary}>
          {d.hero.openApp}
        </a>
        <Link href={lp(locale, "/getting-started/")} className={btnSecondary}>
          {d.hero.install}
        </Link>
        <a href={GITHUB_URL} target="_blank" rel="noreferrer" className={linkMuted}>
          <GitHubIcon className="h-4 w-4" />
          {d.hero.source}
        </a>
      </div>
      <p className="mx-auto mt-7 font-mono text-[10px] uppercase tracking-wider text-muted">
        {d.hero.strip}
      </p>
      <AppWindow d={d} />
    </section>
  );
}

function AppWindow({ d }: { d: Dict }) {
  const nav = [
    { label: d.mock.nav.myFiles, icon: paths.folder, active: true },
    { label: d.mock.nav.shared, icon: paths.share, active: false },
    { label: d.mock.nav.recent, icon: paths.clock, active: false },
    { label: d.mock.nav.trash, icon: paths.trash, active: false },
    { label: d.mock.nav.lan, icon: paths.p2p, active: false },
  ];
  const rows = d.mock.rows;

  return (
    <div className="relative mx-auto mt-12 w-full max-w-4xl text-start">
      <div className="overflow-hidden rounded-xl border border-paper/12 bg-abyss shadow-[0_30px_60px_-30px_rgba(16,20,30,0.45)]">
        <div className="flex items-center gap-3 border-b border-paper/10 bg-paper/[0.04] px-4 py-2.5">
          <span className="flex gap-1.5" aria-hidden>
            <span className="h-2.5 w-2.5 rounded-full bg-[#ff5f57]" />
            <span className="h-2.5 w-2.5 rounded-full bg-[#febc2e]" />
            <span className="h-2.5 w-2.5 rounded-full bg-[#28c840]" />
          </span>
          <span className="mx-auto flex w-48 items-center justify-center gap-1.5 rounded-md border border-paper/10 bg-paper/[0.05] px-3 py-1 font-mono text-[10px] text-muted">
            <Ico d={paths.search} className="h-3 w-3" />
            fileshare.local
          </span>
          <span className="flex h-6 w-6 items-center justify-center rounded-full bg-accent text-[9px] font-semibold text-white">
            AK
          </span>
        </div>

        <div className="grid sm:grid-cols-[190px_1fr]">
          <aside className="hidden border-e border-paper/10 p-3 sm:block">
            <ul className="space-y-0.5">
              {nav.map((item) => (
                <li
                  key={item.label}
                  className={`flex items-center gap-2 rounded px-2.5 py-1.5 text-[12px] ${
                    item.active
                      ? "bg-accent-dim font-medium text-accent"
                      : "text-muted"
                  }`}
                >
                  <Ico d={item.icon} className="h-3.5 w-3.5" />
                  {item.label}
                </li>
              ))}
            </ul>
            <div className="mt-6 border-t border-paper/10 pt-3">
              <div className="flex items-center justify-between font-mono text-[9px] uppercase tracking-wider text-muted">
                <span>{d.mock.storage}</span>
                <span>{d.mock.storageValue}</span>
              </div>
              <div className="mt-2 h-1.5 bg-paper/10">
                <div className="h-full w-[21%] bg-accent" />
              </div>
            </div>
          </aside>

          <div className="p-4 sm:p-5">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="font-mono text-[11px] text-muted">
                {d.mock.breadcrumb.myFiles}{" "}
                <span className="text-paper/40">/</span>{" "}
                {d.mock.breadcrumb.projects}{" "}
                <span className="text-paper/40">/</span>{" "}
                <span className="text-paper">{d.mock.breadcrumb.current}</span>
              </p>
              <div className="flex gap-2">
                <span className="flex items-center gap-1 rounded border border-paper/15 px-2.5 py-1 text-[11px] text-paper/85">
                  <Ico d={paths.plus} className="h-3 w-3" />
                  {d.mock.newFolder}
                </span>
                <span className="flex items-center gap-1 rounded bg-accent px-2.5 py-1 text-[11px] font-medium text-white">
                  {d.mock.upload}
                </span>
              </div>
            </div>

            <div className="mt-4 overflow-hidden rounded border border-paper/10">
              <div className="grid grid-cols-[1fr_64px_72px] gap-2 border-b border-paper/10 bg-paper/[0.04] px-3 py-2 font-mono text-[9px] uppercase tracking-wider text-muted sm:grid-cols-[1fr_72px_80px]">
                <span>{d.mock.cols.name}</span>
                <span>{d.mock.cols.size}</span>
                <span className="hidden sm:block">{d.mock.cols.modified}</span>
              </div>
              <ul className="divide-y divide-paper/8">
                {rows.map((row) => (
                  <li
                    key={row.name}
                    className="grid grid-cols-[1fr_64px_72px] items-center gap-2 px-3 py-2.5 sm:grid-cols-[1fr_72px_80px]"
                  >
                    <span className="flex min-w-0 items-center gap-2.5">
                      <Ico
                        d={row.kind === "dir" ? paths.folder : paths.file}
                        className={
                          row.kind === "dir"
                            ? "h-4 w-4 shrink-0 text-accent"
                            : "h-4 w-4 shrink-0 text-muted"
                        }
                      />
                      <span className="truncate text-[13px] text-paper">
                        {row.name}
                      </span>
                      {row.chip ? (
                        <Chip tone={row.chip.tone}>{row.chip.text}</Chip>
                      ) : null}
                    </span>
                    <span className="font-mono text-[10px] text-muted">
                      {row.size}
                    </span>
                    <span className="hidden font-mono text-[10px] text-muted sm:block">
                      {row.when}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </div>

      <div className="absolute -bottom-7 start-2 w-[290px] rounded-lg border border-paper/12 bg-abyss p-4 shadow-[0_20px_44px_-18px_rgba(16,20,30,0.4)] sm:start-10">
        <div className="flex items-center justify-between">
          <p className="text-xs font-semibold text-paper">
            {d.mock.card.uploading}
          </p>
          <span className="font-mono text-[10px] font-medium text-accent">
            62%
          </span>
        </div>
        <p className="mt-1 truncate font-mono text-[10px] text-muted">
          {d.mock.card.fileLine}
        </p>
        <div className="mt-2.5 h-1.5 rounded bg-paper/10">
          <div className="h-full w-[62%] rounded bg-accent" />
        </div>
        <div className="mt-3 flex items-center gap-2">
          <span className="flex items-center gap-1 rounded border border-paper/15 px-2 py-1 text-[10px] text-paper/85">
            <Ico d={paths.pause} className="h-3 w-3" />
            {d.mock.card.pause}
          </span>
          <span className="rounded border border-paper/15 px-2 py-1 text-[10px] text-muted">
            {d.mock.card.cancel}
          </span>
          <span className="ms-auto font-mono text-[8px] uppercase tracking-wider text-muted">
            {d.mock.card.resumeNote}
          </span>
        </div>
      </div>
    </div>
  );
}

function FeatureRow({
  id,
  row,
  visual,
  flip = false,
}: {
  id?: string;
  row: Dict["rows"][number];
  visual: ReactNode;
  flip?: boolean;
}) {
  return (
    <section
      id={id}
      className="mx-auto w-full max-w-5xl border-t border-paper/10 px-5 py-16"
    >
      <div className="grid items-center gap-10 lg:grid-cols-2">
        <div className={flip ? "lg:order-2" : undefined}>
          <p className="eyebrow">{row.eyebrow}</p>
          <h2 className="mt-3 font-display text-2xl font-semibold tracking-tight text-paper sm:text-3xl">
            {row.title}
          </h2>
          <p className="mt-4 text-[15px] leading-relaxed text-muted">{row.body}</p>
          <ul className="mt-5 space-y-2.5">
            {row.points.map((p) => (
              <li
                key={p}
                className="flex items-start gap-2.5 text-sm text-paper/85"
              >
                <Ico
                  d={paths.check}
                  className="mt-0.5 h-4 w-4 shrink-0 text-accent"
                />
                {p}
              </li>
            ))}
          </ul>
        </div>
        <div className={flip ? "lg:order-1" : undefined}>{visual}</div>
      </div>
    </section>
  );
}

function UploadsVisual({ d }: { d: Dict }) {
  const v = d.visuals.uploads;
  const queue = [
    { name: "Q3-report.pdf", meta: v.meta1, state: "62%", tone: "accent" as const },
    { name: "brand-assets.zip", meta: "118 MB", state: v.done, tone: "neutral" as const },
    { name: "demo-recording.mp4", meta: "640 MB", state: v.queued, tone: "neutral" as const },
  ];
  return (
    <div className="panel rounded-lg p-5 shadow-[0_18px_44px_-24px_rgba(16,20,30,0.35)]">
      <div className="flex items-center justify-between">
        <p className="text-sm font-semibold text-paper">{v.title}</p>
        <span className="font-mono text-[10px] uppercase tracking-wider text-muted">
          {v.count}
        </span>
      </div>
      <ul className="mt-4 space-y-3">
        {queue.map((q, i) => (
          <li key={q.name} className="rounded border border-paper/10 p-3">
            <div className="flex items-center justify-between gap-3">
              <span className="flex min-w-0 items-center gap-2.5">
                <Ico d={paths.file} className="h-4 w-4 shrink-0 text-muted" />
                <span className="truncate text-[13px] text-paper">{q.name}</span>
              </span>
              <Chip tone={q.tone}>{q.state}</Chip>
            </div>
            {i === 0 ? (
              <>
                <div className="mt-2.5 h-1.5 rounded bg-paper/10">
                  <div className="h-full w-[62%] rounded bg-accent" />
                </div>
                <div className="mt-2 flex items-center justify-between font-mono text-[10px] text-muted">
                  <span>{q.meta}</span>
                  <span className="flex gap-3">
                    <span className="text-paper/70">{v.pause}</span>
                    <span>{v.cancel}</span>
                  </span>
                </div>
              </>
            ) : (
              <p className="mt-1.5 font-mono text-[10px] text-muted">{q.meta}</p>
            )}
          </li>
        ))}
      </ul>
      <p className="mt-4 border-t border-paper/10 pt-3 font-mono text-[10px] uppercase tracking-wider text-muted">
        {v.footer}
      </p>
    </div>
  );
}

function ShareVisual({ d }: { d: Dict }) {
  const v = d.visuals.share;
  return (
    <div className="panel rounded-lg p-5 shadow-[0_18px_44px_-24px_rgba(16,20,30,0.35)]">
      <div className="flex items-center justify-between gap-3">
        <p className="min-w-0 truncate text-sm font-semibold text-paper">
          {v.title}
        </p>
        <Chip tone="accent">{v.linkChip}</Chip>
      </div>
      <div className="mt-4 flex items-center gap-2 rounded border border-paper/15 bg-paper/[0.04] px-3 py-2">
        <span className="min-w-0 flex-1 truncate font-mono text-[11px] text-paper/85">
          fileshare.local/s/8Kf2-xQ9r
        </span>
        <span className="shrink-0 rounded bg-accent px-2.5 py-1 text-[11px] font-semibold text-white">
          {v.copy}
        </span>
      </div>
      <dl className="mt-4 divide-y divide-paper/8 text-[13px]">
        <div className="flex items-center justify-between py-2.5">
          <dt className="text-muted">{v.expires}</dt>
          <dd className="text-paper">{v.expiresIn}</dd>
        </div>
        <div className="flex items-center justify-between py-2.5">
          <dt className="text-muted">{v.password}</dt>
          <dd className="flex items-center gap-2 text-paper">
            {v.off}
            <span className="inline-flex h-4 w-7 items-center rounded-full bg-paper/15 px-0.5">
              <span className="h-3 w-3 rounded-full bg-white shadow" />
            </span>
          </dd>
        </div>
        <div className="flex items-center justify-between py-2.5">
          <dt className="text-muted">{v.downloads}</dt>
          <dd className="text-paper">{v.allowCount}</dd>
        </div>
      </dl>
      <div className="mt-4 flex items-center justify-between border-t border-paper/10 pt-4">
        <span className="text-[12px] font-medium text-red-600/80">
          {v.revoke}
        </span>
        <span className="rounded bg-accent px-3 py-1.5 text-[12px] font-semibold text-white">
          {v.done}
        </span>
      </div>
    </div>
  );
}

function AdminVisual({ d }: { d: Dict }) {
  const v = d.visuals.admin;
  const pcts = [62, 18, 4];
  return (
    <div className="panel rounded-lg p-5 shadow-[0_18px_44px_-24px_rgba(16,20,30,0.35)]">
      <div className="flex items-center justify-between">
        <p className="text-sm font-semibold text-paper">{v.title}</p>
        <span className="font-mono text-[10px] uppercase tracking-wider text-muted">
          {v.quota}
        </span>
      </div>
      <ul className="mt-4 space-y-3.5">
        {v.people.map((m, i) => (
          <li key={m.name}>
            <div className="flex items-center justify-between gap-3">
              <span className="flex min-w-0 items-center gap-2.5">
                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-accent-dim font-mono text-[10px] font-semibold text-accent">
                  {m.initials}
                </span>
                <span className="truncate text-[13px] text-paper">{m.name}</span>
                <Chip tone={m.role === v.admin ? "accent" : "neutral"}>
                  {m.role}
                </Chip>
              </span>
              <span className="font-mono text-[10px] text-muted">{m.used}</span>
            </div>
            <div className="mt-1.5 ms-9 h-1.5 rounded bg-paper/10">
              <div
                className="h-full rounded bg-accent/70"
                style={{ width: `${pcts[i] ?? 0}%` }}
              />
            </div>
          </li>
        ))}
      </ul>
      <div className="mt-4 flex items-center justify-between border-t border-paper/10 pt-3">
        <span className="flex items-center gap-1.5 text-[12px] text-muted">
          <Ico d={paths.plus} className="h-3.5 w-3.5" />
          {v.invite}
        </span>
        <span className="font-mono text-[10px] uppercase tracking-wider text-muted">
          {v.footer}
        </span>
      </div>
    </div>
  );
}

function LanVisual({ d }: { d: Dict }) {
  const v = d.visuals.lan;
  return (
    <div className="panel rounded-lg p-5 shadow-[0_18px_44px_-24px_rgba(16,20,30,0.35)]">
      <div className="flex items-center justify-between">
        <p className="text-sm font-semibold text-paper">{v.title}</p>
        <Chip tone="accent">
          <Ico d={paths.p2p} className="h-3 w-3" />
          {v.direct}
        </Chip>
      </div>
      <div className="mt-5 flex items-center justify-between gap-3">
        <div className="flex-1 rounded border border-paper/12 bg-paper/[0.04] p-3 text-center">
          <Ico d={paths.laptop} className="mx-auto h-6 w-6 text-accent" />
          <p className="mt-2 text-[12px] font-medium text-paper">
            {v.thisLaptop}
          </p>
          <p className="font-mono text-[9px] uppercase tracking-wider text-muted">
            {v.sender}
          </p>
        </div>
        <div className="flex flex-col items-center gap-1 text-accent">
          <Ico d={paths.arrow} className="h-5 w-5 rtl:-scale-x-100" />
          <span className="font-mono text-[8px] uppercase tracking-wider text-muted">
            WebRTC
          </span>
        </div>
        <div className="flex-1 rounded border border-paper/12 bg-paper/[0.04] p-3 text-center">
          <Ico d={paths.phone} className="mx-auto h-6 w-6 text-accent" />
          <p className="mt-2 text-[12px] font-medium text-paper">
            {v.phoneName}
          </p>
          <p className="font-mono text-[9px] uppercase tracking-wider text-muted">
            {v.receiver}
          </p>
        </div>
      </div>
      <p className="mt-5 border-t border-paper/10 pt-3 font-mono text-[10px] uppercase tracking-wider text-muted">
        {v.footer}
      </p>
    </div>
  );
}

function Capabilities({ d }: { d: Dict }) {
  return (
    <section
      id="capabilities"
      className="mx-auto w-full max-w-5xl border-t border-paper/10 px-5 py-16"
    >
      <div className="mb-10 max-w-2xl">
        <p className="eyebrow">{d.capabilities.eyebrow}</p>
        <h2 className="mt-3 font-display text-3xl font-semibold tracking-tight text-paper sm:text-[2.15rem]">
          {d.capabilities.title}
        </h2>
        <p className="mt-3 text-[15px] leading-relaxed text-muted">
          {d.capabilities.sub}
        </p>
      </div>
      <ol className="grid gap-x-10 sm:grid-cols-2 lg:grid-cols-3">
        {d.capabilities.items.map((f) => (
          <li key={f.n} className="border-t border-paper/10 py-6">
            <span className="font-mono text-[11px] tracking-wider text-accent">
              {f.n}
            </span>
            <h3 className="mt-2 font-display text-base font-semibold text-paper">
              {f.title}
            </h3>
            <p className="mt-2 text-sm leading-relaxed text-muted">{f.body}</p>
          </li>
        ))}
      </ol>
    </section>
  );
}

function Comparison({ d }: { d: Dict }) {
  return (
    <section
      id="compare"
      className="mx-auto w-full max-w-5xl border-t border-paper/10 px-5 py-16"
    >
      <div className="mb-10 max-w-2xl">
        <p className="eyebrow">{d.compare.eyebrow}</p>
        <h2 className="mt-3 font-display text-3xl font-semibold tracking-tight text-paper sm:text-[2.15rem]">
          {d.compare.title}
        </h2>
        <p className="mt-3 text-[15px] leading-relaxed text-muted">
          {d.compare.sub}
        </p>
      </div>
      <div className="overflow-x-auto rounded-lg border border-paper/10">
        <table className="w-full min-w-[640px] text-start text-sm">
          <thead>
            <tr className="border-b border-paper/10 bg-paper/[0.04]">
              <th className="px-4 py-3 font-mono text-[10px] font-medium uppercase tracking-wider text-muted">
                {d.compare.dimension}
              </th>
              <th className="px-4 py-3 font-mono text-[10px] font-medium uppercase tracking-wider text-muted">
                Dropbox
              </th>
              <th className="px-4 py-3 font-mono text-[10px] font-medium uppercase tracking-wider text-muted">
                MinIO
              </th>
              <th className="px-4 py-3 font-mono text-[10px] font-medium uppercase tracking-wider text-accent">
                FileShare
              </th>
            </tr>
          </thead>
          <tbody>
            {d.compare.rows.map((row) => (
              <tr
                key={row.dim}
                className="border-b border-paper/8 last:border-0"
              >
                <td className="px-4 py-3 text-muted">{row.dim}</td>
                <td className="px-4 py-3 text-muted/80">{row.dropbox}</td>
                <td className="px-4 py-3 text-muted/80">{row.minio}</td>
                <td className="px-4 py-3 font-medium text-paper">
                  {row.fileshare}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function HowItWorks({ locale, d }: { locale: Locale; d: Dict }) {
  return (
    <section
      id="how"
      className="mx-auto w-full max-w-5xl border-t border-paper/10 px-5 py-16"
    >
      <div className="mb-10 max-w-2xl">
        <p className="eyebrow">{d.how.eyebrow}</p>
        <h2 className="mt-3 font-display text-3xl font-semibold tracking-tight text-paper sm:text-[2.15rem]">
          {d.how.title}
        </h2>
      </div>
      <ol className="grid gap-0 md:grid-cols-3">
        {d.how.steps.map((s, i) => (
          <li
            key={s.n}
            className={`border-t border-paper/10 py-6 md:border-t-0 md:pe-8 ${
              i > 0 ? "md:border-s md:border-paper/10 md:ps-8" : ""
            }`}
          >
            <span className="font-display text-2xl font-bold text-accent">
              {s.n}
            </span>
            <h3 className="mt-2 font-display text-lg font-semibold text-paper">
              {s.title}
            </h3>
            <p className="mt-2 text-sm leading-relaxed text-muted">{s.body}</p>
          </li>
        ))}
      </ol>
      <div className="panel mt-6 flex flex-wrap items-center justify-between gap-3 rounded-lg px-4 py-3.5">
        <code className="font-mono text-[13px] text-paper">
          $ docker compose up --build
        </code>
        <span className="font-mono text-[10px] uppercase tracking-wider text-muted">
          {d.how.cmdNote}
        </span>
      </div>
      <p className="mt-6 font-mono text-[11px] uppercase tracking-wider text-muted">
        <Link
          href={lp(locale, "/getting-started/")}
          className="underline-accent text-accent"
        >
          {d.how.guide}
        </Link>
      </p>
    </section>
  );
}

function Stack({ d }: { d: Dict }) {
  return (
    <section
      id="stack"
      className="mx-auto w-full max-w-5xl border-t border-paper/10 px-5 py-16"
    >
      <div className="mb-10 max-w-2xl">
        <p className="eyebrow">{d.stack.eyebrow}</p>
        <h2 className="mt-3 font-display text-3xl font-semibold tracking-tight text-paper sm:text-[2.15rem]">
          {d.stack.title}
        </h2>
        <p className="mt-3 text-[15px] leading-relaxed text-muted">
          {d.stack.sub}
        </p>
      </div>
      <div className="grid gap-8 md:grid-cols-3">
        {d.stack.cols.map((col) => (
          <div key={col.label} className="border-t border-paper/10 pt-4">
            <h3 className="font-mono text-[11px] uppercase tracking-wider text-accent">
              {col.label}
            </h3>
            <ul className="mt-4 space-y-2 text-sm text-muted">
              {col.items.map((item) => (
                <li key={item} className="border-b border-paper/8 pb-2">
                  {item}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </section>
  );
}

function Closing({ locale, d }: { locale: Locale; d: Dict }) {
  return (
    <section className="mx-auto w-full max-w-5xl border-t border-paper/10 px-5 py-16">
      <div className="max-w-2xl">
        <p className="eyebrow">{d.closing.eyebrow}</p>
        <h2 className="mt-3 font-display text-3xl font-semibold tracking-tight text-paper">
          {d.closing.title}
        </h2>
        <p className="mt-3 text-[15px] leading-relaxed text-muted">
          {d.closing.sub}
        </p>
        <div className="mt-6 flex flex-wrap gap-3">
          <a href={APP_URL} className={btnPrimary}>
            {d.closing.openApp}
          </a>
          <a
            href={GITHUB_URL}
            target="_blank"
            rel="noreferrer"
            className={btnSecondary}
          >
            <GitHubIcon className="me-2 h-4 w-4" />
            GitHub
          </a>
          <Link
            href={lp(locale, "/getting-started/")}
            className={btnSecondary}
          >
            {d.closing.install}
          </Link>
        </div>
      </div>
    </section>
  );
}

export function Landing({ locale }: { locale: Locale }) {
  const d = getDict(locale);
  return (
    <SiteShell light locale={locale}>
      <main>
        <Hero locale={locale} d={d} />
        <FeatureRow id="features" row={d.rows[0]} visual={<UploadsVisual d={d} />} />
        <FeatureRow flip row={d.rows[1]} visual={<ShareVisual d={d} />} />
        <FeatureRow row={d.rows[2]} visual={<AdminVisual d={d} />} />
        <FeatureRow flip row={d.rows[3]} visual={<LanVisual d={d} />} />
        <Capabilities d={d} />
        <Comparison d={d} />
        <HowItWorks locale={locale} d={d} />
        <Stack d={d} />
        <Closing locale={locale} d={d} />
      </main>
    </SiteShell>
  );
}

