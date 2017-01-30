package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mailgun/logrus-udplog"
	"github.com/thrawn01/args"
)

func checkErr(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s - %s\n", msg, err)
		os.Exit(1)
	}
}

func ToFields(items map[string]string) logrus.Fields {
	result := logrus.Fields{}
	for key, value := range items {
		result[key] = value
	}
	return result
}

func main() {
	desc := args.Dedent(`CLI for udplog

	Examples:
	   export UDPLOG_ADDRESS=localhost:55647

	   Send a log message to udplog
	   $ udplog "This is a message line"

	   Send a log message with 'other' fields attached
	   $ udplog "This is a message line" -o "http.request=http://foo/bar" -o "foo=bar"`)

	parser := args.NewParser(args.EnvPrefix("UDPLOG_"), args.Desc(desc, args.IsFormated))
	parser.AddOption("--verbose").Alias("-v").IsTrue().Help("be verbose")
	parser.AddOption("--other").IsStringMap().Help("additional fields to be sent in the udplog payload")
	parser.AddArgument("message").Required().Help("the message to log")
	parser.AddOption("--address").Env("ADDRESS").Default("localhost:55647").
		Help("address where udplog is listening")

	// Parser and set global options
	opts := parser.ParseArgsSimple(nil)

	parts := strings.Split(opts.String("address"), ":")
	if len(parts) != 2 {
		fmt.Fprint(os.Stderr, "address '%s' invalid must be in format 'host:port'", opts.String("address"))
		os.Exit(1)
	}

	port, err := strconv.ParseInt(parts[1], 10, 64)
	checkErr("ParseInt Port Number", err)

	hook, err := udploghook.New(parts[0], int(port))
	checkErr("NowLogHook Error", err)

	if opts.Bool("verbose") {
		hook.SetDebug(true)
	}

	logrus.SetOutput(ioutil.Discard)
	logrus.AddHook(hook)

	if opts.IsSet("other") {
		logrus.WithFields(ToFields(opts.StringMap("other"))).Info(opts.String("message"))
	} else {
		logrus.Info(opts.String("message"))
	}
	os.Exit(0)
}
