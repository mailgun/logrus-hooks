package common

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mailgun/holster/v3/callstack"
	"github.com/mailru/easyjson/jwriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var DefaultFormatter = NewJSONFormater()

func NewJSONFormater() *JSONFormater {
	f := JSONFormater{}
	var err error

	if f.hostName, err = os.Hostname(); err != nil {
		f.hostName = "unknown"
	}
	f.appName = filepath.Base(os.Args[0])
	if f.pid = os.Getpid(); f.pid == 1 {
		f.pid = 0
	}
	f.cid = GetDockerCID()
	return &f
}

func (f *JSONFormater) Format(entry *logrus.Entry) ([]byte, error) {
	var caller *callstack.FrameInfo

	caller = GetLogrusCaller()

	rec := &LogRecord{
		Category:  "logrus",
		AppName:   f.appName,
		HostName:  f.hostName,
		LogLevel:  strings.ToUpper(entry.Level.String()),
		FileName:  caller.File,
		FuncName:  caller.Func,
		LineNo:    caller.LineNo,
		Message:   entry.Message,
		Context:   nil,
		Timestamp: Number(float64(entry.Time.UnixNano()) / 1000000000),
		CID:       f.cid,
		PID:       f.pid,
	}
	rec.FromFields(entry.Data)

	var w jwriter.Writer
	rec.MarshalEasyJSON(&w)
	if w.Error != nil {
		return nil, errors.Wrap(w.Error, "while marshalling json")
	}

	// Append a newline to the formatted record
	w.Buffer.AppendByte(byte(0x0a))

	buf := w.Buffer.BuildBytes()

	return buf, nil
}

type JSONFormater struct {
	appName  string
	hostName string
	cid      string
	pid      int
}
