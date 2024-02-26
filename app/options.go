package app

import (
	"io"
	"log/slog"

	actx "github.com/friendlycaptcha/fcov/app/context"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

// Option is a function that allows configuring the application.
type Option func(*App)

// WithArgs sets the command arguments passed to the CLI parser.
func WithArgs(args []string) Option {
	return func(app *App) {
		app.args = args
	}
}

// WithEnv sets the process environment used by the application.
func WithEnv(env actx.Environment) Option {
	return func(app *App) {
		app.ctx.Env = env
	}
}

// WithExit sets the function that stops the application.
func WithExit(fn func(int)) Option {
	return func(app *App) {
		app.Exit = fn
	}
}

// WithFDs sets the file descriptors used by the application.
func WithFDs(stdin io.Reader, stdout, stderr io.Writer) Option {
	return func(app *App) {
		app.ctx.Stdin = stdin
		app.ctx.Stdout = stdout
		app.ctx.Stderr = stderr
	}
}

// WithFS sets the filesystem used by the application.
func WithFS(fs vfs.FileSystem) Option {
	return func(app *App) {
		app.ctx.FS = fs
	}
}

// WithLogger sets the logger used by the application.
func WithLogger(logger *slog.Logger) Option {
	return func(app *App) {
		app.ctx.Logger = logger
	}
}
