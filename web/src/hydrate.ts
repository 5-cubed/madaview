import mermaid from 'mermaid'
import katex from 'katex'

// Mermaid's SVG text has its own font config independent of page CSS, so it
// won't inherit the body font-stack automatically — read the resolved
// value at runtime so the CSS stack (index.css) stays the single source of
// truth and the two can't drift apart.
mermaid.initialize({ startOnLoad: false, fontFamily: getComputedStyle(document.body).fontFamily })

let mermaidCounter = 0

// Scans a rendered markdown container for the mermaid/katex-math
// placeholders emitted by internal/markdown (see the ADR's rendering
// pipeline section) and renders them in place. Both libraries run entirely
// client-side: no network call is made for either.
export async function hydrate(container: HTMLElement) {
  await Promise.all([hydrateMermaid(container), hydrateKatex(container)])
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
