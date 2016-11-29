package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

// Config object
type Config struct {
	exProcesses []string
	exPorts     []string
	hostCheck   []string
	user        string
	sshKey      string
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		excludeProcesses string
		excludePorts     string
		hostCheck        string
		user             string
		i                string
		version          bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.StringVar(&excludeProcesses, "exclude-processes", "", "")
	flags.StringVar(&excludePorts, "exclude-ports", "", "")
	flags.StringVar(&hostCheck, "host-check", "", "")

	flags.StringVar(&user, "user", "", "")
	flags.StringVar(&user, "u", "", "(Short)")

	flags.StringVar(&i, "ssh-key", "", "")
	flags.StringVar(&i, "i", "", "(Short)")

	flags.BoolVar(&version, "version", false, "Print version information and quit.")
	flags.BoolVar(&version, "v", false, "(Short)")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeError
	}

	// Show version
	if version {
		fmt.Fprintf(cli.errStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	var conf Config
	conf.exProcesses = strings.Split(excludeProcesses, ",")
	conf.exPorts = strings.Split(excludePorts, ",")
	conf.hostCheck = strings.Split(hostCheck, ",")
	conf.user = user
	conf.sshKey = i

	firstHost := flags.Args()[0]
	sabaviz := &Sabaviz{outStream: cli.outStream, errStream: cli.errStream, conf: conf}
	sabaviz.main(firstHost)

	return ExitCodeOK
}
