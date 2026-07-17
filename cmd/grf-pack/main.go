package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kivutar/goro/res"
)

func main() {
	output := flag.String("o", "", "output GRF path")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s -o output.grf input-dir\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 || *output == "" {
		flag.Usage()
		os.Exit(2)
	}

	stats, err := res.PackGRF(*output, flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "grf-pack: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("packed %d files (%d bytes) into %s\n", stats.Files, stats.Bytes, *output)
}
