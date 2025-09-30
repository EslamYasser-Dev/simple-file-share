import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react-swc';
import tailwindcss from '@tailwindcss/vite';
import { resolve } from 'path';
import { visualizer } from 'rollup-plugin-visualizer';

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const target = env.VITE_API_URL || 'http://localhost:3000';
  
  return {
    plugins: [
      react(),
      tailwindcss(),
      mode === 'analyze' && visualizer({
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
        '/api': {
          target,
          changeOrigin: true,
          secure: false,
          // Do not rewrite; backend expects paths to start with /api
        },
        // Upload endpoint
        '/upload': {
          target,
          changeOrigin: true,
          secure: false,
          // Rewrite to backend's /api/upload endpoint
          rewrite: (path) => path.replace(/^\/upload/, '/api/upload'),
        },
      },
    },
  }
})
