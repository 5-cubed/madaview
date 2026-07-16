# Architecture: Add react-icons dependency for VSCode-style icons

**ADR:** `.context/adr/react-icons.md`

## Static View
> This change is manifest-only — no classes, no Objects/Logics/Usecase/External layers, no directory structure change. The only artifact is a dependency declaration.

**Directory structure**
```
web/
  package.json        # +1 line: "react-icons": "^5.7.0" in dependencies
  package-lock.json    # +10 lines: react-icons@5.7.0 lock entry, no transitive deps
```

**Classes**
None. No source file imports `react-icons` yet.

**Dependencies**
```mermaid
graph LR
    pkg["web/package.json · dependencies"] --> ri["react-icons@5.7.0"]
    ri -.->|peerDependencies: react *| react["react@19.2.7"]
```

## Dynamic View
> The requirements spec's User Scenario ("Developer implements the sidebar toggle") describes future usage of this dependency, not code delivered by this change. This change stops at making the package installable; the import/build/lint flow below has not run against real code yet — it is the contract the sidebar-folding change must satisfy.

### Developer implements the sidebar toggle (not yet implemented — contract only)
```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Pkg as web/package.json
    participant Src as Sidebar.tsx (future)
    participant Build as tsc -b / vite build
    participant Lint as oxlint

    Note over Pkg: This ADR's actual scope ends here
    Dev->>Pkg: react-icons@5.7.0 already declared
    Note over Src,Lint: Deferred to sidebar-folding change
    Dev->>Src: import { VscLayoutSidebarLeft, VscLayoutSidebarLeftOff } from 'react-icons/vsc'
    Src->>Build: tsc -b && vite build
    Build-->>Dev: type-checks, bundles only the two imported icons
    Src->>Lint: oxlint
    Lint-->>Dev: passes, no new errors
```
