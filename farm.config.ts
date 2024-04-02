import { defineConfig } from '@farmfe/core'
import { join } from 'path'
import { cwd } from 'process'

// https://www.farmfe.org/docs/config/configuring-farm
export default defineConfig({
  clearScreen: false,
  compilation: {
    persistentCache: true,
    define: {
      'import.meta.env.FARM_IS_DEMO': process.env.FARM_IS_DEMO
    },
    resolve: {
      alias: {
        '/@': join(cwd(), 'client/src'),
      }
    },
    input: {
      index: './client/index.html'
    },
    output: {
      path: 'static',
      publicPath: '/cosmos-ui/'
    },
  },
  server: {
    proxy: {
      '/cosmos/api': {
        target: 'https://localhost:8443'
      }
    }
  },
  plugins: [
    ['@farmfe/plugin-react', { runtime: 'automatic' }]
  ],
})
