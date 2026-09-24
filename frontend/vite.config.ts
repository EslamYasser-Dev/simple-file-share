import { defineConfig, loadEnv, type ProxyOptions } from 'vite';
import react from '@vitejs/plugin-react-swc';
import tailwindcss from '@tailwindcss/vite';
import { resolve } from 'path';
import { visualizer } from 'rollup-plugin-visualizer';

// https://vite.dev/config/

const isTTY = Boolean(process.stderr.isTTY);
const code = (n: number) => String.fromCharCode(27) + '[' + n + 'm';
const dim = (s: string) => (isTTY ? code(2) + s + code(0) : s);
const yellow = (s: string) => (isTTY ? code(33) + s + code(0) : s);
const red = (s: string) => (isTTY ? code(31) + s + code(0) : s);

let proxyWarned = false;

function withProxyErrorHandling(opts: ProxyOptions): ProxyOptions {
  return {
    ...opts,
    configure(proxy) {
      proxy.on('error', (err, req, res) => {
        if (!proxyWarned) {
          proxyWarned = true;
          const target = process.env.VITE_API_URL || 'http://localhost:3000';
          console.warn(
            yellow('[vite]') +
              ' ' +
              dim('backend unreachable at') +
              ' ' +
              red(target) +
              ' ' +
              dim('- start the API, or set VITE_API_URL'),
          );
          setTimeout(() => {
            proxyWarned = false;
          }, 10_000);
        }
        // http-proxy may hand us a ServerResponse; sockets (upgrade) lack end().
        if (res && typeof (res as { end?: unknown }).end === 'function') {
          const anyRes = res as {
            writableEnded?: boolean;
            headersSent?: boolean;
            writeHead?: (code: number, headers: Record<string, string>) => void;
            end: (body?: string) => void;
          };
          if (!anyRes.writableEnded) {
            const body = JSON.stringify({
              error: 'backend_unavailable',
              message: 'API server is not running. Start the backend on port 3000.',
              path: req?.url,
            });
            if (!anyRes.headersSent && anyRes.writeHead) {
              anyRes.writeHead(502, { 'Content-Type': 'application/json' });
            }
            anyRes.end(body);
          }
        }
        void err;
      });
    },
  };
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const target = env.VITE_API_URL || 'http://localhost:3000';
  // Sub-path the app is served from. GitHub project pages live under
  // /<repo>/, so the deploy script/workflow sets VITE_BASE_PATH accordingly.
  const base = env.VITE_BASE_PATH || '/';

  return {
    base,
    plugins: [
      react(),
      tailwindcss(),
      mode === 'analyze' &&
        visualizer({
          open: true,
          filename: 'dist/stats.html',
          gzipSize: true,
          brotliSize: true,
        }),
    ].filter(Boolean),

    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },

    server: {
      port: 5173,
      proxy: {
        // API endpoints
        '/api': withProxyErrorHandling({
          target,
          changeOrigin: true,
          secure: false,
          // Do not rewrite; backend expects paths to start with /api
        }),
        // Upload endpoint
        '/upload': withProxyErrorHandling({
          target,
          changeOrigin: true,
          secure: false,
          // Rewrite to backend's /api/upload endpoint
          rewrite: (path) => path.replace(/^\/upload/, '/api/upload'),
        }),
      },
    },
  };
});
