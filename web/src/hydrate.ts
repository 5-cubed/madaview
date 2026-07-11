import mermaid from 'mermaid'
import katex from 'katex'
import '@plantuml/core/viz-global.js'
import { renderToString as plantumlRenderToString } from '@plantuml/core'

// Mermaid's SVG text has its own font config independent of page CSS, so it
// won't inherit the body font-stack automatically — read the resolved
// value at runtime so the CSS stack (index.css) stays the single source of
// truth and the two can't drift apart.
mermaid.initialize({ startOnLoad: false, fontFamily: getComputedStyle(document.body).fontFamily })

let mermaidCounter = 0

// Scans a rendered markdown container for the mermaid/katex-math/plantuml
// placeholders emitted by internal/markdown (see the ADR's rendering
// pipeline section) and renders them in place. All three run entirely
// client-side: no network call is made for any of them.
export async function hydrate(container: HTMLElement) {
  await Promise.all([hydrateMermaid(container), hydrateKatex(container), hydratePlantuml(container)])
}

async function hydrateMermaid(container: HTMLElement) {
  const blocks = container.querySelectorAll<HTMLElement>('div.mermaid')
  for (const block of blocks) {
    const source = block.textContent ?? ''
    if (!source.trim()) continue
    const id = `mermaid-${mermaidCounter++}`
    try {
      const { svg } = await mermaid.render(id, source)
      block.innerHTML = svg
    } catch (err) {
      block.textContent = `Mermaid render error: ${err instanceof Error ? err.message : String(err)}`
    }
  }
}

async function hydrateKatex(container: HTMLElement) {
  const nodes = container.querySelectorAll<HTMLElement>('.katex-math')
  for (const node of nodes) {
    const source = node.textContent ?? ''
    const displayMode = node.dataset.display === 'true'
    try {
      katex.render(source, node, { displayMode, throwOnError: false })
    } catch (err) {
      node.textContent = `KaTeX render error: ${err instanceof Error ? err.message : String(err)}`
    }
  }
}

// PlantUML's engine always returns *some* SVG, even for invalid input —
// the error is baked into the image as text rather than surfaced via
// onError (empirically confirmed; see the ADR). These two patterns cover
// both native error-card templates PlantUML produces.
const PLANTUML_ERROR_PATTERNS = [
  /Diagram not supported by this release of PlantUML/,
  /\(Assumed diagram type:/,
]

function renderPlantuml(lines: string[]): Promise<string> {
  return new Promise((resolve, reject) => {
    plantumlRenderToString(lines, resolve, reject)
  })
}

// Detects PlantUML's own native error card inside a resolved SVG and
// extracts a one-line message from it. Returns null when the SVG is a
// genuine diagram (no error-card signature found).
function extractPlantumlError(svg: string): string | null {
  if (!PLANTUML_ERROR_PATTERNS.some((pattern) => pattern.test(svg))) return null
  const doc = new DOMParser().parseFromString(svg, 'image/svg+xml')
  for (const node of doc.querySelectorAll('text')) {
    const text = (node.textContent ?? '').trim()
    if (text && PLANTUML_ERROR_PATTERNS.some((pattern) => pattern.test(text))) {
      return text
    }
  }
  return 'invalid PlantUML syntax'
}

async function hydratePlantuml(container: HTMLElement) {
  const blocks = container.querySelectorAll<HTMLElement>('div.plantuml')
  for (const block of blocks) {
    const source = block.textContent ?? ''
    const trimmed = source.trim()
    if (!trimmed) continue
    const wrapped = trimmed.startsWith('@start') ? source : `@startuml\n${source}\n@enduml`
    try {
      const svg = await renderPlantuml(wrapped.split('\n'))
      const error = extractPlantumlError(svg)
      if (error) {
        block.textContent = `PlantUML render error: ${error}`
      } else {
        block.innerHTML = svg
      }
    } catch (err) {
      block.textContent = `PlantUML render error: ${err instanceof Error ? err.message : String(err)}`
    }
  }
}
