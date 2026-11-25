import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// The Go API listens on :3000 (see ../.env). We run the UI on :3001 and proxy
// /api so the browser sees a single origin: the auth cookie just works and no
// CORS setup is needed in development.
const API = process.env.IRONLOBBY_API ?? 'http://localhost:3000'

export default defineConfig({
  plugins: [svelte()],
  server: { port: 3001, proxy: { '/api': API } },
  preview: { port: 3001, proxy: { '/api': API } },
})
