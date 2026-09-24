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
} from "../../../components/site";

export const metadata: Metadata = {
  title: "First steps",
  description:
    "Sign in, upload files, create shares, manage users, and use LAN Share.",
};

export default function FirstStepsPage() {
  return (
    <DocsShell
      eyebrow="Getting started / First steps"
      title="First steps in the app"
      description="Once the stack is running — locally or in Docker — these are the core workflows."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">On this page</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>
              <a href="#sign-in">Sign in</a>
            </li>
            <li>
              <a href="#upload">Upload & resume</a>
            </li>
            <li>
              <a href="#share">Share links</a>
            </li>
            <li>
              <a href="#team">Team & roles</a>
            </li>
            <li>
              <a href="#lan">LAN Share</a>
            </li>
            <li>
              <a href="#language">Language</a>
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
      <section id="sign-in">
        <ProseH2>1. Sign in</ProseH2>
        <UL
          items={[
            <>
              <strong>Auth disabled (dev):</strong> you land as the system
              admin view — no login form.
            </>,
            <>
              <strong>Auth enabled:</strong> use the bootstrap{" "}
              <code>ADMIN_USERNAME</code>/<code>ADMIN_PASSWORD</code>, or
              register if <code>ENABLE_SIGNUP=true</code> (first account
              becomes admin).
            </>,
          ]}
        />
        <Callout>
          Self-signup is gated by <code>ENABLE_SIGNUP</code>. Disable it after
          onboarding if you only want invited users.
        </Callout>
      </section>

      <section id="upload">
        <ProseH2>2. Upload & resume</ProseH2>
        <Step n="1" title="Upload">
          <P>
            Use <strong>Upload</strong> in the sidebar or drop files on the
            file list. Progress shows speed and percentage; multi-file batches
            aggregate into one indicator.
          </P>
        </Step>
        <Step n="2" title="Pause or cancel">
          <P>
            While an upload runs, hit <strong>Pause</strong> in the toast.
            Chunks stop between sends; <strong>Resume</strong> continues from
            the staged offset. Cancel aborts and clears the session.
          </P>
        </Step>
        <Step n="3" title="Recover after reload">
          <P>
            Interrupted resumable jobs appear under <strong>Pending uploads</strong>.
            Resume re-opens the file picker and validates path, name, and size
            before continuing.
          </P>
          <Callout>
            Files above the resumable threshold (1&nbsp;MB) use chunked
            sessions. Smaller files upload in one shot.
          </Callout>
        </Step>
      </section>

      <section id="share">
        <ProseH2>3. Share links</ProseH2>
        <UL
          items={[
            "Open a file or folder → Share",
            "Copy the link (token-based URL)",
            "Revoke the share when done",
            "Shared folder: everyone signed-in can read; writes are admin-only",
          ]}
        />
        <P>
          Folders download as a ZIP on the fly — no temp archive on disk.
        </P>
      </section>

      <section id="team">
        <ProseH2>4. Team & roles</ProseH2>
        <UL
          items={[
            <>
              <strong>Admin console</strong> — user list, usage, quotas
            </>,
            <>
              Roles: <code>admin</code>, <code>member</code>, plus custom
              roles with a permission matrix
            </>,
            "Per-user storage quotas and password resets",
            "Private home directories; other users’ namespaces return 403",
          ]}
        />
        <P>Typical invite flow:</P>
        <CodeBlock
          code={`1. Enable ENABLE_SIGNUP (or create users as admin)
2. Share the app URL
3. New users register → land in their private home
4. Assign roles / quotas from the admin console`}
        />
      </section>

      <section id="lan">
        <ProseH2>5. LAN Share (peer-to-peer)</ProseH2>
        <UL
          items={[
            "Open LAN Share in the sidebar",
            "Both devices must reach the same server (same LAN / VPN)",
            "Select a peer → send a file; bytes go device-to-device over WebRTC",
            "Server only brokers SDP/ICE signaling — not file contents",
          ]}
        />
        <Callout>
          Turn LAN Share off anytime in <strong>Settings → LAN Share</strong>.
          The nav item hides and the signaling stream closes.
        </Callout>
      </section>

      <section id="language">
        <ProseH2>6. Language & theme</ProseH2>
        <UL
          items={[
            "Toggle EN / AR from the header (full RTL for Arabic)",
            "Preference persists in the browser",
            "Settings: light / dark / system theme and max parallel uploads",
          ]}
        />
      </section>

      <ProseH2>Explore next</ProseH2>
      <UL
        items={[
          <>
            <Link href="/getting-started/deploy/">Deploy to production</Link>
          </>,
          <>
            API reference in the repo README and{" "}
            <code>/swagger</code> on a running server
          </>,
          <>
            <Link href="/getting-started/local/">Local development</Link> for
            contributing
          </>,
        ]}
      />
    </DocsShell>
  );
}
