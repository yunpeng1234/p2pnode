package main

import (
	"flag"
)

type config struct {
	identifier string
	protocol   string
	listenHost string
	listenPort int
	main       bool
}

func parseFlags() *config {
	c := &config{}

	flag.StringVar(&c.identifier, "indentier", "SnapInnovation", "Unique string to identify")
	flag.StringVar(&c.listenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&c.protocol, "pid", "/messaging", "Sets a protocol id for stream headers")
	flag.IntVar(&c.listenPort, "port", 4001, "Node listen port")
	flag.BoolVar(&c.main, "main", false, "Is this node the main node you are interacting with")
	flag.Parse()
	return c
}
