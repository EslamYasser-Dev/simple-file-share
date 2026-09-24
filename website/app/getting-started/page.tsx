import type { Metadata } from "next";
import Link from "next/link";
import {
  Callout,
  CodeBlock,
  DocsShell,
  P,
  ProseH2,
  UL,
} from "../../components/site";
import { APP_URL, GITHUB_URL } from "../../lib/site";

export const metadata: Metadata = {
  title: "Getting started",
  description:
    "Install FileShare locally, with Docker, or deploy it — then take your first steps in the app.",
};

const guides = [
  {
    href: "/getting-started/local/",
    title: "Local development",
    body: "Run the Go API and React UI on your machine with hot reload.",
    tag: "Developers",
  },
  {
    href: "/getting-started/docker/",
    title: "Docker",
    body: "One command with Compose, or build and run a single container.",
    tag: "Ops",
  },
  {
    href: "/getting-started/first-steps/",
    title: "First steps",
    body: "Sign in, upload files, share links, and invite your team.",
    tag: "Product",
  },
  {
    href: "/getting-started/deploy/",
    title: "Deploy",
    body: "All-in-one Docker host or split GitHub Pages + API.",
    tag: "Production",
  },
] as const;

export default function GettingStartedPage() {
  return (
    <DocsShell
      eyebrow="Getting started"
      title="Ship FileShare in minutes"
      description="Pick a path: develop locally, run Docker, or go straight to production. Each guide is short and copy-paste ready."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">Prerequisites</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>Go 1.25+</li>
            <li>Node.js 20+</li>
            <li>Docker (optional)</li>
            <li>Git</li>
          </ul>
          <a
            href={`${GITHUB_URL}#readme`}
            target="_blank"
            rel="noreferrer"
            className="mt-5 inline-flex text-sm text-accent transition hover:text-paper"
          >
            Full README →
          </a>
        </div>
      }
    >
      <div className="grid gap-4 sm:grid-cols-2">
        {guides.map((g) => (
          <Link
            key={g.href}
            href={g.href}
            className="panel group p-6 transition hover:border-accent/40"
          >
            <span className="font-mono text-[10px] uppercase tracking-wider text-muted">
              {g.tag}
            </span>
            <h2 className="mt-3 font-display text-lg font-semibold text-paper">
              {g.title}
            </h2>
            <p className="mt-2 text-sm leading-relaxed text-muted">
              {g.body}
            </p>
            <span className="mt-4 inline-block text-sm text-accent transition group-hover:text-paper">
              Open guide →
            </span>
          </Link>
        ))}
      </div>

      <ProseH2>Quick path</ProseH2>
      <P>
        Short on time? Use Docker Compose from the repo root — it builds the
        UI and API together and exposes HTTP on port{" "}
        <code>22010</code>.
      </P>
      <CodeBlock
        label="Terminal"
        code={`git clone ${GITHUB_URL}.git
cd simple-file-share
export JWT_SECRET="$(openssl rand -hex 32)"
export ADMIN_PASSWORD='your-strong-password'
docker compose up --build
# open http://localhost:22010`}
      />
      <Callout>
        Compose requires <code>JWT_SECRET</code> (random, 32+ characters) and{" "}
        <code>ADMIN_PASSWORD</code> — there is no default login, and signup is
        off unless you set <code>ENABLE_SIGNUP=true</code>. See{" "}
        <Link href="/getting-started/docker/">the Docker guide</Link>.
      </Callout>

      <ProseH2>What you get</ProseH2>
      <UL
        items={[
          "Single Go binary (or container) serving API + React UI",
          "Resumable uploads, pause/resume, and pending-job recovery",
          "LAN peer-to-peer transfer with WebRTC signaling",
          "Roles, quotas, share links, and an admin console",
          "English/Arabic UI with full RTL",
          "Local disk or S3-compatible storage",
        ]}
      />

      <ProseH2>Next after install</ProseH2>
      <UL
        items={[
          <>
            Follow{" "}
            <Link href="/getting-started/first-steps/">First steps</Link> to
            upload, share, and invite users.
          </>,
          <>
            Point the mobile app with{" "}
            <code>EXPO_PUBLIC_API_URL</code> (see the main README).
          </>,
          <>
            Open the app at <a href={APP_URL}>{APP_URL}</a> once your stack is
            running.
          </>,
        ]}
      />
    </DocsShell>
  );
}
