package main

import (
	"flag"
	"strconv"

	"github.com/hzyitc/mnh/log"
)

func main() {
	flag.Parse()
	args := flag.Args()

	port, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error(err.Error())
		return
	}

	s, err := NewProxy(port, args[1])
	if err != nil {
		log.Error(err.Error())
		return
	}

	<-s.ClosedChan()
}
