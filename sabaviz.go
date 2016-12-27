package main

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type Sabaviz struct {
	outStream, errStream io.Writer
	conf                 Config
}

type Connection struct {
	hostName string
	port     string
}

type Share struct {
	found   int
	stated  int
	queue   []string
	hostMap map[string]bool // hashmap for check host
	mu      sync.Mutex
}

func (s Sabaviz) main(target string) {
	g := &Graph{}
	g.NewGraph()

	share := Share{found: 1, stated: 0}
	share.hostMap = make(map[string]bool)
	share.queue = append(share.queue, target)
	share.hostMap[target] = true
	g.AddNode(target)

	ch1 := make(chan string)
	ch2 := make(chan string)
	ch3 := make(chan string)
	go fanoutWorker(ch1, share, s.conf, g)
	go fanoutWorker(ch2, share, s.conf, g)
	go fanoutWorker(ch3, share, s.conf, g)
	var localQueue []string
	cancelFlag := false

	for share.found != share.stated {
		if s.conf.hostThreshold != -1 && share.found >= s.conf.hostThreshold {
			// fix to break safely
			share.mu.Lock()
			cancelFlag = true
			break
		}
		share.mu.Lock()
		if len(share.queue) > 0 {
			localQueue = append(localQueue, share.queue...)
			share.queue = share.queue[len(share.queue):]
		}
		share.mu.Unlock()
		for _, host := range localQueue {
			select {
			case ch1 <- host:
			case ch2 <- host:
			case ch3 <- host:
			default:
			}
		}
		localQueue = localQueue[len(localQueue):]
	}
	if cancelFlag {
		share.mu.Unlock()
	}
	fmt.Println(g.graph.String())
}

func fanoutWorker(ch chan string, share Share, conf Config, g *Graph) {
	for {
		host, ok := <-ch
		if !ok {
			return
		}

		connections := netstat(host, conf)
		if len(connections) >= conf.connectionLimit {
			continue
		}

		share.mu.Lock()
		for _, conn := range connections {
			g.AddConnectionOnce(host, conn)
			_, ok := share.hostMap[conn.hostName]
			if !ok {
				g.AddNode(conn.hostName)
				share.queue = append(share.queue, conn.hostName)
				share.hostMap[conn.hostName] = true
				share.found += 1
			}
		}
		share.stated += 1
		share.mu.Unlock()
	}
}

// return slice of Connection object which is unique by port and hostname
func netstat(host string, conf Config) []Connection {
	// hostはチャネルから受け取る
	var ret []Connection
	connMap := make(map[Connection]bool)

	netstatOption := ""
	distri := checkDistri(host)
	switch distri {
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

	out, err := exec.Command("ssh", host, "netstat", netstatOption).Output()
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

func makeConnectionObj(host string, l []string) Connection {
	local := strings.Split(l[3], ":")[0]
	localPort := strings.Split(l[3], ":")[1]
	foreign := strings.Split(l[4], ":")[0]
	foreignPort := strings.Split(l[4], ":")[1]

	var conn Connection
	if local == host {
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
	local := strings.Split(l[3], ":")[0]
	localPort := strings.Split(l[3], ":")[1]
	foreign := strings.Split(l[4], ":")[0]
	foreignPort := strings.Split(l[4], ":")[1]
	processName := l[6]

	// targetが特定の文字列を含んでいなかったらskip
	for _, h := range conf.hostCheck {
		if !strings.Contains(local, h) || !strings.Contains(foreign, h) {
			return false
		}
	}

	for _, exProcess := range conf.exProcesses {
		if exProcess != "" && strings.Contains(processName, exProcess) {
			return false
		}
	}

	for _, exPort := range conf.exPorts {
		if exPort != "" && (strings.Contains(localPort, exPort) || strings.Contains(foreignPort, exPort)) {
			return false
		}
	}

	return true
}

func check_regexp(reg, str string) bool {
	return regexp.MustCompile(reg).Match([]byte(str))
}
