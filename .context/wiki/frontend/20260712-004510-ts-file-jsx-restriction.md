# `.ts` files in web/ cannot contain JSX syntax

`web/tsconfig.app.json` sets `"jsx": "react-jsx"` but TypeScript only
parses JSX syntax (`<Foo>...</Foo>`) inside `.tsx` files — a `.ts` file
using JSX fails to compile even with the same compiler options. If a
design explicitly calls for a `.ts` (not `.tsx`) file that still needs to
render/return a React element (e.g. a provider component colocated with
plain state/reducer logic in one file for cohesion), use
`React.createElement(...)` directly instead of JSX syntax in that one
spot — everything else in the file can stay ordinary TypeScript.

Example: `web/src/useWorkspace.ts` (types + reducer + `WorkspaceProvider`
+ `useWorkspace()` in one file per its ADR) uses
`createElement(WorkspaceContext.Provider, { value }, children)` for the
provider's single render, rather than splitting into a separate `.tsx`
file just for that one JSX expression.
