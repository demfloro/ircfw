package ircfw

// Mock object complying with net.Conn interface to log and proxy all activity. For testing purposes only.

import (
	"log"
	"net"
	"os"
	"time"
)

const logdir = "/tmp/mockproxy"

type mockProxy struct {
	log    *log.Logger
	file   *os.File
	socket net.Conn
}

func newMockProxy(socket net.Conn, logger *log.Logger) *mockProxy {
	var file *os.File
	if logger == nil {
		err := os.Mkdir("/tmp/mockproxy", 0700)
		if err != nil {
			if !os.IsExist(err) {
				log.Fatal(err)
			}
		}
		file, err := os.OpenFile(logdir+"/log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		logger = log.New(file, "mockProxy:", 0)
	}
	m := mockProxy{
		log:    logger,
		file:   file,
		socket: socket,
	}
	return &m
}

func (m *mockProxy) Read(b []byte) (n int, err error) {
	n, err = m.socket.Read(b)
	m.log.Printf("Read b=%v, n=%d, err=%v", string(b), n, err)
	return
}

func (m *mockProxy) Write(b []byte) (n int, err error) {
	//m.log.Printf("About to write b=%v", string(b))
	n, err = m.socket.Write(b)
	m.log.Printf("Wrote b=%v, n=%d, err=%v", string(b), n, err)
	return
}

func (m *mockProxy) Close() error {
	err := m.socket.Close()
	m.log.Printf("Closed socket, %v", err)
	m.file.Close()
	return err
}

func (m *mockProxy) LocalAddr() net.Addr {
	addr := m.socket.LocalAddr()
	m.log.Printf("Requested local addr %v", addr)
	return addr
}

func (m *mockProxy) RemoteAddr() net.Addr {
	addr := m.socket.RemoteAddr()
	m.log.Printf("Requested remote addr %v", addr)
	return addr
}

func (m *mockProxy) SetDeadline(t time.Time) error {
	err := m.socket.SetDeadline(t)
	m.log.Printf("Set Deadline %v, err=%v", t, err)
	return err
}

func (m *mockProxy) SetReadDeadline(t time.Time) error {
	err := m.socket.SetReadDeadline(t)
	m.log.Printf("Set Read Deadline %v, err=%v", t, err)
	return err
}

func (m *mockProxy) SetWriteDeadline(t time.Time) error {
	err := m.socket.SetWriteDeadline(t)
	m.log.Printf("Set Write Deadline %v, err=%v", t, err)
	return err
}
