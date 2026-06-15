import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// https://vite.dev/config/
export default defineConfig({
  base: '/',
  plugins: [react()],
  server: {
    // Allow tunneling the dev server (e.g. ngrok) for testing on real phones.
    // Leading-dot entries match any subdomain of that host.
    host: true,
    allowedHosts: ['.ngrok-free.dev', '.ngrok-free.app', '.ngrok.app', '.ngrok.io'],
  },
  resolve: {
    alias: {
      '@poltergeist/types': path.resolve(__dirname, '../types/src'),
    },
  },
  optimizeDeps: {
    exclude: ['@poltergeist/types'],
  },
  build: {
    assetsDir: 'assets',
  },
})
