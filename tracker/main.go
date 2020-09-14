package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"

	Core "github.com/eyedeekay/opentracker"
	//"github.com/i19/autorestart"
	"github.com/vvampirius/retracker/core/common"
)

var config common.Config

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		PrintRepo()
	}
	age := flag.Float64("a", 1800, "Keep 'n' minutes peer in memory")
	debug := flag.Bool("d", false, "Debug mode")
	ver := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *ver {
		fmt.Println(VERSION)
		PrintRepo()
		syscall.Exit(0)
	}

	config = common.Config{
		Listen:  "127.0.0.1:80",
		Debug:   *debug,
		Age:     *age,
		XRealIP: false,
	}

	//	autorestart.Run(worker)
	_, err := Core.New(&config)
	if err != nil {
		panic(err)
	}
}

const VERSION = 0.2

func PrintRepo() {
	fmt.Fprintln(os.Stderr, "\n# https://github.com/eyedeekay/opentracker")
}

func worker() {
	Core.New(&config)
	for i := 0; i <= 10; i++ {
		time.Sleep(time.Second)
		fmt.Println(i)
	}
	panic("Panicing ... ")
}
