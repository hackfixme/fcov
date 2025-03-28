package context

import (
	"context"
	"io"
	"log/slog"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

// Context contains common objects used by the application. It is passed around
// the application to avoid direct dependencies on external systems, and make
// testing easier.
type Context struct {
	Ctx    context.Context // global context
	FS     vfs.FileSystem  // filesystem
	Env    Environment     // process environment
	Logger *slog.Logger    // global logger

	// Standard streams
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Metadata
	Version *VersionInfo
}

// Environment is the interface to the process environment.
type Environment interface {
	Get(string) string
	Set(string, string) error
}
