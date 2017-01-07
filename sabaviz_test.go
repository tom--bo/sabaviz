package main

import (
	"os"
	// "testing"
)

type dummyNetstatImpl struct{}

func (d dummyNetstatImpl) netstat(conf Config, host string) []Connection {
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
