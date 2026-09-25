import type { NextConfig } from "next";
import { fileURLToPath } from "node:url";
import path from "node:path";

// GitHub project Pages serve under /<repo>/ — set NEXT_PUBLIC_BASE_PATH in CI
// (defaults to no prefix for local dev and user/org pages).
const basePath = process.env.NEXT_PUBLIC_BASE_PATH || "";

const nextConfig: NextConfig = {
  output: "export",
  trailingSlash: true,
  ...(basePath ? { basePath } : {}),
  outputFileTracingRoot: path.dirname(fileURLToPath(import.meta.url)),
  images: { unoptimized: true },
};

export default nextConfig;
