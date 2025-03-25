import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

import dotenv from 'dotenv';
dotenv.config();

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/kubernetes': {
        target: process.env.VITE_KUBERNETES_ENDPOINT,
        changeOrigin: true,
        secure: false,
        ws: true,
        rewrite: path => path.replace('/kubernetes', ''),
      },
      '/autopilot': {
        target: process.env.VITE_AUTOPILOT_ENDPOINT,
        changeOrigin: true,
        secure: false,
        ws: true,
        rewrite: path => path.replace('/autopilot', ''),
      }
    },
  },
});
