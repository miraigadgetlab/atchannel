const BASE = 'atchannel'

/**
 * Sets the browser tab title as "<page> · atchannel". The router calls this for
 * every navigation using the route's static label; content views (a post, a
 * channel, a profile) call it again once their data loads so the tab names the
 * thing actually being shown. With no page it falls back to the bare brand.
 */
export function setTitle(page?: string): void {
  document.title = page ? `${page} · ${BASE}` : BASE
}
