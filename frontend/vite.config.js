import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: { 
    proxy: {
      // Redirigir solicitudes que comiencen con /api al backend en localhost:8080
      '/api': {
        target: 'http://localhost:8080', // El servidor de tu backend Go
        changeOrigin: true, // Necesario para vhosts
        // Opcional: reescribir la ruta si es necesario, aunque para /api -> /api no hace falta
        // rewrite: (path) => path.replace(/^\/api/, '/api') 
      }
    }
  }
})
