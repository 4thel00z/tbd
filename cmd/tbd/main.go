package main

import (
	"flag"
	"fmt"
	"github.com/4thel00z/tbd/pkg/libtbd"
	"os"
)

var (
	path   = flag.String("path", "", "Path to the file to be compiled and built")
	debug  = flag.Bool("debug", false, "Do you want to read debug output")
	owner  = flag.String("owner", "4thel00z", "Owner of the repository (for uploading the artefacts)")
	repo   = flag.String("repo", "", "Name of the repository (for uploading the artefacts)")
	apiKey = flag.String("api-key", "", "API key for github - it needs to have private repo perms")
)

func main() {
	flag.Parse()
	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *owner == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *repo == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *apiKey == "" {
		flag.Usage()
		os.Exit(1)
	}

	tbd, err := libtbd.DefaultTBD(*owner, *repo, *apiKey, *debug)
	libtbd.Must(err)
	res, err := tbd.BuildImage(*path)
	libtbd.Must(err)
	fmt.Println(res)
}
