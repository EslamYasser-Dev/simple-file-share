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
  title: "Docker",
  description: "Run FileShare with Docker Compose or a single container.",
};

export default function DockerPage() {
  return (
    <DocsShell
      eyebrow="Getting started / Docker"
      title="Docker"
      description="All-in-one image: React UI + Go API in one container. Named volume keeps files under /data."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">Defaults</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>
              HTTP <code>22010</code>
            </li>
            <li>
              gRPC <code>50051</code>
            </li>
            <li>
              Volume <code>file-data → /data</code>
            </li>
            <li>
              Compose requires <code>JWT_SECRET</code> +{" "}
              <code>ADMIN_PASSWORD</code>
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
              href="/getting-started/deploy/"
              className="block text-accent transition hover:text-paper"
            >
              Next: Deploy →
            </Link>
          </div>
        </div>
      }
    >
      <ProseH2>Option A — Docker Compose</ProseH2>
      <Step n="1" title="Provide production secrets">
        <Callout tone="warn">
          Compose <strong>requires</strong> <code>JWT_SECRET</code> (the server
          refuses to start in production without a random value of at least 32
          characters) and <code>ADMIN_PASSWORD</code> (no default login ships).
          Export them for the shell or put them in a <code>.env</code> file
          next to <code>docker-compose.yml</code> — Docker Compose reads it
          automatically.
        </Callout>
        <CodeBlock
          label=".env (next to docker-compose.yml)"
          code={`# openssl rand -hex 32
JWT_SECRET=paste-a-random-32-byte-hex-string-here
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-strong-password
ENABLE_SIGNUP=false`}
        />
      </Step>

      <Step n="2" title="Clone and start">
        <CodeBlock
          code={`git clone ${GITHUB_URL}.git
cd simple-file-share
docker compose up --build`}
        />
        <P>
          Open <code>http://localhost:22010</code>. Storage persists in the{" "}
          <code>file-data</code> volume.
        </P>
      </Step>

      <Step n="3" title="Useful commands">
        <CodeBlock
          code={`docker compose logs -f     # tail logs
docker compose down         # stop (volume kept)
docker compose down -v      # stop and delete volume`}
        />
      </Step>

      <ProseH2>Option B — Build and run</ProseH2>
      <CodeBlock
        code={`docker build -t simple-file-share .

docker run -d --name file-share \\
  -p 22010:22010 \\
  -v file-share-data:/data \\
  -e APP_ENV=production \\
  -e JWT_SECRET="$(openssl rand -hex 32)" \\
  -e ADMIN_USERNAME=admin \\
  -e ADMIN_PASSWORD='your-strong-password' \\
  simple-file-share`}
      />
      <Callout>
        Prefer a named volume over a bind mount. The server runs as{" "}
        <code>nobody</code> (uid 65534) and must own <code>/data</code> to lock
        it to <code>0700</code>. For <code>{`-v "$PWD/data:/data"`}</code>:
        <br />
        <code>mkdir -p data && sudo chown 65534:65534 data</code>
      </Callout>

      <ProseH2>Option C — Published image</ProseH2>
      <P>
        Releases publish to GitHub Container Registry when you tag{" "}
        <code>vX.Y.Z</code>:
      </P>
      <CodeBlock
        code={`docker pull ghcr.io/eslamyasser-dev/simple-file-share:latest
docker run -d --name file-share -p 22010:22010 \\
  -v file-share-data:/data \\
  -e APP_ENV=production \\
  -e JWT_SECRET="$(openssl rand -hex 32)" \\
  -e ADMIN_USERNAME=admin \\
  -e ADMIN_PASSWORD='your-strong-password' \\
  ghcr.io/eslamyasser-dev/simple-file-share:latest`}
      />

      <ProseH2>Health check</ProseH2>
      <CodeBlock
        code={`curl -fsS http://localhost:22010/health
# {"status":"healthy"}`}
      />

      <ProseH2>Key environment variables</ProseH2>
      <UL
        items={[
          <>
            <code>APP_ENV=production</code> — enables auth by default
          </>,
          <>
            <code>JWT_SECRET</code> — required in production, random 32+
            characters
          </>,
          <>
            <code>ADMIN_PASSWORD</code> — required for the first-run admin seed
          </>,
          <>
            <code>PORT</code>, <code>ROOT_DIR</code>, <code>STATIC_DIR</code>
          </>,
          <>
            <code>ENABLE_SIGNUP</code>, <code>ENABLE_TLS</code>,{" "}
            <code>MAX_UPLOAD_BYTES</code>
          </>,
          <>
            <code>STORAGE_BACKEND=local|s3</code> plus S3_* credentials
          </>,
        ]}
      />
      <P>
        Next: <Link href="/getting-started/deploy/">production deploy</Link> or{" "}
        <Link href="/getting-started/first-steps/">first steps</Link>.
      </P>
    </DocsShell>
  );
}
