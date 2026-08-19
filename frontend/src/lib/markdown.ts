import { marked } from 'marked'
import hljs from 'highlight.js/lib/core'
import c from 'highlight.js/lib/languages/c'
import cpp from 'highlight.js/lib/languages/cpp'
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import go from 'highlight.js/lib/languages/go'
import rust from 'highlight.js/lib/languages/rust'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'
import diff from 'highlight.js/lib/languages/diff'
import sql from 'highlight.js/lib/languages/sql'
import xml from 'highlight.js/lib/languages/xml'

hljs.registerLanguage('c', c)
hljs.registerLanguage('cpp', cpp)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('js', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('ts', typescript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('go', go)
hljs.registerLanguage('rust', rust)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('sh', bash)
hljs.registerLanguage('json', json)
hljs.registerLanguage('diff', diff)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('xml', xml)

marked.use({
  breaks: true,
  gfm: true,
  renderer: {
    image() {
      return ''
    },
    html({ text }: { text: string }) {
      return text.replace(/<[^>]*>/g, '')
    },
    code({ text, lang }: { text: string; lang?: string }) {
      const language = lang && hljs.getLanguage(lang) ? lang : ''
      const highlighted = language
        ? hljs.highlight(text, { language }).value
        : hljs.highlightAuto(text).value
      const cls = language ? ` class="language-${language}"` : ''
      return `<pre><code${cls}>${highlighted}</code></pre>`
    },
    codespan({ text }: { text: string }) {
      const escaped = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      return `<code>${escaped}</code>`
    },
  },
})

function sanitizeLinks(html: string): string {
  return html.replace(
    /<a\s+href="([^"]*)"/gi,
    (_, href: string) => {
      try {
        const url = new URL(href)
        if (url.protocol !== 'http:' && url.protocol !== 'https:') {
          return '<a href="#"'
        }
        return `<a href="${url.href}" target="_blank" rel="noopener noreferrer nofollow"`
      } catch {
        return '<a href="#"'
      }
    }
  )
}

export function renderMarkdown(src: string): string {
  let html = marked.parse(src) as string
  html = html.replace(/<img\b[^>]*>/gi, '')
  html = sanitizeLinks(html)
  return html
}
