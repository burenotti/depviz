package main

import (
	"context"
	"depviz/internal/app"
	"flag"
	"fmt"
	"os"
)

func main() {
	var packageNamePip string
	var packageNameNpm string
	var packageManager string

	flag.StringVar(&packageNamePip, app.Pip, "", "fetch dependency graph of package from pip")
	flag.StringVar(&packageNameNpm, app.Npm, "", "fetch dependency graph of package from npm")
	flag.Parse()

	if packageNameNpm != "" && packageNamePip != "" {
		exitWithMessage("You may specify only one package manager")
	} else if packageNameNpm != "" {
		packageManager = app.Npm
	} else if packageNamePip != "" {
		packageManager = app.Pip
	}

	c := &app.Config{
		PackageName:    packageNamePip + packageNameNpm,
		PackageManager: packageManager,
	}
	ctx := context.Background()

	if err := app.Run(ctx, c); err != nil {
		exitWithMessage(err.Error())
	}
}

func exitWithMessage(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	flag.PrintDefaults()
	os.Exit(1)
}
