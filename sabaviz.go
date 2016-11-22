package main

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

type Sabaviz struct {
	outStream, errStream io.Writer
	conf                 Config
}

type Host struct {
	hostName     string
	distribution string
	conns        []Connection
}

type Connection struct {
	hostName string
	port     string
}

func (s Sabaviz) main(firstHost string) {
	g := &Graph{}
	g.NewGraph()

	first := Host{hostName: firstHost}

	// hashmap for check host
	hostMap := make(map[string]bool)

	// queue作成
	var queue []Host
	queue = append(queue, first)
	hostMap[firstHost] = true

	// for queueが空になるまで
	for len(queue) > 0 {
		host := queue[0]
		queue = queue[1:]
		host.distribution = checkDistri(host.hostName)
		g.AddNode(host)

		// netstatでConnectionオブジェクトのスライスを返す
		// これは対象ホストとportでuniqになったものにしておく
		connections := netstat(host, s.conf)
		for _, conn := range connections {
			g.AddConnectionOnce(host, conn)
			_, ok := hostMap[conn.hostName]
			if !ok {
				queue = append(queue, Host{hostName: conn.hostName})
				hostMap[conn.hostName] = true
			}
		}
	}
	fmt.Println(g.graph.String())
}

func checkDistri(host string) string {
	distri := ""
	out, _ := exec.Command("ssh", host, "cat", "/etc/issue").Output()
	issue := string(out)
	if strings.Contains(issue, "Amazon Linux AMI") {
		distri = "AmazonLinuxAMI"
	} else if strings.Contains(issue, "Debian") {
		distri = "Debian"
	} else if strings.Contains(issue, "CentOS") {
		distri = "CentOS"
	} else if strings.Contains(issue, "Ubuntu") {
		distri = "Ubuntu"
	}
	return distri
}

func netstat(host Host, conf Config) []Connection {
	var ret []Connection
	connMap := make(map[Connection]bool)

	netstatOption := ""
	switch host.distribution {
	case "Amazon Linux AMI":
		netstatOption = "-atp"
	case "Ubuntu":
		netstatOption = "-atpW"
	case "Debian":
		netstatOption = "-atpW"
	case "CentOS":
		netstatOption = "-atpT"
	default:
		netstatOption = "-atp"
	}

	out, err := exec.Command("ssh", host.hostName, "netstat", netstatOption).Output()

	if err != nil {
		return nil
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		l := strings.Fields(line)
		if len(l) > 6 && strings.Contains(l[5], "ESTABLISHED") {
			if checkExcludePattern(conf, l) {
				// fmt.Println(l)
				conn := makeConnectionObj(host, l)
				_, ok := connMap[conn]
				if !ok {
					ret = append(ret, conn)
					connMap[conn] = true
				}
			}
		}
	}
	return ret
}

func makeConnectionObj(host Host, l []string) Connection {
	local := strings.Split(":", l[3])[0]
	localPort := strings.Split(":", l[3])[1]
	foreign := strings.Split(":", l[4])[0]
	foreignPort := strings.Split(":", l[4])[1]

	var conn Connection
	if local == host.hostName {
		conn.hostName = foreign
	} else {
		conn.hostName = local
	}

	conn.port = pickPort(localPort, foreignPort)

	return conn
}

func pickPort(l, f string) string {
	if check_regexp(`[a-zA-Z]`, l) {
		return l
	} else if check_regexp(`[a-zA-Z]`, f) {
		return f
	}

	if l < f {
		return l
	}
	return f
}

func checkExcludePattern(conf Config, l []string) bool {
	local := strings.Split(":", l[3])[0]
	localPort := strings.Split(":", l[3])[1]
	foreign := strings.Split(":", l[4])[0]
	foreignPort := strings.Split(":", l[4])[1]
	processName := l[6]

	// targetが特定の文字列を含んでいなかったらskip
	for _, h := range conf.hostCheck {
		if !strings.Contains(local, h) || !strings.Contains(foreign, h) {
			return false
		}
	}

	for _, exProcess := range conf.exProcesses {
		if strings.Contains(processName, exProcess) {
			return false
		}
	}

	for _, exPort := range conf.exPorts {
		if strings.Contains(localPort, exPort) || strings.Contains(foreignPort, exPort) {
			return false
		}
	}

	return true
}

func check_regexp(reg, str string) bool {
	return regexp.MustCompile(reg).Match([]byte(str))
}
