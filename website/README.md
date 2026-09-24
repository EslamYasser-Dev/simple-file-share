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
npm install
npm run dev
```

Open [http://localhost:3001](http://localhost:3001).

## Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Dev server on port **3001** |
| `npm run build` | Production static export → `out/` |
| `npm run start` | Serve the production build (port 3001) |
| `npm run lint` | ESLint (`next/core-web-vitals` + TypeScript) |

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

These are inlined at **build time** — set them before `npm run build`.

## Page sections

Defined in `app/page.tsx` (product screenshot-forward layout):

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

## Static hosting

The build writes a fully static site to `out/`:

```bash
npm run build
# serve locally
npx serve out
# or
npm run start
```

- **GitHub Pages / any static host:** upload `out/` (or wire `scripts/deploy-pages.sh` style deploy).
- Trailing slashes are enabled (`trailingSlash: true`) for static hosts that expect directory URLs.
- Images are not optimized (`images.unoptimized: true`) so the export works without an image optimizer.

## Project layout

```
website/
├── app/
│   ├── layout.tsx      # Root layout, metadata, fonts/base styles
│   ├── page.tsx        # Landing page (all sections)
│   └── globals.css     # Tailwind + theme + utilities
├── public/
│   └── favicon.svg
├── next.config.ts      # Static export + outputFileTracingRoot
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
npm run build
npx tsc --noEmit
npm run lint
```
