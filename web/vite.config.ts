import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

// https://vitejs.dev/config/
export default defineConfig({
  base: '/dashboard/',
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 19190,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:9878',
        changeOrigin: true,
      },
      '/v1': {
        target: 'http://localhost:9878',
        changeOrigin: true,
      },
      '/admin': {
        target: 'http://localhost:9878',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
