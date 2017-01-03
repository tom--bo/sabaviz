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
	exProcesses     []string
	exPorts         []string
	hostCheck       []string
	user            string
	sshKey          string
	hostThreshold   int
	connectionLimit int
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		excludeProcesses string
		excludePorts     string
		hostCheck        string
		user             string
		hostThreshold    int
		connectionLimit  int
		i                string
		version          bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.StringVar(&excludeProcesses, "exclude-processes", "", "comma separated exclude processes like ssh,ldap,syslog")
	flags.StringVar(&excludePorts, "exclude-ports", "", "comma separated exclude processes lile 22,53,389")
	flags.StringVar(&hostCheck, "host-check", "", "specify domain name to filter only local domain host")

	flags.StringVar(&user, "user", "", "ssh user")
	flags.StringVar(&user, "u", "", "(Short of user option)")

	flags.IntVar(&hostThreshold, "test", -1, "Limit of host to stat(not node count in graph)\n\t")
	flags.IntVar(&connectionLimit, "max", 20, "Limit of connection per one host.\n\tThe host which has connections more than this limit is ignored.\n\t")

	flags.StringVar(&i, "ssh-key", "", "")
	flags.StringVar(&i, "i", "", "(Short of ssh-key option)")

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
	conf.hostThreshold = hostThreshold
	conf.connectionLimit = connectionLimit

	firstHost := flags.Args()[0]
	sabaviz := &Sabaviz{outStream: cli.outStream, errStream: cli.errStream, conf: conf}
	sabaviz.main(firstHost)

	return ExitCodeOK
}
