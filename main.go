package main

import (
	"errors"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/netutils"
)

var version = "v0.0.0"

func usage() {
	print(os.Args[0] + " (" + version + ")\n")
	print("Usage:\n")
	print("  " + os.Args[0] + " tcp <listen port> <mnh query url>\n")
	print("  " + os.Args[0] + " udp <listen port> <mnh query url>\n")
}

func main() {
	args := os.Args[1:]

	if len(args) < 3 || len(args) > 3 {
		usage()
		return
	}

	port, err := strconv.Atoi(args[1])
	if err != nil {
		log.Error("Parse port error:", err.Error())
		return
	}

	service := ""
	connectFunc := func(local net.Conn) (net.Conn, error) {
		log.Info("new connection", local.RemoteAddr().String())

		remote, err := func() (net.Conn, error) {
			if service != "" {
				remote, err := net.Dial(args[0], service)
				if err == nil {
					return remote, nil
				}
			}

			log.Info("Querying...")
			addr, err := mnhv1_query(args[2])
			if err != nil {
				return nil, err
			}
			log.Info("Result: " + addr)

			service = addr
			return net.Dial(args[0], service)
		}()

		if args[0] == "udp" && remote != nil {
			remote = netutils.NewAutoTimeoutConn(remote, time.Minute)
		}

		return remote, err
	}

	switch args[0] {
	case "tcp":
		err = TCPProxy(port, connectFunc)
	case "udp":
		err = UDPProxy(port, connectFunc)
	default:
		err = errors.New("Unknown mode: " + args[0])
	}
	if err != nil {
		log.Error(err.Error())
		return
	}
}
