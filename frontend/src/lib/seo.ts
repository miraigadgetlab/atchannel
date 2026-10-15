const BASE = 'atchannel'
const SITE_URL = (import.meta.env.VITE_SITE_URL as string | undefined) ?? 'https://atchannel.top'

export function getBaseUrl(): string {
  if (typeof window !== 'undefined' && window.location.origin) {
    return window.location.origin
  }
  return SITE_URL
}

function setMeta(name: string, content: string): void {
  if (typeof document === 'undefined') return
  let el = document.querySelector(`meta[name="${name}"]`) as HTMLMetaElement | null
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute('name', name)
    document.head.appendChild(el)
  }
  el.setAttribute('content', content)
}

function setProperty(property: string, content: string): void {
  if (typeof document === 'undefined') return
  let el = document.querySelector(`meta[property="${property}"]`) as HTMLMetaElement | null
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute('property', property)
    document.head.appendChild(el)
  }
  el.setAttribute('content', content)
}

function setLink(rel: string, href: string): void {
  if (typeof document === 'undefined') return
  let el = document.querySelector(`link[rel="${rel}"]`) as HTMLLinkElement | null
  if (!el) {
    el = document.createElement('link')
    el.setAttribute('rel', rel)
    document.head.appendChild(el)
  }
  el.setAttribute('href', href)
}

export interface SeoOptions {
  title?: string
  description?: string
  type?: 'website' | 'article' | 'profile'
  url?: string
  image?: string
  noindex?: boolean
  jsonLd?: Record<string, unknown> | null
}

const DEFAULT_DESCRIPTION =
  'atchannel — a minimalist community for focused discussions. Browse channels, read posts, join the conversation.'

/**
 * Single writer for everything a search engine reads on the page: tab title,
 * meta description, robots directive, Open Graph / Twitter cards, the
 * canonical link and optional JSON-LD structured data. Because this is a SPA,
 * Google sees only this first HTML document unless these tags are rewritten
 * on every navigation — which is exactly what the router afterEach does.
 */
export function setSeo(options: SeoOptions = {}): void {
  if (typeof document === 'undefined') return

  const pageTitle = options.title ? `${options.title} · ${BASE}` : BASE
  document.title = pageTitle

  const description = options.description ?? DEFAULT_DESCRIPTION
  setMeta('description', description)
  setMeta(
    'robots',
    options.noindex
      ? 'noindex,nofollow'
      : 'index,follow,max-image-preview:large,max-snippet:-1,max-video-preview:-1',
  )

  const url = options.url ?? (typeof window !== 'undefined' ? window.location.href : SITE_URL)
  const image = options.image ?? `${getBaseUrl()}/og-image.png`

  setProperty('og:site_name', BASE)
  setProperty('og:type', options.type ?? 'website')
  setProperty('og:title', pageTitle)
  setProperty('og:description', description)
  setProperty('og:url', url)
  setProperty('og:image', image)

  setMeta('twitter:card', 'summary_large_image')
  setMeta('twitter:title', pageTitle)
  setMeta('twitter:description', description)
  setMeta('twitter:image', image)

  setLink('canonical', url)

  document.getElementById('seo-jsonld')?.remove()
  if (options.jsonLd) {
    const script = document.createElement('script')
    script.type = 'application/ld+json'
    script.id = 'seo-jsonld'
    script.textContent = JSON.stringify(options.jsonLd)
    document.head.appendChild(script)
  }
}
