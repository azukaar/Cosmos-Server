import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  root: 'client',
  build: {
    outDir: '../static',
  },
  server: {
    proxy: {
      '/cosmos/api': {
        target: 'http://192.168.1.170:8080',
        secure: false,
        ws: true,
      }
    }
  }
})
