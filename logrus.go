package udploghook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type Number float64

func (n Number) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%f", n)), nil
}

type UDPHook struct {
	hostName string
	appName  string
	conn     net.Conn
	pid      int
	debug    bool
}

type LogRecord struct {
	Context   map[string]interface{} `json:"context"`
	AppName   string                 `json:"appname"`
	HostName  string                 `json:"hostname"`
	LogLevel  string                 `json:"logLevel"`
	FileName  string                 `json:"filename"`
	FuncName  string                 `json:"funcName"`
	LineNo    int                    `json:"lineno"`
	Message   string                 `json:"message"`
	Timestamp Number                 `json:"timestamp"`
}

func New(host string, port int) (*UDPHook, error) {
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

func expandNested(key string, value interface{}, dest map[string]interface{}) map[string]interface{} {
	if dest == nil {
		dest = make(map[string]interface{})
	}

	if strings.ContainsRune(key, '.') {
		parts := strings.SplitN(key, ".", 2)
		// This nested value might already exist
		nested, isMap := dest[parts[0]].(map[string]interface{})
		if !isMap {
			// if not a map, overwrite current entry and make it a map
			nested = nil
		}
		dest[parts[0]] = expandNested(parts[1], value, nested)
		return dest
	}
	switch value.(type) {
	case *http.Request:
		dest[key] = RequestToMap(value.(*http.Request))
	default:
		dest[key] = value
	}
	return dest
}

// Given a *http.Request return a map with detailed information about the request
func RequestToMap(req *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"headers":   req.Header,
		"ip":        req.RemoteAddr,
		"method":    req.Method,
		"params":    req.Form,
		"size":      req.ContentLength,
		"url":       req.URL.String(),
		"useragent": req.Header.Get("User-Agent"),
	}
}

// Given a logrus.Fields struct convert the values to a standard map
func FieldsToMap(fields logrus.Fields) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range fields {
		result = expandNested(key, value, result)
	}
	return result
}

func (hook *UDPHook) Fire(entry *logrus.Entry) error {
	var caller *CallerInfo
	var err error

	caller = getCallerInfo()

	rec := &LogRecord{
		AppName:   hook.appName,
		HostName:  hook.hostName,
		LogLevel:  strings.ToUpper(entry.Level.String()),
		FileName:  caller.FilePath,
		FuncName:  caller.FuncName,
		LineNo:    caller.LineNo,
		Message:   entry.Message,
		Context:   nil,
		Timestamp: Number(float64(entry.Time.UnixNano()) / 1000000000),
	}

	if len(entry.Data) != 0 {
		rec.Context = FieldsToMap(entry.Data)
	}

	dump, err := json.Marshal(rec)
	if err != nil {
		return errors.Wrap(err, "UDPHook.Fire() - json.Marshall() error")
	}
	if hook.debug {
		fmt.Printf("Send: %s\n", string(dump))
	}

	// Add a category to the beginning of the log entry followed by the json entry
	buf := bytes.NewBuffer([]byte("logrus:"))
	buf.Write(dump)

	// Send the buffer to udplog
	length, err := hook.conn.Write(buf.Bytes())
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

func (hook *UDPHook) SetDebug(set bool) {
	hook.debug = set
}
