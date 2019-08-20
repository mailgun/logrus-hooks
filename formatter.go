package logrus_hooks

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mailgun/holster/errors"
	"github.com/mailgun/holster/stack"
	"github.com/mailgun/logrus-hooks/common"
	"github.com/mailru/easyjson/jwriter"
	"github.com/sirupsen/logrus"
)

type JSONFormater struct {
	appName string
	hostName string
	cid string
	pid int
}

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
	f.cid = common.GetDockerCID()
	return &f
}


func (f *JSONFormater) Format(entry *logrus.Entry) ([]byte, error) {
	var caller *stack.FrameInfo

	caller = common.GetLogrusCaller()

	rec := &common.LogRecord{
		Category:  "logrus",
		AppName:   f.appName,
		HostName:  f.hostName,
		LogLevel:  strings.ToUpper(entry.Level.String()),
		FileName:  caller.File,
		FuncName:  caller.Func,
		LineNo:    caller.LineNo,
		Message:   entry.Message,
		Context:   nil,
		Timestamp: common.Number(float64(entry.Time.UnixNano()) / 1000000000),
		CID:       f.cid,
		PID:       f.pid,
	}
	rec.FromFields(entry.Data)

	var w jwriter.Writer
	rec.MarshalEasyJSON(&w)
	if w.Error != nil {
		return nil, errors.Wrap(w.Error, "while marshalling json")
	}
	buf := w.Buffer.BuildBytes()

	return buf, nil
}
