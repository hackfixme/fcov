package context

import (
	"io"
	"log/slog"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

// Context contains common objects used by the application. It is passed around
// the application to avoid direct dependencies on external systems, and make
// testing easier.
type Context struct {
	FS     vfs.FileSystem
	Env    Environment
	Logger *slog.Logger
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Environment is the interface to the process environment.
type Environment interface {
	Get(string) string
	Set(string, string) error
}
