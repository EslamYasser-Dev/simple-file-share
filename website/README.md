# FileShare Website

Standalone Next.js marketing landing page for [Simple FileShare](../readme.md). Lives in its own directory and does not depend on `frontend/` or `backend/`.

## Stack

| Layer | Choice |
|-------|--------|
| Framework | Next.js 15 (App Router) |
| UI | React 19 + TypeScript |
| Styling | Tailwind CSS 4 |
| Output | Static export (`output: "export"`) |

## Quick start

```bash
cd website
yarn install
yarn dev
```

Open [http://localhost:3001](http://localhost:3001).

## Scripts

| Command | Description |
|---------|-------------|
| `yarn dev` | Dev server on port **3001** |
| `yarn build` | Production static export → `out/` |
| `yarn start` | Serve the production build (port 3001) |
| `yarn lint` | ESLint (`next/core-web-vitals` + TypeScript) |

Typecheck separately:

```bash
npx tsc --noEmit
```

## Configuration

Optional environment variables (defaults in parentheses):

| Variable | Purpose | Default |
|----------|---------|---------|
| `NEXT_PUBLIC_APP_URL` | CTA link to the FileShare app | `http://localhost:5173` |
| `NEXT_PUBLIC_GITHUB_URL` | Source / docs links | project GitHub repo |

Example `.env.local`:

```bash
NEXT_PUBLIC_APP_URL=https://app.example.com
NEXT_PUBLIC_GITHUB_URL=https://github.com/EslamYasser-Dev/simple-file-share
```

These are inlined at **build time** — set them before `yarn build`.

## Page sections

Defined in `components/landing.tsx` (product screenshot-forward layout, locale-driven):

1. **Header** — logo, in-page nav (Features / Compare / Install / GitHub), Open app
2. **Hero** — tagline, CTAs, and a large file-browser app window mock (sidebar, file table, floating upload card)
3. **Feature rows** (alternating, `#features`) — Uploads, Sharing, Admin, LAN — each with a dedicated UI visual (upload queue, share dialog, member/quota table, P2P presence)
4. **Capabilities** — numbered grid of the remaining product surface
5. **Comparison** (`#compare`) — Dropbox / MinIO / FileShare table
6. **How it works** (`#how`) — three install steps + command strip
7. **Stack** (`#stack`) — backend / frontend / clients
8. **Closing** — final CTAs
9. **Footer**

Design: the landing page renders inside `.light-site`, a scoped token override in `app/globals.css` that flips the Tailwind theme vars to a light palette (near-white void, ink text, IBM blue `#0f62fe` accent) while the `/getting-started/*` docs keep the original dark editorial theme (void background, warm accent, serif display headings). Mock UI surfaces are built from the same token vars, so they follow the scope. No aurora blobs, glass cards, or gradient CTAs.

## Internationalization (EN / AR)

The site ships English (default, untranslated URLs) and Arabic under the `/ar/` URL prefix, chosen over `next-intl` because `output: "export"` makes a simple route-group split the least invasive fit:

- **Routing** — two root layouts via route groups: `app/(en)/…` renders `/…`, `app/(ar)/…` renders `/ar/…`. Each sets `<html lang>` + `dir` (`en`/`ltr`, `ar`/`rtl`). No middleware or `i18n` config (Next 15 rejects those in static export).
- **Landing** — one shared component (`components/landing.tsx`) takes a `locale` prop; copy comes from the typed dictionaries in `lib/i18n/{en,ar}.ts` (`ar.ts` is checked against the `Dict` type so keys can't drift). `lp(locale, path)` prefixes internal links for `/ar`.
- **Docs** — the five getting-started pages are **twin files** per locale (`app/(en)/getting-started/**`, `app/(ar)/ar/getting-started/**`). Code samples and technical identifiers stay in English in both; layout is `dir="rtl"` with mirrored back/next arrows.
- **Components** — `Header`, `Footer`, `SiteShell`, `DocsShell`, `Step`, and `Landing` accept `locale` (default `"en"`). There is no React context: the export build rejects `createContext` in server components, so locale flows as explicit props. `components/language-switcher.tsx` (client) derives the current locale from `usePathname()` and mirrors the path across `/ar`.
- **RTL styling** — logic-flipped utilities (`text-start`, `border-s/e`, `ms/me`) plus `html[lang="ar"]` rules in `globals.css`: GE SS Text (copied from `frontend/public/fonts/`) for body/headings/labels, `letter-spacing: normal` overrides for `.eyebrow`/`tracking-*`, and `html[dir="rtl"] code/pre` forced LTR (`direction: ltr; unicode-bidi: isolate`).
- No `hreflang` alternates are emitted — there is no canonical origin env var (and thus no `metadataBase`) to build absolute URLs from.

Verify both locales after changes: `out/index.html` must stay `lang="en" dir="ltr"` with English copy, and every page under `out/ar/**` must be `lang="ar" dir="rtl"` with `/ar/…` internal links.

## Static hosting

The build writes a fully static site to `out/`:

```bash
yarn build
# serve locally
npx serve out
# or
yarn start
```

- **Any static host:** upload `out/`.
- Trailing slashes are enabled (`trailingSlash: true`) for static hosts that expect directory URLs.
- Images are not optimized (`images.unoptimized: true`) so the export works without an image optimizer.

## Project layout

```
website/
├── app/
│   ├── (en)/                 # English root layout — / , /getting-started/**
│   ├── (ar)/ar/              # Arabic root layout (lang=ar, dir=rtl) — /ar/**
│   └── globals.css           # Tailwind + theme + .light-site + RTL rules
├── components/
│   ├── landing.tsx           # Shared locale-driven landing page
│   ├── site.tsx              # Header/Footer/SiteShell/DocsShell/Step/Cards…
│   └── language-switcher.tsx # EN ⇄ AR toggle (client, path-mirroring)
├── lib/
│   ├── i18n/{en,ar,index}.ts # Typed dictionaries + Locale/getDict/lp
│   └── site.ts               # APP_URL / GITHUB_URL constants
├── public/
│   ├── fonts/                # GE SS Text (Arabic body font)
│   └── favicon.svg
├── next.config.ts            # Static export + outputFileTracingRoot
├── eslint.config.mjs
├── postcss.config.mjs
├── tsconfig.json
└── package.json
```

## Development notes

- Port **3001** avoids clashing with Vite (`5173`) and the API (`3000`).
- `outputFileTracingRoot` is set to this directory so Next does not pick a parent lockfile.
- `next-env.d.ts` and `.next/` / `out/` are gitignored (root `.gitignore` also ignores `.next` and `out`).
- No backend calls — marketing copy only; CTAs are plain links.

## Verify before commit

```bash
cd website
yarn build
yarn tsc --noEmit
yarn lint
```
