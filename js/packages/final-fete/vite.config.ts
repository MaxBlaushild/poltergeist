import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// https://vite.dev/config/
export default defineConfig({
  base: '/',
  plugins: [react()],
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
