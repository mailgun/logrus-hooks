package common

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

type Number float64

func (n Number) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%f", n)), nil
}

//easyjson:json
type LogRecord struct {
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
	TID       string                 `json:"tid,omitempty"`
	ExcType   string                 `json:"excType,omitempty"`
	ExcText   string                 `json:"excText,omitempty"`
	ExcValue  string                 `json:"excValue,omitempty"`
}

func (r *LogRecord) FromFields(fields logrus.Fields) {
	if len(fields) == 0 {
		return
	}
	r.Context = make(map[string]interface{})
	for k, v := range fields {
		switch k {
		case "tid":
			if v, ok := v.(string); ok {
				r.TID = v
				continue
			}
		case "excValue":
			if v, ok := v.(string); ok {
				r.ExcValue = v
				continue
			}
		case "excType":
			if v, ok := v.(string); ok {
				r.ExcType = v
				continue
			}
		case "excText":
			if v, ok := v.(string); ok {
				r.ExcText = v
				continue
			}
		case "excFuncName":
			if v, ok := v.(string); ok {
				r.FuncName = v
				continue
			}
		case "excLineno":
			if v, ok := v.(int); ok {
				r.LineNo = v
				continue
			}
		case "excFileName":
			if v, ok := v.(string); ok {
				r.FileName = v
				continue
			}
		}
		ExpandNested(k, v, r.Context)
	}
}
