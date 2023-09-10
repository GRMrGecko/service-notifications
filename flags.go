package main

import (
	"flag"
	"fmt"
	"os"
)

// Flags supplied to cli.
type Flags struct {
	ConfigPath string
	HTTPBind   string
	HTTPPort   uint
	Update     bool
}

// Parse the supplied flags.
func (a *App) ParseFlags() {
	app.flags = new(Flags)
	flag.Usage = func() {
		fmt.Printf(serviceName + ": " + serviceDescription + ".\n\nUsage:\n")
		flag.PrintDefaults()
	}

	// If version is requested.
	var printVersion bool
	flag.BoolVar(&printVersion, "v", false, "Print version")

	// Override configuration path.
	usage := "Load configuration from `FILE`"
	flag.StringVar(&app.flags.ConfigPath, "config", "", usage)
	flag.StringVar(&app.flags.ConfigPath, "c", "", usage+" (shorthand)")

	// Config overrides for http configurations.
	flag.StringVar(&app.flags.HTTPBind, "http-bind", "", "Bind address for http server")
	flag.UintVar(&app.flags.HTTPPort, "http-port", 0, "Bind port for http server")

	// Runs database update for Slack and Planning Center information,
	// then it creates slack channels if needed.
	usage = "Update database and create channels"
	flag.BoolVar(&app.flags.Update, "update", false, usage)
	flag.BoolVar(&app.flags.Update, "u", false, usage+" (shorthand)")

	// Parse the flags.
	flag.Parse()

	// Print version and exit if requested.
	if printVersion {
		fmt.Println(serviceName + ": " + serviceVersion)
		os.Exit(0)
	}
}
