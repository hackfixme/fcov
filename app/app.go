package app

import (
	"log/slog"

	"github.com/mandelsoft/vfs/pkg/memoryfs"

	"go.hackfix.me/fcov/app/cli"
	actx "go.hackfix.me/fcov/app/context"
)

// App is the application.
type App struct {
	ctx  *actx.Context
	cli  *cli.CLI
	args []string

	Exit func(int)
}

// New initializes a new application.
func New(opts ...Option) *App {
	defaultCtx := &actx.Context{
		FS:     memoryfs.New(),
		Logger: slog.Default(),
	}
	app := &App{ctx: defaultCtx, Exit: func(int) {}}

	for _, opt := range opts {
		opt(app)
	}

	slog.SetDefault(app.ctx.Logger)

	var err error
	if app.cli, err = cli.New(app.ctx, app.args, app.Exit); err != nil {
		app.FatalIfErrorf(err)
	}

	return app
}

// Run starts execution of the application.
func (app *App) Run() {
	err := app.cli.Execute(app.ctx)
	app.FatalIfErrorf(err)
}

// FatalIfErrorf terminates the application with an error message if err != nil.
func (app *App) FatalIfErrorf(err error, args ...interface{}) {
	if err != nil {
		app.ctx.Logger.Error(err.Error(), args...)
		app.Exit(1)
	}
}
