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

	flag.IntVar(&c.listenPort, "port", 4001, "Node listen port")
	flag.Parse()
	return c
}
