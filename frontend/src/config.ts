// Central place for build-time configuration. Vite only exposes variables
// prefixed with VITE_ to the client, and only those that were present at
// build time.
export const API_BASE_URL: string = import.meta.env.VITE_API_BASE_URL || '/api/v1'

// Where uploaded avatars are served from. The backend returns paths like
// "/uploads/abc.png"; when the API lives on another origin the caller may need
// to prefix them, so keep the base here too.
export const UPLOADS_BASE_URL: string = import.meta.env.VITE_UPLOADS_BASE_URL || ''
