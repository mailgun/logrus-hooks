package udploghook

import (
	"net/http"
	"strings"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mailgun/holster/stack"
)

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

// Returns the context and stacktrace information for the underlying error as logrus.Fields{}
// returns empty logrus.Fields{} if err has no context or no stacktrace
//
// 	logrus.WithFields(udploghook.FromErr(err)).WithField("tid", 1).Error(err)
//
func FromErr(err error) logrus.Fields {
	type hasContext interface {
		Context() map[string]interface{}
	}

	result := logrus.Fields{
		"excValue": err.Error(),
		"excType":  fmt.Sprintf("%T", err),
		"excText":  fmt.Sprintf("%+v", err),
	}

	// Add the stack info if provided
	if cast, ok := err.(stack.HasStackTrace); ok {
		trace := cast.StackTrace()
		caller := stack.GetLastFrame(trace)
		result["excFuncName"] = caller.Func
		result["excLineno"] = caller.LineNo
		result["excFileName"] = caller.File
	}

	// Add context if provided
	child, ok := err.(hasContext)
	if !ok {
		return result
	}

	// Append the context map to our results
	for key, value := range child.Context() {
		result[key] = value
	}
	return result
}
