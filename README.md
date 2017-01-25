# UDPlog hook for Logrus

Simple Usage
```go
// Ship logs from logrus to UDPLog
hook, err := logrusUDP.NewLogHook("localhost", 514)
if err != nil {
    panic(err)
}

logrus.AddHook(hook)
logrus.Info("Your mother milk chicken for a living")
````
