package udploghook

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/mailgun/holster/stack"
	"github.com/mailgun/logrus-hooks/common"
	"github.com/mailru/easyjson/jwriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type UDPHook struct {
	hostName string
	appName  string
	conn     net.Conn
	pid      int
	debug    bool
}

func New(host string, port int) (*UDPHook, error) {
	h := UDPHook{}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	h.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	if h.hostName, err = os.Hostname(); err != nil {
		h.hostName = "unknown_host"
	}
	h.appName = filepath.Base(os.Args[0])
	h.pid = os.Getpid()

	return &h, nil
}

func (h *UDPHook) Fire(entry *logrus.Entry) error {
	var caller *stack.FrameInfo
	var err error

	caller = common.GetLogrusCaller()

	rec := &common.LogRecord{
		AppName:   h.appName,
		HostName:  h.hostName,
		LogLevel:  strings.ToUpper(entry.Level.String()),
		FileName:  caller.File,
		FuncName:  caller.Func,
		LineNo:    caller.LineNo,
		Message:   entry.Message,
		Context:   nil,
		Timestamp: common.Number(float64(entry.Time.UnixNano()) / 1000000000),
		PID:       h.pid,
	}
	rec.FromFields(entry.Data)

	// Marshal the log record to JSON string with a category prefix.
	var w jwriter.Writer
	w.RawString("logrus:")
	rec.MarshalEasyJSON(&w)
	if w.Error != nil {
		return errors.Wrap(w.Error, "UDPHook.Fire() - json.Marshall() error")
	}
	buf := w.Buffer.BuildBytes()

	if h.debug {
		fmt.Printf("%s\n", string(buf))
	}

	// Send the buffer to udplog
	err = h.sendUDP(buf)
	if err != nil {
		return errors.Wrap(err, "UDPHook.Fire()")
	}
	return nil
}

func (h *UDPHook) sendUDP(buf []byte) error {
	length, err := h.conn.Write(buf)
	if err != nil {
		return errors.Wrap(err, "Write() error")
	}
	if length != len(buf) {
		return errors.Wrapf(err, "Write() only wrote %d of %d bytes", length, len(buf))
	}
	return nil
}

// Given an io reader send the contents of the reader to udplog
func (h *UDPHook) SendIO(input io.Reader) error {
	// Append our identifier
	buf := bytes.NewBuffer([]byte("logrus:"))
	_, err := buf.ReadFrom(input)
	if err != nil {
		return errors.Wrap(err, "UDPHook.SendIO()")
	}

	if h.debug {
		fmt.Printf("%s\n", buf.String())
	}

	// Send to UDPLog
	err = h.sendUDP(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "UDPHook.SendIO()")
	}
	return nil
}

// Levels returns the available logging levels.
func (h *UDPHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *UDPHook) SetDebug(set bool) {
	h.debug = set
}
