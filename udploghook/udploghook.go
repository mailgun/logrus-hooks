package udploghook

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/mailgun/logrus-hooks/common"
	"github.com/mailru/easyjson/jwriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type UDPHook struct {
	formatter logrus.Formatter
	conn      net.Conn
	debug     bool
}

func New(host string, port int) (*UDPHook, error) {
	h := UDPHook{
		formatter: common.DefaultFormatter,
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	h.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return &h, nil
}

func (h *UDPHook) Fire(entry *logrus.Entry) error {
	var w jwriter.Writer
	w.RawString("logrus:")
	w.Raw(h.formatter.Format(entry))

	if w.Error != nil {
		return errors.Wrap(w.Error, "while formatting entry")
	}
	buf := w.Buffer.BuildBytes()

	if h.debug {
		fmt.Printf("%s\n", string(buf))
	}

	// Send the buffer to udplog
	err := h.sendUDP(buf)
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
