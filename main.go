package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adamkirchberger/mtufind/pkg/mtufind"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func usage() {
	fmt.Printf("MTUFind is a tool to find the maximum transmission unit (MTU) to a destination")
	fmt.Printf("\n\nUsage: mtufind [OPTIONS] host\n")
	flag.PrintDefaults()
}

func main() {
	debug := flag.Bool("debug", false, "enable debug")
	flag.Usage = usage
	flag.Parse()

	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(0)
	}

	p, err := mtufind.New(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	maxSize, err := p.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("MTU to %s is %d bytes\n", p.Destination, maxSize)
}
