// Central place for build-time configuration. Vite only exposes variables
// prefixed with VITE_ to the client, and only those that were present at
// build time.
export const API_BASE_URL: string = import.meta.env.VITE_API_BASE_URL || '/api/v1'

// Where uploaded avatars are served from. The backend returns paths like
// "/uploads/abc.png"; when the API lives on another origin the caller may need
// to prefix them, so keep the base here too.
export const UPLOADS_BASE_URL: string = import.meta.env.VITE_UPLOADS_BASE_URL || ''

// Social links shown, quietly, at the right of the footer. Discord is optional:
// leave VITE_DISCORD_URL unset and it is simply not rendered, so the row grows
// to "discord github" the moment an invite URL is provided.
export const GITHUB_URL: string =
  import.meta.env.VITE_GITHUB_URL || 'https://github.com/miraigadgetlab/atchannel'
export const DISCORD_URL: string = import.meta.env.VITE_DISCORD_URL || ''
