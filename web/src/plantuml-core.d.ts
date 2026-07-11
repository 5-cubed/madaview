// @plantuml/core ships no type declarations (verified: no .d.ts in the
// published package). Ambient module covering only the two entry points
// this project uses.
declare module '@plantuml/core' {
  export function render(lines: string[], targetId: string, opts?: Record<string, unknown>): void
  export function renderToString(
    lines: string[],
    onSuccess: (svg: string) => void,
    onError: (message: string) => void,
  ): void
}

declare module '@plantuml/core/viz-global.js'
