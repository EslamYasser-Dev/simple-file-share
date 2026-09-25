import type { Metadata } from "next";
import Link from "next/link";
import {
  Callout,
  CodeBlock,
  DocsShell,
  P,
  ProseH2,
  Step,
  UL,
} from "../../../../components/site";
import { GITHUB_URL } from "../../../../lib/site";

export const metadata: Metadata = {
  title: "Local development",
  description:
    "Run the FileShare Go API and React frontend locally with hot reload.",
};

export default function LocalDevPage() {
  return (
    <DocsShell
      eyebrow="Getting started / Local"
      title="Local development"
      description="Two processes: the Go API on :3000 and Vite on :5173. Proxying is built into the frontend config."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">Ports</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>
              <code>3000</code> — HTTP API
            </li>
            <li>
              <code>5173</code> — Vite UI
            </li>
            <li>
              <code>50051</code> — gRPC (optional)
            </li>
          </ul>
          <div className="mt-5 space-y-2 text-sm">
            <Link
              href="/getting-started/"
              className="block text-accent transition hover:text-paper"
            >
              ← All guides
            </Link>
            <Link
              href="/getting-started/docker/"
              className="block text-accent transition hover:text-paper"
            >
              Next: Docker →
            </Link>
          </div>
        </div>
      }
    >
      <ProseH2>Prerequisites</ProseH2>
      <UL
        items={[
          "Go 1.25 or later",
          "Node.js 20+ and npm",
          "Git",
        ]}
      />

      <Step n="1" title="Clone the repository">
        <CodeBlock
          code={`git clone ${GITHUB_URL}.git
cd simple-file-share`}
        />
      </Step>

      <Step n="2" title="Start API + UI together">
        <P>
          From the repo root, the Makefile runs the backend and Vite in one
          terminal:
        </P>
        <CodeBlock label="Terminal A" code={`make dev`} />
        <P>
          Or start them separately if you prefer two terminals:
        </P>
        <CodeBlock
          label="Terminal A — API"
          code={`cd backend
go mod download
go run ./cmd/server`}
        />
        <CodeBlock
          label="Terminal B — UI"
          code={`cd frontend
npm install
npm run dev`}
        />
      </Step>

      <Step n="3" title="Configure (optional)">
        <P>Development defaults work out of the box (auth off, local disk). Override as needed:</P>
        <CodeBlock
          label="env"
          code={`PORT=3000
ROOT_DIR=.file-share-data          # or FILE_SHARE_ROOT
ENABLE_AUTH=false                  # true requires login
ENABLE_SIGNUP=true
ENABLE_TLS=false
APP_ENV=development
ADMIN_USERNAME=admin               # seeded only on first run
ADMIN_PASSWORD=admin`}
        />
        <Callout tone="warn">
          Prefer <code>ADMIN_USERNAME</code>/<code>ADMIN_PASSWORD</code> over
          legacy <code>USERNAME</code>/<code>PASSWORD</code> — the shell may
          already define <code>USERNAME</code>.
        </Callout>
      </Step>

      <Step n="4" title="Open the app">
        <UL
          items={[
            <>
              UI: <code>http://localhost:5173</code>
            </>,
            <>
              API health: <code>http://localhost:3000/health</code>
            </>,
            <>
              Swagger: <code>http://localhost:3000/swagger</code>
            </>,
          ]}
        />
        <Callout>
          Vite proxies <code>/api/*</code> to <code>:3000</code>. If you see
          “backend unreachable”, start the Go server first.
        </Callout>
      </Step>

      <ProseH2>Useful Make targets</ProseH2>
      <CodeBlock
        code={`make run          # backend only
make dev          # backend + frontend
make test-backend # go test -race
make lint         # fmt + vet + frontend lint
make frontend-build`}
      />

      <ProseH2>Mobile (optional)</ProseH2>
      <CodeBlock
        code={`cd mobile
npm install
echo 'EXPO_PUBLIC_API_URL=http://10.0.2.2:3000' > .env.local  # Android emulator
npx expo start
npx tsc --noEmit && npx expo lint`}
      />
      <P>
        Next: <Link href="/getting-started/docker/">Docker setup</Link> or{" "}
        <Link href="/getting-started/first-steps/">first steps in the app</Link>.
      </P>
    </DocsShell>
  );
}
