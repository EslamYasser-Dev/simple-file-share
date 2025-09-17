import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react-swc'
import tailwindcss from '@tailwindcss/vite'


// https://vite.dev/config/
export default defineConfig(({ mode }: { mode: string }) => {
  const env = loadEnv(mode, '', '')
  const target = env.VITE_BASE_URL || 'https://127.0.0.1:22010'
  return {
    plugins: [react(),
      tailwindcss()
    ],
    server: {
      proxy: {
        '/upload': {
          target,
          changeOrigin: true,
          secure: false,
        },
      },
    },
  }
})
