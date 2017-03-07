package udploghook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type logRecord struct {
	Context   map[string]interface{} `json:"context,omitempty"`
	AppName   string                 `json:"appname"`
	HostName  string                 `json:"hostname"`
	LogLevel  string                 `json:"logLevel"`
	FileName  string                 `json:"filename"`
	FuncName  string                 `json:"funcName"`
	LineNo    int                    `json:"lineno"`
	Message   string                 `json:"message"`
	Timestamp Number                 `json:"timestamp"`
	PID       int                    `json:"pid"`
	TID       string                 `json:"tid"`
	ExcType   string                 `json:"excType"`
	ExcText   string                 `json:"excText"`
	ExcValue  string                 `json:"excValue"`
}

func (r *logRecord) fromFields(fields logrus.Fields) {
	if len(fields) == 0 {
		return
	}
	r.Context = make(map[string]interface{})
	for k, v := range fields {
		switch k {
		case "err":
			if v, ok := v.(error); ok {
				r.ExcValue = v.Error()
				r.ExcType = fmt.Sprintf("%T", v)
				r.ExcText = fmt.Sprintf("%+v", v)
				continue
			}
		case "tid":
			if v, ok := v.(string); ok {
				r.TID = v
				continue
			}
		}
		expandNested(k, v, r.Context)
	}
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

func expandNested(key string, value interface{}, dest map[string]interface{}) {
	if strings.ContainsRune(key, '.') {
		parts := strings.SplitN(key, ".", 2)
		// This nested value might already exist
		nested, isMap := dest[parts[0]].(map[string]interface{})
		if !isMap {
			// if not a map, overwrite current entry and make it a map
			nested = make(map[string]interface{})
			dest[parts[0]] = nested
		}
		expandNested(parts[1], value, nested)
	}
	switch value.(type) {
	case *http.Request:
		dest[key] = requestToMap(value.(*http.Request))
	default:
		dest[key] = value
	}
}

// Given a *http.Request return a map with detailed information about the request
func requestToMap(req *http.Request) map[string]interface{} {
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

func (h *UDPHook) Fire(entry *logrus.Entry) error {
	var caller *CallerInfo
	var err error

	caller = getCallerInfo()

	rec := &logRecord{
		AppName:   h.appName,
		HostName:  h.hostName,
		LogLevel:  strings.ToUpper(entry.Level.String()),
		FileName:  caller.FilePath,
		FuncName:  caller.FuncName,
		LineNo:    caller.LineNo,
		Message:   entry.Message,
		Context:   nil,
		Timestamp: Number(float64(entry.Time.UnixNano()) / 1000000000),
		PID:       h.pid,
	}
	rec.fromFields(entry.Data)

	dump, err := json.Marshal(rec)
	if err != nil {
		return errors.Wrap(err, "UDPHook.Fire() - json.Marshall() error")
	}

	// Add a category to the beginning of the log entry followed by the json entry
	buf := bytes.NewBuffer([]byte("logrus:"))
	buf.Write(dump)

	if h.debug {
		fmt.Printf("%s\n", buf.String())
	}

	// Send the buffer to udplog
	err = h.sendUDP(buf.Bytes())
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
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (h *UDPHook) SetDebug(set bool) {
	h.debug = set
}
