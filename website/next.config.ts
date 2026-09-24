import type { NextConfig } from "next";
import { fileURLToPath } from "node:url";
import path from "node:path";

const nextConfig: NextConfig = {
  output: "export",
  trailingSlash: true,
  outputFileTracingRoot: path.dirname(fileURLToPath(import.meta.url)),
  images: { unoptimized: true },
};

export default nextConfig;
