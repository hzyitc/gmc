package main

import (
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/hzyitc/mnh/log"
)

type Interface interface {
	ClosedChan() <-chan struct{}
	Close() error
}

type proxy struct {
	port int
	url  string

	service string

	worker      *sync.WaitGroup
	closingChan chan struct{}
	closedChan  chan struct{}

	listener net.Listener
}

func (s *proxy) server_tryConnect() (net.Conn, error) {
	if s.service != "" {
		c, err := net.Dial("tcp", s.service)
		if err == nil {
			return c, nil
		}
	}

	addr, err := mnhv1_query(s.url)
	if err != nil {
		return nil, err
	}

	s.service = addr

	return net.Dial("tcp", s.service)
}

func (s *proxy) server_handle(conn net.Conn) {
	s.worker.Add(+1)
	defer s.worker.Done()

	defer conn.Close()

	log.Info("new connection", conn.RemoteAddr().String())

	c, err := s.server_tryConnect()
	if err != nil {
		log.Error(conn.RemoteAddr().String(), err.Error())
		return
	}
	defer c.Close()

	prefix := conn.RemoteAddr().String() + "->" + c.RemoteAddr().String() + ":"

	log.Info(prefix, "connected")
	defer log.Info(prefix, "closed")

	closing := make(chan int)

	go func() {
		io.Copy(conn, c)
		conn.Close()
		closing <- 1
	}()

	go func() {
		io.Copy(c, conn)
		c.Close()
		closing <- 1
	}()

	running := 2
	for {
		select {
		case <-s.closingChan:
			return
		case <-closing:
			running--
			if running == 0 {
				return
			}
		}
	}
}

func (s *proxy) server_main() {
	s.worker.Add(+1)
	defer s.worker.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.closingChan:
				return
			default:
				log.Error("server_main error", err.Error())
				continue
			}
		}
		go s.server_handle(conn)
	}
}

func NewProxy(port int, url string) (Interface, error) {
	local := "0.0.0.0:" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", local)
	if err != nil {
		return nil, err
	}

	s := &proxy{
		port,
		url,

		"",

		new(sync.WaitGroup),
		make(chan struct{}),
		make(chan struct{}),

		listener,
	}

	go s.server_main()

	return s, nil
}

func (s *proxy) ClosedChan() <-chan struct{} {
	return s.closedChan
}

func (s *proxy) Close() error {
	select {
	case <-s.closingChan:
		return nil
	default:
		break
	}
	close(s.closingChan)

	err := s.listener.Close()

	s.worker.Wait()

	close(s.closedChan)
	return err
}
