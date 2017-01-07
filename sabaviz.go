package main

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type NetstatInterface interface {
	netstat(conf Config, host string) []Connection
}

type Sabaviz struct {
	outStream, errStream io.Writer
	conf                 Config
	share                *Share
	netstatImpl          NetstatInterface
}

type netstatImpl struct{}

type Connection struct {
	hostName string
	port     string
}

type Share struct {
	found   int
	checked int
	queue   []string
	hostMap map[string]bool // hashmap for check host
	mu      sync.Mutex
}

func (s *Sabaviz) exec(target string) {
	g := &Graph{}
	g.NewGraph()
	g.AddNode(target)

	s.share = &Share{found: 1, checked: 0}
	s.share.queue = append([]string{}, target)
	s.share.hostMap = make(map[string]bool)
	s.share.hostMap[target] = true

	var localQueue []string
	cancelFlag := false

	var chs [3]chan string
	for i := range chs {
		chs[i] = make(chan string)
		go s.fanoutWorker(chs[i], g)
	}

	for s.share.found != s.share.checked {
		if s.conf.hostThreshold != -1 && s.share.checked >= s.conf.hostThreshold {
			// [fix] to break safely
			s.share.mu.Lock()
			cancelFlag = true
			break
		}
		s.share.mu.Lock()
		if len(s.share.queue) > 0 {
			localQueue = append(localQueue, s.share.queue...)
			s.share.queue = s.share.queue[len(s.share.queue):]
		}
		s.share.mu.Unlock()
		for _, host := range localQueue {
			select {
			case chs[0] <- host:
			case chs[1] <- host:
			case chs[2] <- host:
			default:
				s.share.mu.Lock()
				s.share.queue = append([]string{host}, s.share.queue...)
				s.share.mu.Unlock()
			}
		}
		localQueue = localQueue[len(localQueue):]
		time.Sleep(100 * time.Millisecond)
	}
	if cancelFlag {
		s.share.mu.Unlock()
	}
	fmt.Println(g.graph.String())
}

func (s *Sabaviz) fanoutWorker(ch chan string, g *Graph) {
	for {
		host, ok := <-ch
		if !ok {
			return
		}

		connections := s.netstatImpl.netstat(s.conf, host)
		if len(connections) >= s.conf.connectionLimit {
			s.share.mu.Lock()
			s.share.checked += 1
			s.share.mu.Unlock()
			continue
		}

		s.share.mu.Lock()
		for _, conn := range connections {
			g.AddConnectionOnce(host, conn)
			_, ok := s.share.hostMap[conn.hostName]
			if !ok {
				g.AddNode(conn.hostName)
				s.share.queue = append(s.share.queue, conn.hostName)
				s.share.hostMap[conn.hostName] = true
				s.share.found += 1
			}
		}
		s.share.checked += 1
		s.share.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

// return slice of Connection object which is unique by port and hostname
func (n netstatImpl) netstat(conf Config, host string) []Connection {
	var ret []Connection
	connMap := make(map[Connection]bool)

	netstatOption := ""
	distri := n.checkDistri(host)
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
		if len(l) > 6 && (strings.Contains(l[5], "ESTABLISHED") || strings.Contains(l[5], "TIME_WAIT")) {
			if n.checkExcludePattern(conf, l) {
				conn := n.makeConnectionObj(host, l)
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

func (n netstatImpl) checkDistri(host string) string {
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

func (n netstatImpl) makeConnectionObj(host string, l []string) Connection {
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
	conn.port = n.pickPort(localPort, foreignPort)
	return conn
}

func (n netstatImpl) pickPort(l, f string) string {
	if n.checkRegexp(`[a-zA-Z]`, l) {
		return l
	} else if n.checkRegexp(`[a-zA-Z]`, f) {
		return f
	}

	if l < f {
		return l
	}
	return f
}

func (n netstatImpl) checkExcludePattern(conf Config, l []string) bool {
	local := strings.Split(l[3], ":")[0]
	localPort := strings.Split(l[3], ":")[1]
	foreign := strings.Split(l[4], ":")[0]
	foreignPort := strings.Split(l[4], ":")[1]
	processName := l[6]

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

func (n netstatImpl) checkRegexp(reg, str string) bool {
	return regexp.MustCompile(reg).Match([]byte(str))
}
