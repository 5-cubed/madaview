.PHONY: build frontend test

# Full release-style build: frontend, then copy it into the go:embed
# target, then the Go binary. This is the only target that requires npm; a
# bare `go build`/`go vet`/`go test` works on a fresh clone without it,
# because internal/assets/dist ships with a checked-in placeholder.
build: frontend
	go build -o madaview ./cmd/madaview

frontend:
	cd web && npm ci && npm run build
	rm -rf internal/assets/dist
	cp -r web/dist internal/assets/dist

test:
	go test ./...
