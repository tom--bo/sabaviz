package main

import (
	"os"
	"time"
)

type dummyNetstatImpl struct{}

func (d dummyNetstatImpl) netstat(conf Config, host string) []Connection {

	switch host {
	case "test.server.local":
		return []Connection{}
	// For ExampleTestSabaviz()
	case "t1.server.local":
		c2 := Connection{hostName: "t2.server.local", port: "22"}
		return []Connection{c2}
	case "t2.server.local":
		c1 := Connection{hostName: "t1.server.local", port: "22"}
		c3 := Connection{hostName: "t3.server.local", port: "ssh"}
		return []Connection{c1, c3}
	case "t3.server.local":
		c2 := Connection{hostName: "t2.server.local", port: "ssh"}
		return []Connection{c2}
	// For ExampleTestSabavizMultihostsLongTime()
	case "L1.server.local":
		l2 := Connection{hostName: "L2.server.local", port: "ssh"}
		l3 := Connection{hostName: "L3.server.local", port: "ssh"}
		l4 := Connection{hostName: "L4.server.local", port: "ssh"}
		return []Connection{l2, l3, l4}
	case "L2.server.local":
		time.Sleep(3000 * time.Millisecond)
		l1 := Connection{hostName: "L1.server.local", port: "ssh"}
		return []Connection{l1}
	case "L3.server.local":
		time.Sleep(3000 * time.Millisecond)
		l1 := Connection{hostName: "L1.server.local", port: "ssh"}
		return []Connection{l1}
	case "L4.server.local":
		l1 := Connection{hostName: "L1.server.local", port: "ssh"}
		time.Sleep(3000 * time.Millisecond)
		return []Connection{l1}
	}

	time.Sleep(1 * time.Millisecond)
	return []Connection{}
}

func ExampleTestSabaviz() {
	conf := Config{
		exProcesses:     []string{},
		exPorts:         []string{},
		hostCheck:       []string{},
		user:            "",
		sshKey:          "",
		hostThreshold:   -1,
		connectionLimit: 20,
	}

	dummyInetstat := dummyNetstatImpl{}
	sabaviz := &Sabaviz{outStream: os.Stdout, errStream: os.Stderr, conf: conf, netstatImpl: dummyInetstat}
	sabaviz.exec("test.server.local")
	// Unordered output:
	// graph G {
	// 	"test.server.local";
	//
	// }

}

func ExampleTestSabavizMultihosts() {
	conf := Config{
		exProcesses:     []string{},
		exPorts:         []string{},
		hostCheck:       []string{},
		user:            "",
		sshKey:          "",
		hostThreshold:   -1,
		connectionLimit: 20,
	}

	dummyInetstat := dummyNetstatImpl{}
	sabaviz := &Sabaviz{outStream: os.Stdout, errStream: os.Stderr, conf: conf, netstatImpl: dummyInetstat}
	sabaviz.exec("t1.server.local")
	// Unordered output:
	// graph G {
	// 	"t1.server.local"--"t2.server.local"[ label="22" ];
	// 	"t2.server.local"--"t3.server.local"[ label="ssh" ];
	// 	"t1.server.local";
	// 	"t2.server.local";
	// 	"t3.server.local";
	//
	// }
}

func ExampleTestSabavizMultihostsLongTime() {
	conf := Config{
		exProcesses:     []string{},
		exPorts:         []string{},
		hostCheck:       []string{},
		user:            "",
		sshKey:          "",
		hostThreshold:   -1,
		connectionLimit: 20,
	}

	dummyInetstat := dummyNetstatImpl{}
	sabaviz := &Sabaviz{outStream: os.Stdout, errStream: os.Stderr, conf: conf, netstatImpl: dummyInetstat}
	sabaviz.exec("L1.server.local")
	// Unordered output:
	// graph G {
	// 	"L1.server.local"--"L2.server.local"[ label="ssh" ];
	// 	"L1.server.local"--"L3.server.local"[ label="ssh" ];
	// 	"L1.server.local"--"L4.server.local"[ label="ssh" ];
	// 	"L1.server.local";
	// 	"L2.server.local";
	// 	"L3.server.local";
	// 	"L4.server.local";
	//
	// }
}
