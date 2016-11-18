package main

import (
	"flag"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"io"
	"os/exec"
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

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		excludeProcesses string
		excludePorts     string
		user             string
		i                string

		version bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.StringVar(&excludeProcesses, "exclude-processes", "", "")

	flags.StringVar(&excludePorts, "exclude-ports", "", "")

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

	exProcesses := string.Split(",", excludeProcesses)
	exPorts := string.Split(",", excludePorts)
	_ = user
	_ = i

	// graph
	graphAst, _ := gographviz.Parse([]byte(`digraph G {}`))
	graph := gographviz.NewGraph()
	gographviz.Analyse(graphAst, graph)

	// queue作成
	var queue []string
	// set 作成
	hostMap := make(map[string]bool)

	// for queueが空になるまで
	for len(queue) < 1 {
		host = queue[0]
		queue = queue[1:]
		os = checkOS(host)
		hosts := netstat(os, host, exProcesses, exPorts, graph)
		for _, host := range hosts {
			_, ok := hostMap[host]
			if !ok {
				queue = append(set, host)
				hostMap[host] = true
			}
		}
	}

	return ExitCodeOK
}

func checkOS(host string) string {
	os := "Ubuntu"
	out, err := exec.Command("ssh", "-i", "~/.ssh/isucon-aws.pem", host, "uname", "-a").Output()
	uname := string(out)
	if strings.Contains(uname, "amzn") {
		os = "Amazon Linux AMI"
	} else if strings.Contains(uname, "debian") {
		os = "debian"
	}
	return os
}

func netstat(os, host string, exProcesses, exPorts []string, graph *gographviz.Graph) []string {
	var ret []string
	netstatOption := ""

	switch os {
	case "Amazon Linux AMI":
		option = "-tp"
	case "Ubuntu":
		option = "-tpW"
	case "Ubuntu":
		option = "-tp"
	}

	out, err := exec.Command("ssh", "-i", "~/.ssh/isucon-aws.pem", host, "netstat", netstatOption, "--numeric-ports").Output()

	if err != nil {
		return ExitCodeError
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		l := strings.Fields(line)
		if len(l) > 6 && strings.Contains(l[5], "ESTABLISHED") {
			// 接続している側か逆か
			send := false
			if l[6] == "-" {
				host = string.Split(":", l[3])[0]
			} else {
				host = string.Split(":", l[4])[0]
				send = true
			}
			flag := false
			// exclude-processesだったらcontinue
			for _, exProcess := range exProcesses {
				if strings.Contains(l[6], exProcess) {
					flag = true
					break
				}
			}
			// exclude-portsだったらcontinue
			for _, exPort := range exPorts {
				if strings.Contains(string.Split(":", l[4])[1], exPort) {
					flag = true
					break
				}
			}
			if flag {
				continue
			}
			fmt.Println(l)
			// l[6]が"-"でなければ矢印
			if send {
				graph.AddEdge(origin, host, true, nil)
			}
			// queueに追加するhosts
			ret = append(ret, host)
		}
	}

	return ret
}
