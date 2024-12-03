import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import visualizer from 'rollup-plugin-visualizer';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  root: 'client',
  build: {
    outDir: '../static',
    rollupOptions: {
      plugins: [visualizer({ open: true })],
    },
  },
  server: {
    proxy: {
      '/cosmos/api': {
        // target: 'https://localhost:8443',
        target: 'http://192.168.1.170:8080',
        secure: false,
        ws: true,
      },
      '/cosmos/rclone': {
        // target: 'https://localhost:8443',
        target: 'http://192.168.1.170:8080',
        secure: false,
        ws: true,
      }
    }
  }
})
