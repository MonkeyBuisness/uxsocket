package uxsocket

import (
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	network = "unix"
)

type Pipe struct {
	comRead  chan []byte
	comWrite chan []byte
	conn     net.Conn
	closed   bool
}

type Server struct {
	listener net.Listener
	pipes    []*Pipe
}

type Client struct {
	pipe *Pipe
}

func makePipe(conn net.Conn) *Pipe {
	return &Pipe{
		comRead:  make(chan []byte, 1),
		comWrite: make(chan []byte, 1),
		conn:     conn,
	}
}

func NewClient(address string) (*Client, error) {
	// connect to socket
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	return &Client{
		pipe: makePipe(conn),
	}, nil
}

func (p *Pipe) listen() error {
	// write cycle for pipe
	for {
		select {
		case data, ok := <-p.comWrite:
			if !ok {
				return nil
			}

			if _, err := p.conn.Write(data); err != nil {
				return err
			}
		}
	}
}

func (p *Pipe) close() error {
	if p.closed {
		return nil
	}

	close(p.comRead)
	close(p.comWrite)

	p.closed = true

	return p.conn.Close()
}

func (s *Server) Listen(address string) (err error) {
	// unlink existing socket
	syscall.Unlink(address)
	if s.listener, err = net.Listen(network, address); err != nil {
		return
	}

	// listen for system interrupt
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		<-c
		s.Close()
		os.Exit(0)
	}(s.listener, sigc)

	// start connection listener
	for {
		// wait for accept new connection
		fd, err := s.listener.Accept()
		if err != nil {
			return err
		}

		// create new pipe
		pipe := makePipe(fd)

		// bind pipe to server
		s.pipes = append(s.pipes, pipe)

		// start pipe listener
		go func(p *Pipe, i int) {
			_ = p.listen()
			p.close()
		}(pipe, len(s.pipes)-1)
	}
}

func (s *Server) Close() error {
	// close all opened pipes
	for _, pipe := range s.pipes {
		pipe.close()
	}

	// close listener
	return s.listener.Close()
}

func (s *Server) Write(data []byte) {
	for _, pipe := range s.pipes {
		if !pipe.closed {
			pipe.comWrite <- data
		}
	}
}

func (c *Client) Listen() error {
	// start pipe listener
	return c.pipe.listen()
}

func (c *Client) Close() error {
	return c.pipe.close()
}

func (c *Client) Write(data []byte) {
	c.pipe.comWrite <- data
}

func (c *Client) Read(buf []byte) (int, error) {
	return c.pipe.conn.Read(buf)
}
