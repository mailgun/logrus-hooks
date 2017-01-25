package logrusUDP

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type UDPHook struct {
	hostName string
	appName  string
	conn     net.Conn
	pid      int
}

type LogRecord struct {
	AppName   string  `json:"appname"`
	HostName  string  `json:"hostname"`
	LogLevel  string  `json:"logLevel"`
	FileName  string  `json:"filename"`
	FuncName  string  `json:"funcName"`
	LineNo    int     `json:"lineno"`
	Context   string  `json:"context,omitempty"`
	Message   string  `json:"message"`
	Timestamp float64 `json:"timestamp"`
}

func NewLogHook(host string, port int) (*UDPHook, error) {
	hook := UDPHook{}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	hook.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	if hook.hostName, err = os.Hostname(); err != nil {
		hook.hostName = "unknown_host"
	}
	hook.appName = filepath.Base(os.Args[0])
	hook.pid = os.Getpid()

	return &hook, nil
}

func (hook *UDPHook) Fire(entry *logrus.Entry) error {
	var caller *CallerInfo
	var context []byte
	var err error
	caller = getCallerInfo()

	if len(entry.Data) != 0 {
		context, err = json.Marshal(entry.Data)
		if err != nil {
			return errors.Wrap(err, "UDPHook.Fire() - json.Marshall() Data error")
		}
	}

	rec := &LogRecord{
		AppName:   hook.appName,
		HostName:  hook.hostName,
		LogLevel:  strings.ToUpper(entry.Level.String()),
		FileName:  caller.FilePath,
		FuncName:  caller.FuncName,
		LineNo:    caller.LineNo,
		Message:   entry.Message,
		Context:   string(context),
		Timestamp: float64(entry.Time.UnixNano()) / 1000000000,
	}

	dump, err := json.Marshal(rec)
	if err != nil {
		return errors.Wrap(err, "UDPHook.Fire() - json.Marshall() error")
	}
	length, err := hook.conn.Write(dump)
	if err != nil {
		return errors.Wrap(err, "UDPHook.Fire() - Write() error")
	}
	if length != len(dump) {
		return errors.Wrapf(err, "UDPHook.Fire() - Write() only wrote %d of %d bytes", length, len(dump))
	}
	return nil
}

// Levels returns the available logging levels.
func (hook *UDPHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
