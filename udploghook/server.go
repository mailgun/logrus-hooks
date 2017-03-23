package udploghook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Test UDP Server
type Server struct {
	conn *net.UDPConn
	done chan struct{}
	resp chan []byte
}

func NewServer(host string, port int) (*Server, error) {
	var err error

	bind := fmt.Sprintf("%s:%d", host, port)
	address, err := net.ResolveUDPAddr("udp", bind)
	if err != nil {
		return nil, errors.Wrapf(err, "net.ResolveUDPAddr(%s)", bind)
	}

	udp := Server{}

	udp.conn, err = net.ListenUDP("udp", address)
	if err != nil {
		return nil, errors.Wrap(err, "net.ListenUDP()")
	}

	udp.resp = make(chan []byte)
	udp.done = make(chan struct{})
	go func() {
		buf := make([]byte, 2024)
		for {
			length, _, _ := udp.conn.ReadFromUDP(buf)
			select {
			case udp.resp <- buf[0:length]:
			case <-udp.done:
				return

			}
		}
	}()
	return &udp, nil
}

func (udp *Server) Host() string {
	parts := strings.Split(udp.conn.LocalAddr().String(), ":")
	return parts[0]
}

func (udp *Server) Port() int {
	parts := strings.Split(udp.conn.LocalAddr().String(), ":")
	value, _ := strconv.ParseInt(parts[1], 10, 64)
	return int(value)
}

func (udp *Server) GetRequest() map[string]interface{} {
	var result map[string]interface{}
	data := <-udp.resp
	fmt.Printf("%s\n", string(data))
	parts := bytes.SplitN(data, []byte(":"), 2)
	if err := json.Unmarshal(parts[1], &result); err != nil {
		fmt.Printf("json.Unmarshal() error: %s\n", err)
	}
	return result
}

func (udp *Server) Close() {
	close(udp.done)
	udp.conn.Close()
}
