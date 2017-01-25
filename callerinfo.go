package logrusUDP

import (
	"runtime"
	"strings"
)

type CallerInfo struct {
	FilePath string
	FuncName string
	LineNo   int
}

// Returns the file, function and line number of the function that called logrus
func getCallerInfo() *CallerInfo {
	var rpc [5]uintptr

	// iterate until we find non logrus function
	length := runtime.Callers(5, rpc[:])
	for idx := 0; idx < (length - 1); idx++ {
		f := runtime.FuncForPC(rpc[idx])
		funcName := f.Name()
		if strings.HasPrefix(funcName, "github.com/Sirupsen") {
			continue
		}
		filePath, lineNo := f.FileLine(rpc[idx])
		return &CallerInfo{filePath, funcName, lineNo}
	}
	return &CallerInfo{}
}
