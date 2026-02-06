import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/auth': 'http://localhost:8080',
      '/users': 'http://localhost:8080',
      '/passes': 'http://localhost:8080',
      '/guest-requests': 'http://localhost:8080',
      '/openapi.yaml': 'http://localhost:8080',
      '/docs': 'http://localhost:8080'
    }
  }
});
