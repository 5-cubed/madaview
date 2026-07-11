// Command madaview is a local HTTP server that renders a browsable,
// read-only view of markdown files rooted at a chosen folder.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/5-cubed/madaview/internal/assets"
	"github.com/5-cubed/madaview/internal/config"
	"github.com/5-cubed/madaview/internal/rootfs"
	"github.com/5-cubed/madaview/internal/server"
)

// version is overridden via -ldflags at release build time.
var version = "dev"

func main() {
	var (
		rootFlag string
		port     int
		verbose  bool
	)
	flag.StringVar(&rootFlag, "root", "", "root folder to serve (default: last-used root, or the current working directory)")
	flag.IntVar(&port, "port", 4800, "port to listen on")
	flag.BoolVar(&verbose, "v", false, "enable verbose (debug) logging")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose (debug) logging")
	flag.Parse()

	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	if err := run(rootFlag, port, logger); err != nil {
		fmt.Fprintln(os.Stderr, "madaview:", err)
		os.Exit(1)
	}
}

func run(rootFlag string, port int, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		logger.Warn("failed to load persisted config", "error", err.Error())
	}

	initialPath, source, err := rootfs.ResolveInitial(rootFlag, cfg.Root)
	if err != nil {
		return fmt.Errorf("resolving initial root: %w", err)
	}

	root, err := rootfs.New(initialPath)
	if err != nil {
		return fmt.Errorf("opening root %q: %w", initialPath, err)
	}
	logger.Info("root resolved", "root", root.Current(), "source", source)

	srv := server.New(server.Options{
		Root:       root,
		RootSource: source,
		Version:    version,
		Assets:     assets.FS,
		Logger:     logger,
		PersistRoot: func(path string) error {
			return config.Save(config.Config{Root: path})
		},
	})

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	logger.Info("starting server", "addr", addr, "version", version)
	return http.ListenAndServe(addr, srv.Handler())
}
