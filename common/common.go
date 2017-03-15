package common

import (
	"net/http"
	"runtime"
	"strings"

	"github.com/mailgun/holster/stack"
)

func ExpandNested(key string, value interface{}, dest map[string]interface{}) {
	if strings.ContainsRune(key, '.') {
		parts := strings.SplitN(key, ".", 2)
		// This nested value might already exist
		nested, isMap := dest[parts[0]].(map[string]interface{})
		if !isMap {
			// if not a map, overwrite current entry and make it a map
			nested = make(map[string]interface{})
			dest[parts[0]] = nested
		}
		ExpandNested(parts[1], value, nested)
	}
	switch value.(type) {
	case *http.Request:
		dest[key] = RequestToMap(value.(*http.Request))
	default:
		dest[key] = value
	}
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

// Returns the file, function and line number of the function that called logrus
func GetLogrusCaller() *stack.FrameInfo {
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

// Returns true if the key exists in the map
func Exists(haystack map[string]interface{}, needle string) bool {
	_, exists := haystack[needle]
	return exists
}
