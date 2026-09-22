# FileShare — Frontend

React 19 + TypeScript + Vite single-page app for the Simple File Share server.
The production build is embedded into and served by the Go backend.

## Stack

- **React 19** with hooks and `useActionState` form actions
- **Vite 7** (SWC) for dev server and builds
- **TypeScript** (strict)
- **Tailwind CSS v4** for styling
- **Zustand** for global state (auth, files, toasts)
- **lucide-react** for icons
- **Custom i18n** (`src/i18n`) with English + Arabic and RTL support

## Development

The dev server proxies `/api` to the backend (default `http://localhost:3000`).

```bash
npm install
npm run dev      # http://localhost:5173
```

Override the proxy target with `VITE_API_URL` in `.env` if the backend runs
elsewhere. Leave it empty to use the same origin.

## Scripts

| Command | Description |
| --- | --- |
| `npm run dev` | Start the Vite dev server |
| `npm run build` | Type-check (`tsc -b`) and build to `dist/` |
| `npm run preview` | Preview the production build locally |
| `npm run lint` | Run ESLint |
| `npx vite build --mode analyze` | Build with the bundle visualizer |

## Configuration

Build-time variables (see `src/config/index.ts`):

- `VITE_API_URL` — API base URL; empty uses the current origin.
- `VITE_MAX_UPLOAD_MB` — client-side upload size limit in MB; `0` = unlimited.
- `VITE_BASE_PATH` — sub-path the app is served from; defaults to `/`. GitHub
  project pages use `/<repo>/` (set automatically by the Pages deploy).

## Project structure

```
src/
├── components/   # Layout, Modal, Toast, FilePreview, FileIcon
├── config/       # API base URL and upload limits
├── hooks/        # useToast
├── i18n/         # language provider and translations
├── lib/          # shared utilities (formatting)
├── pages/        # home, summary, chat, login, register, admin
├── services/     # typed API client
├── store/        # Zustand stores (auth, files, toasts)
└── App.tsx       # routing and auth gate
```
