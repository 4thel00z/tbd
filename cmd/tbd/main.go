package main

import (
	"flag"
	"fmt"
	"github.com/4thel00z/tbd/pkg/libtbd"
	"os"
)

var (
	path  = flag.String("path", "", "Path to the file to be compiled and built")
	debug = flag.Bool("debug", false, "Do you want to read debug output")
)

func main() {
	flag.Parse()
	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}
	tbd, err := libtbd.DefaultTBD(*debug)
	libtbd.Must(err)
	res, err := tbd.BuildImage(*path)
	libtbd.Must(err)
	fmt.Println(res)
}
