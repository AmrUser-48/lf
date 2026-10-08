package main

import (
	"fmt"
	"os"
)

type internalExpr struct {
	name string
}

func (e *internalExpr) String() string { return "internal " + e.name }

func (e *internalExpr) eval(app *app, args []string) {
	if len(args) == 0 {
		app.ui.echoerrf("internal %s: missing file", e.name)
		return
	}
	switch e.name {
	case "vi":
		app.runInternalSync(func() error { return runBuiltinVi(args[0]) })
	case "less":
		app.runInternalSync(func() error { return runBuiltinLess(args[0]) })
	default:
		app.ui.echoerrf("unknown internal command: %s", e.name)
	}
}

func (app *app) runInternalSync(fn func() error) {
	app.nav.previewChan <- ""
	if err := app.ui.suspend(); err != nil {
		logInternal("suspend: %v", err)
	}
	defer func() {
		if err := app.ui.resume(); err != nil {
			app.quit()
			os.Exit(3)
		}
		app.ui.loadFile(app, true)
		app.nav.renew()
	}()
	if err := fn(); err != nil {
		app.ui.echoerrf("%v", err)
	}
}

func logInternal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
