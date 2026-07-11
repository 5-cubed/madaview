// Package rootfs resolves and guards access to a single root folder that
// bounds all filesystem operations exposed over HTTP.
package rootfs

import "os"

const (
	SourceCLI     = "cli"
	SourceUI      = "ui"
	SourceDefault = "default"
)

// ResolveInitial determines the initial root path and its source, applying
// priority: CLI arg > persisted UI setting > current working directory.
func ResolveInitial(cliArg, persisted string) (path, source string, err error) {
	if cliArg != "" {
		return cliArg, SourceCLI, nil
	}
	if persisted != "" {
		return persisted, SourceUI, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	return cwd, SourceDefault, nil
}
