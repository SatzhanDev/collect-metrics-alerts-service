package main

import "flag"

type Config struct {
	Addr string
}

func parseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "HTTP server address")
	flag.Parse()
	return cfg
}
