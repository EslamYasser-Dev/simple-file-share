export type ChipTone = "neutral" | "accent";

export interface MockRow {
  kind: "dir" | "file";
  name: string;
  size: string;
  when: string;
  chip: { text: string; tone: ChipTone } | null;
}

export const en = {
  meta: {
    title: "FileShare — self-hosted file sharing in one Go binary",
    description:
      "Open-source file workspace: resumable uploads, LAN P2P, roles and quotas, S3 or local disk, English and Arabic. Built with Go, React, and Flutter.",
  },
  nav: {
    features: "Features",
    compare: "Compare",
    install: "Install",
    source: "Source",
    openApp: "Open app",
    home: "Home",
  },
  docs: {
    step: "Step",
    breadcrumb: "Breadcrumb",
  },
  footer: {
    tagline: "self-hosted",
    features: "Features",
    install: "Install",
    app: "App",
  },
  hero: {
    eyebrow: "Open source · self-hosted · Go + React",
    titleA: "Keep the files.",
    titleB: "Skip the subscription.",
    sub: "FileShare is a complete file workspace you run yourself: resumable uploads, share links, roles and quotas, LAN peer-to-peer — with a web and mobile UI. One Go binary, no vendor in the middle.",
    openApp: "Open the app",
    install: "Install guide",
    source: "Source on GitHub",
    strip: "One binary · Local disk or S3 · EN / AR + RTL · REST + gRPC · Mobile app",
  },
  mock: {
    nav: {
      myFiles: "My files",
      shared: "Shared with me",
      recent: "Recent",
      trash: "Trash",
      lan: "LAN Share",
    },
    storage: "Storage",
    storageValue: "4.2 / 20 GB",
    breadcrumb: { myFiles: "My files", projects: "Projects", current: "Q3" },
    newFolder: "New folder",
    upload: "Upload",
    cols: { name: "Name", size: "Size", modified: "Modified" },
    rows: [
      {
        kind: "dir",
        name: "Contracts",
        size: "12 items",
        when: "Today",
        chip: null,
      },
      {
        kind: "file",
        name: "Q3-report.pdf",
        size: "2.4 MB",
        when: "Today",
        chip: { text: "Shared", tone: "accent" },
      },
      {
        kind: "file",
        name: "brand-assets.zip",
        size: "118 MB",
        when: "Mon",
        chip: null,
      },
      {
        kind: "file",
        name: "demo-recording.mp4",
        size: "640 MB",
        when: "Fri",
        chip: { text: "Large", tone: "neutral" },
      },
      {
        kind: "file",
        name: "notes.md",
        size: "12 KB",
        when: "Fri",
        chip: null,
      },
    ] as MockRow[],
    card: {
      uploading: "Uploading 1 of 3",
      fileLine: "Q3-report.pdf · 41.3 of 66 MB",
      pause: "Pause",
      cancel: "Cancel",
      resumeNote: "resumes after reload",
    },
  },
  rows: [
    {
      eyebrow: "Uploads",
      title: "Transfers that survive a closed tab",
      body: "Big files leave the browser in chunks, not one giant request. Pause, reload, switch networks — the server-side offset picks up where it stopped instead of starting over.",
      points: [
        "Pause and resume from the queue, including after a refresh",
        "Per-user quotas with a visible meter before you hit the cap",
        "Chunked sessions keep flaky connections from costing an hour",
      ],
    },
    {
      eyebrow: "Sharing",
      title: "Links with an expiry and a kill switch",
      body: "Any file or folder can become a public link with an expiration, an optional password, and a download counter. Revoke it and the token dies immediately — no CDN cache to wait out.",
      points: [
        "Cryptographically random tokens, rate-limited against guessing",
        "Expiry windows and revocation stored with the share, not the file",
        "Everything else stays behind login and role policy",
      ],
    },
    {
      eyebrow: "Admin",
      title: "Roles and quotas an admin can actually read",
      body: "Members see what policy allows; admins get a member table with usage bars, per-user quotas, and rollups of uploads, downloads, and shares — without exporting anything.",
      points: [
        "Built-in admin and member roles, plus custom roles to extend",
        "Per-user storage caps enforced at upload time",
        "Analytics events logged locally as JSONL, never phoned home",
      ],
    },
    {
      eyebrow: "LAN",
      title: "Same network, no cloud hop",
      body: "On a shared network, peers talk directly over a WebRTC data channel. The server only swaps signaling messages — the bytes never leave the building.",
      points: [
        "Presence, SDP, and ICE relayed over a rate-limited endpoint",
        "Works while the internet is down, as long as the LAN is up",
        "Falls back to normal server upload when peers cannot connect",
      ],
    },
  ],
  capabilities: {
    eyebrow: "Also in the box",
    title: "The rest of the product surface",
    sub: "Not a roadmap — these are wired up and covered by tests today.",
    items: [
      {
        n: "01",
        title: "TLS 1.3 and careful file paths",
        body: "PBKDF2-SHA256 passwords, path-traversal checks on every open, rate limits and CSP headers on by default.",
      },
      {
        n: "02",
        title: "Local disk or S3",
        body: "Same port, different adapter. A folder on the box or MinIO/Ceph/AWS-compatible credentials.",
      },
      {
        n: "03",
        title: "REST and gRPC side by side",
        body: "OpenAPI for the web app and scripts; protobuf services when you want streaming or another language.",
      },
      {
        n: "04",
        title: "Versions and restore",
        body: "Every save keeps a snapshot. Roll back or pull an older copy without leaving the browser.",
      },
      {
        n: "05",
        title: "Admin analytics",
        body: "Uploads, downloads, shares, and storage rollups per user — logged locally, readable from the admin console.",
      },
      {
        n: "06",
        title: "Web, mobile, static docs",
        body: "React app inside the binary, Flutter client for phones, and this site plus install guides as a static export.",
      },
    ],
  },
  compare: {
    eyebrow: "Compare",
    title: "Dropbox, MinIO, and the gap between",
    sub: "People usually arrive asking which one this is. Short answer: neither — it is the product layer that sits on storage you already run.",
    dimension: "Dimension",
    rows: [
      {
        dim: "Where it runs",
        dropbox: "Their cloud",
        minio: "Your servers (infra)",
        fileshare: "Your servers (app)",
      },
      {
        dim: "What it is",
        dropbox: "SaaS sync product",
        minio: "Object storage API",
        fileshare: "Team file workspace",
      },
      {
        dim: "UI on day one",
        dropbox: "Full client",
        minio: "Admin console",
        fileshare: "Web + mobile app",
      },
      {
        dim: "LAN peer-to-peer",
        dropbox: "No",
        minio: "No",
        fileshare: "WebRTC data channel",
      },
      {
        dim: "Resume after reload",
        dropbox: "Yes",
        minio: "SDK concern",
        fileshare: "Built into the UI",
      },
      {
        dim: "Source",
        dropbox: "Closed",
        minio: "Open (AGPL)",
        fileshare: "Open, yours to fork",
      },
    ],
  },
  how: {
    eyebrow: "Install",
    title: "Three steps, then you are done",
    steps: [
      {
        n: "1",
        title: "Run the binary",
        body: "docker compose up, or download the binary. One process serves the API and the UI on :22010.",
      },
      {
        n: "2",
        title: "Create an admin",
        body: "Bootstrap credentials come from env on first boot, then invite the rest of the team.",
      },
      {
        n: "3",
        title: "Drop files in",
        body: "Drag, resume, share a link, or open LAN Share when everyone is on the same Wi-Fi.",
      },
    ],
    cmdNote: "API + UI in one container",
    guide: "Full install guide →",
  },
  stack: {
    eyebrow: "Stack",
    title: "Boring tools, on purpose",
    sub: "Nothing here needs a plugin marketplace to keep running in five years.",
    cols: [
      {
        label: "Backend",
        items: [
          "Go 1.25, stdlib HTTP + gRPC",
          "Hexagonal ports & adapters",
          "PBKDF2-SHA256 · JWT",
          "Filesystem and S3 adapters",
          "go test -race suite",
        ],
      },
      {
        label: "Frontend",
        items: [
          "React 19 · TypeScript · Vite",
          "Zustand · Tailwind",
          "Resumable chunk queue",
          "SSE for live events",
          "EN / AR with RTL",
        ],
      },
      {
        label: "Clients",
        items: [
          "Flutter mobile app",
          "OpenAPI + protobuf",
          "Docker single container",
          "Static marketing export",
          "Static-host-ready UI",
        ],
      },
    ],
  },
  closing: {
    eyebrow: "Next",
    title: "Run it once and decide for yourself",
    sub: "Clone the repo, docker run the image, or read the install notes first. Happy to take issues and patches on GitHub.",
    openApp: "Open the app",
    install: "Install guide",
  },
  visuals: {
    uploads: {
      title: "Upload queue",
      count: "3 files · 826 MB",
      done: "Done",
      queued: "Queued",
      pause: "Pause",
      cancel: "Cancel",
      meta1: "41.3 of 66 MB",
      footer: "Close the tab mid-transfer — the server keeps the offset",
    },
    share: {
      title: "Share “Q3-report.pdf”",
      linkChip: "Public link",
      copy: "Copy",
      expires: "Expires",
      expiresIn: "In 7 days",
      password: "Password",
      off: "Off",
      downloads: "Downloads",
      allowCount: "Allow · count shown",
      revoke: "Revoke link",
      done: "Done",
    },
    admin: {
      title: "Members",
      quota: "quota 20 GB each",
      admin: "Admin",
      member: "Member",
      invite: "Invite member",
      footer: "roles · quotas · analytics",
      people: [
        { initials: "AK", name: "Amina K.", role: "Admin", used: "62%" },
        { initials: "YS", name: "Youssef S.", role: "Member", used: "18%" },
        { initials: "ML", name: "Mariam L.", role: "Member", used: "4%" },
      ],
    },
    lan: {
      title: "LAN Share",
      direct: "Direct",
      thisLaptop: "This laptop",
      sender: "sender",
      receiver: "receiver",
      phoneName: "Nour’s iPhone",
      footer: "Server swaps signaling only — the bytes stay on the LAN",
    },
  },
};

export type Dict = typeof en;
