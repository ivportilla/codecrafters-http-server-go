package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	POOL_SIZE     = 5
	SLEEP_TIMEOUT = 200 * time.Millisecond
)

type ConnectionPool struct {
	connections []net.Conn
	mut         sync.Mutex
	maxSize     int
}

func NewConnectionPool(maxSize int) *ConnectionPool {
	return &ConnectionPool{
		connections: make([]net.Conn, 0, maxSize),
		maxSize:     maxSize,
	}
}

func (cp *ConnectionPool) Add(conn net.Conn) bool {
	cp.mut.Lock()
	defer cp.mut.Unlock()

	if len(cp.connections) >= cp.maxSize {
		return false
	}

	cp.connections = append(cp.connections, conn)
	return true
}

func (cp *ConnectionPool) Remove(connToClose net.Conn) bool {
	cp.mut.Lock()
	defer cp.mut.Unlock()

	idx := -1
	for i, c := range cp.connections {
		if c == connToClose {
			idx = i
			break
		}
	}

	if idx > -1 {
		cp.connections = append(cp.connections[:idx], cp.connections[idx+1:]...)
		fmt.Println("Connection deleted, new pool size: ", len(cp.connections))
		return true
	}

	return false
}

func (cp *ConnectionPool) HandlePool(closeCh chan net.Conn) {
	for conn := range closeCh {
		conn.Close()
		cp.Remove(conn)
	}
}

func handleConnections(listener net.Listener, cp *ConnectionPool, closeCh chan net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		}

		connAdded := cp.Add(conn)
		if !connAdded {
			fmt.Println("Connection pool is full at the moment, rejecting connection")
			conn.Close()
			continue
		}
		go reqHandler(conn, closeCh)
		time.Sleep(SLEEP_TIMEOUT)
	}
}