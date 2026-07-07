import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath } from 'url'
import { dirname, resolve } from 'path'
import tanstackRouter from '@tanstack/router-plugin/vite'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    // The TanStack Router plugin must come before the JSX transform plugin.
    tanstackRouter({ target: 'react', autoCodeSplitting: true }),
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        // Vite 8 bundles with Rolldown, which replaces the object form of
        // manualChunks with advancedChunks groups matched by module id.
        advancedChunks: {
          groups: [
            { name: 'react', test: /node_modules[\\/](react|react-dom)[\\/]/ },
            { name: 'router', test: /node_modules[\\/]@tanstack[\\/]react-router/ },
            { name: 'aggrid', test: /node_modules[\\/]ag-grid-/ },
            { name: 'icons', test: /node_modules[\\/](@tabler[\\/]icons-react|@heroicons[\\/]react)[\\/]/ },
            { name: 'monaco', test: /node_modules[\\/](@monaco-editor[\\/]react|monaco-editor)[\\/]/ },
          ],
        },
      },
    },
  },
})
