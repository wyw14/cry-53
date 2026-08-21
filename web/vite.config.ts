import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '127.0.0.1',
    strictPort: true,
    proxy: {
      '/api': { target: 'http://127.0.0.1:8080', changeOrigin: false },
      '/healthz': { target: 'http://127.0.0.1:8080', changeOrigin: false }
    }
  },
  build: { target: 'es2022', sourcemap: false, assetsDir: 'assets' },
  test: { environment: 'jsdom', include: ['src/**/*.test.ts'] }
})
