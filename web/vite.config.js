import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8887',
        changeOrigin: true,
        secure: false
      },
      '/v1/social': {
        target: 'http://localhost:8889',
        changeOrigin: true,
        secure: false
      }
    }
  }
})
