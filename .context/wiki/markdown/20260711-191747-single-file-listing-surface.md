# Single file-listing surface

`/api/tree` (`internal/server/tree.go` → `internal/rootfs/list.go` `List()`) is the only file-listing/directory-enumeration surface in the entire app. Confirmed by searching `web/src` and `internal` for `TreeEntry`, `fetchTree`, `handleTree`, `rootfs.List`, breadcrumb, and search patterns — no search feature, breadcrumb component, or recent-files list exists anywhere. Any change to directory-listing/filtering behavior only needs to touch this one path; no need to re-search for other consumers.
