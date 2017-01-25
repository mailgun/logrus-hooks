# What is this?

A Logrus Hook for sending log info to [UDPLog](https://github.com/mochi/udplog)

To avoid problems with elastic search we limit the fields ES accepts. Because
we still want the flexiblity of adding context to our log messages that logrus
provides any context information provided by calls to `logrus.WithFields()` is
marshalled into json and sent to ES in a field called ```context```


# Installation
```bash
go get github.com/mailgun/logrus-udplog
```

# Usage
```go
// Ship logs from logrus to UDPLog
hook, err := logrusUDP.NewLogHook("localhost", 514)
if err != nil {
    panic(err)
}

logrus.AddHook(hook)

/*
{  
   "appname":"logrus-udplog.test",
   "hostname":"ljtrrn32.rackspace.corp",
   "logLevel":"INFO",
   "filename":"/Users/thrawn/Development/go/src/github.com/thrawn01/logrusUDP/logrus_test.go",
   "funcName":"github.com/thrawn01/logrus-udplog_test.(*LogrusUDPSuite).TestUDPHookContext",
   "lineno":81,
   "context":"",
   "message":"Your mother milk chicken for a living",
   "timestamp":1.4853036601508412e+09
}{  
*/
logrus.Info("Your mother milk chicken for a living")

/*
{  
   "appname":"logrus-udplog.test",
   "hostname":"ljtrrn32.rackspace.corp",
   "logLevel":"INFO",
   "filename":"/Users/thrawn/Development/go/src/github.com/thrawn01/logrusUDP/logrus_test.go",
   "funcName":"github.com/thrawn01/logrus-udplog_test.(*LogrusUDPSuite).TestUDPHookContext",
   "lineno":81,
   "context":"{\"domain\":\"example.com\",\"http.request\":\"http://localhost\"}",
   "message":"Your mother milk chicken for a living",
   "timestamp":1.4853036601508412e+09
}{  
*/
log.WithFields(logrus.Fields{
	"http.request": "http://localhost",
	"domain":       "example.com",
}).Info("Your mother milk chicken for a living")
````
