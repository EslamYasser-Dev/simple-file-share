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
  title: "Deploy",
  description:
    "Deploy FileShare as an all-in-one container or split static UI + hosted API.",
};

export default function DeployPage() {
  return (
    <DocsShell
      eyebrow="Getting started / Deploy"
      title="Deploy"
      description="Two shapes: one container that serves UI + API, or static UI on any static host with the API elsewhere."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">Shapes</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>
              <strong className="text-paper">All-in-one</strong>
              <br />
              Docker host, one container
            </li>
            <li>
              <strong className="text-paper">Split</strong>
              <br />
              Static UI + separate API
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
              href="/getting-started/first-steps/"
              className="block text-accent transition hover:text-paper"
            >
              ← First steps
            </Link>
          </div>
        </div>
      }
    >
      <ProseH2>All-in-one (recommended)</ProseH2>
      <P>
        Any Docker-capable host (VPS, Fly.io, Koyeb, etc.): build or pull the
        image and mount a volume for <code>/data</code>.
      </P>
      <CodeBlock
        label="docker run"
        code={`docker run -d --name file-share \\
  -p 22010:22010 \\
  -v file-share-data:/data \\
  -e APP_ENV=production \\
  -e JWT_SECRET="$(openssl rand -hex 32)" \\
  -e ADMIN_USERNAME=admin \\
  -e ADMIN_PASSWORD='your-strong-password' \\
  -e ENABLE_AUTH=true \\
  -e ENABLE_TLS=false \\
  ghcr.io/eslamyasser-dev/simple-file-share:latest`}
      />
      <Callout tone="warn">
        Put TLS in front (reverse proxy or <code>ENABLE_TLS=true</code>). Never
        send passwords over plain HTTP in production.
      </Callout>

      <Step n="1" title="Verify">
        <CodeBlock
          code={`curl -fsS https://your-host/health
# open https://your-host/ and sign in`}
        />
      </Step>

      <ProseH2>Split — static UI + API</ProseH2>
      <P>
        Any static host serves the React build; the Go API runs on a Docker
        host. CORS is already permissive on the server.
      </P>

      <Step n="1" title="Deploy the API">
        <P>Use the all-in-one flow above and note the public origin, e.g. https://api.example.com.</P>
      </Step>

      <Step n="2" title="Publish the UI">
        <CodeBlock
          label="build the UI"
          code={`cd frontend
VITE_API_URL=https://api.example.com yarn build
# upload frontend/dist/ to any static host (nginx, S3, Netlify, …)`}
        />
        <Callout>
          Bake <code>VITE_API_URL</code> into the build (empty = same-origin).
          Serve the API over HTTPS so Basic Auth credentials are never sent in
          the clear.
        </Callout>
      </Step>

      <ProseH2>This marketing site</ProseH2>
      <P>
        The landing pages in <code>website/</code> export statically:
      </P>
      <CodeBlock
        code={`cd website
npm install
NEXT_PUBLIC_APP_URL=https://app.example.com \\
NEXT_PUBLIC_GITHUB_URL=${GITHUB_URL} \\
npm run build
# upload website/out/ to any static host`}
      />

      <ProseH2>Local production smoke test</ProseH2>
      <CodeBlock
        code={`mkdir -p bin
(cd backend && go build -o ../bin/file-share ./cmd/server)
(cd frontend && npm run build)

APP_ENV=production PORT=8090 ROOT_DIR=./data \\
STATIC_DIR=./frontend/dist \\
JWT_SECRET="$(openssl rand -hex 32)" \\
ADMIN_USERNAME=admin ADMIN_PASSWORD='local-smoke-only' ENABLE_TLS=false \\
./bin/file-share
# UI + API on http://localhost:8090`}
      />

      <ProseH2>Checklist</ProseH2>
      <UL
        items={[
          "Random JWT_SECRET (32+ chars) — the server refuses to start in production without one",
          "Strong ADMIN_PASSWORD; known defaults are refused on the first-boot seed",
          "ENABLE_AUTH=true and HTTPS termination",
          "Persistent volume for ROOT_DIR (/data)",
          "MAX_UPLOAD_BYTES if you need a cap",
          "ENABLE_SIGNUP=false after onboarding (optional)",
          "Backups of the data volume / S3 bucket",
        ]}
      />
      <P>
        Full variable table: repository <code>readme.md</code> → Deployment →
        Environment variables.
      </P>
    </DocsShell>
  );
}
