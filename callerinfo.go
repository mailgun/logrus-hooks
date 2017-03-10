package udploghook

import (
	"runtime"
	"strings"

	"github.com/mailgun/holster/stack"
)

// Returns the file, function and line number of the function that called logrus
func getCallerInfo() *stack.FrameInfo {
	var rpc [5]uintptr

	// iterate until we find non logrus function
	length := runtime.Callers(5, rpc[:])
	for idx := 0; idx < (length - 1); idx++ {
		f := runtime.FuncForPC(rpc[idx])
		funcName := f.Name()
		if strings.HasPrefix(strings.ToLower(funcName), "github.com/sirupsen") {
			continue
		}
		filePath, lineNo := f.FileLine(rpc[idx])
		return &stack.FrameInfo{File: filePath, Func: funcName, LineNo: lineNo}
	}
	return &stack.FrameInfo{}
}
