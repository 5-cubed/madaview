.PHONY: build frontend test

# Full release-style build: frontend, then copy it into the go:embed
# target, then the Go binary. This is the only target that requires npm; a
# bare `go build`/`go vet`/`go test` works on a fresh clone without it,
# because internal/assets/dist ships with a checked-in placeholder.
#
# go:embed reads index.html's content at compile time, so it's already
# baked into the binary above by the time the checkout below runs — the
# restore only keeps the working tree clean; it doesn't touch what got
# embedded. Vite's content-hashed filenames mean a real build's index.html
# never stays byte-stable across builds anyway, so this is what keeps the
# committed placeholder from drifting every time someone builds locally.
build: frontend
	go build -o madaview ./cmd/madaview
	git checkout -- internal/assets/dist/index.html

frontend:
	cd web && npm ci && npm run build
	rm -rf internal/assets/dist
	cp -r web/dist internal/assets/dist

test:
	go test ./...
