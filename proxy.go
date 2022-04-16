package main

import (
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/netutils"
)

type ConnectFunc func(local net.Conn) (net.Conn, error)

func DuplexForward(c1, c2 net.Conn) {
	worker := sync.WaitGroup{}
	worker.Add(+2)

	go func() {
		defer worker.Done()

		io.Copy(c1, c2)
		c1.Close()
	}()

	go func() {
		defer worker.Done()

		io.Copy(c2, c1)
		c2.Close()
	}()

	worker.Wait()
}

func Proxy(listener net.Listener, connectFunc ConnectFunc) error {
	worker := sync.WaitGroup{}

	for {
		local, err := listener.Accept()
		if err != nil {
			return err
		}

		remote, err := connectFunc(local)
		if err != nil {
			local.Close()
			return err
		}

		if remote != nil {
			worker.Add(+1)
			go func() {
				defer worker.Done()

				prefix := local.RemoteAddr().String() + "->" + remote.RemoteAddr().String() + ":"
				log.Info(prefix, "connected")
				defer log.Info(prefix, "closed")

				DuplexForward(local, remote)
			}()
		}
	}
}

func TCPProxy(port int, connectFunc ConnectFunc) error {
	local := "0.0.0.0:" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", local)
	if err != nil {
		return err
	}

	defer listener.Close()
	return Proxy(listener, connectFunc)
}

func UDPProxy(port int, connectFunc ConnectFunc) error {
	local := "0.0.0.0:" + strconv.Itoa(port)
	listener, err := netutils.NewUDP("udp", local)
	if err != nil {
		return err
	}

	defer listener.Close()
	return Proxy(listener, connectFunc)
}
