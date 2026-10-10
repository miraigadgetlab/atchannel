# atchannel frontend

Vue 3 single-page client for the atchannel Go backend. It talks to the same
`/api/v1` surface documented in the repository's `docs.md`.

## Stack

- **Vue 3** (`<script setup>` + TypeScript)
- **Vite** for the dev server and builds
- **Vue Router** for routing, with auth/role guards
- **Pinia** for state (auth session, notification badge)
- **Tailwind CSS v4** (via the `@tailwindcss/vite` plugin)

## Getting started

```bash
npm install
npm run dev     # http://localhost:5173
```

The dev server proxies `/api` and `/uploads` to the backend (default
`http://localhost:3000`), so the browser only ever talks to one origin. Point it
somewhere else with `VITE_PROXY_TARGET` in `.env.local` (see `.env.example`).

## Scripts

| Script               | What it does                                    |
| -------------------- | ----------------------------------------------- |
| `npm run dev`        | Start the dev server with HMR                   |
| `npm run build`      | Type-check and build for production into `dist/` |
| `npm run preview`    | Serve the production build locally              |
| `npm run type-check` | Run `vue-tsc` over the project                  |
| `npm run lint`       | Lint with oxlint and ESLint                     |
| `npm run format`     | Format `src/` with Prettier                     |

## How it fits together

- `src/api/` — one module per backend resource plus `client.ts`, a `fetch`
  wrapper that attaches the access token and transparently refreshes it once on
  a `401` (rotating refresh tokens, matching the backend).
- `src/stores/auth.ts` — the single source of truth for "who is signed in". The
  user record is always re-fetched from `GET /me`, because roles and ban state
  can change server-side.
- `src/router/index.ts` — route table with `requiresAuth`, `guestOnly` and
  `roles` meta, enforced by a global guard.
- `src/types/api.ts` — hand-written mirror of the Go DTOs.
- `src/views/` and `src/components/` — the UI.

## Environment

See `.env.example`. All variables have working defaults, so an empty `.env`
works for local development.
