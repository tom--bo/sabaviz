package main

import (
	"os"
	"testing"
)

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
	sabaviz := &Sabaviz{outStream: os.Stdout, errStream: os.Stderr, conf: conf}
	sabaviz.main("test.server.local")
	// Unordered output:
	// graph G {
	// 	"test.server.local";
	//
	// }

}
