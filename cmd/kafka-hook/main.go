package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Shopify/sarama"
	"github.com/mailgun/logrus-hooks/common"
	"github.com/mailgun/logrus-hooks/kafkahook"
	"github.com/sirupsen/logrus"
	"github.com/thrawn01/args"
)

func checkErr(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s - %s\n", msg, err)
		os.Exit(1)
	}
}

func main() {
	desc := args.Dedent(`CLI for kafkahook

	Examples:
	   export KAFKAHOOK_ENDPOINTS=kafka1:9092,kafka2:9092

	   Send a log message to kafkahook
	   $ kafka-hook "This is a message line"

	   Send a log message with 'other' fields attached
	   $ kafka-hook "This is a message line" -o "http.request=http://foo/bar" -o "foo=bar"

	   Send custom JSON to kafkahook
	   $ echo -e '{"custom":"json"}' | kafkahook -v`)

	parser := args.NewParser(args.EnvPrefix("KAFKAHOOK_"), args.Desc(desc, args.IsFormated))
	parser.AddOption("--verbose").Alias("-v").IsTrue().Help("be verbose")
	parser.AddArgument("message").Help("the message to log")
	parser.AddOption("--other").Alias("-o").IsStringMap().
		Help("additional fields to be sent in the kafkahook payload")
	parser.AddOption("--endpoints").IsStringSlice().Env("ENDPOINTS").
		Default("localhost:9092").Help("list of endpoints where kafka is listening")
	parser.AddOption("--topic").Alias("-t").Env("TOPIC").Default("udplog").
		Help("the topic to publish the log messge too")

	// Parser and set global options
	opts := parser.ParseSimple(nil)

	hook, err := kafkahook.New(kafkahook.Config{
		Endpoints: opts.StringSlice("endpoints"),
		Topic:     opts.String("topic"),
	})
	checkErr("KafkaHook Error", err)

	if opts.Bool("verbose") {
		hook.SetDebug(true)
	}

	// if stdin has an open pipe send the raw input from stdin to kafkahook
	if args.IsCharDevice(os.Stdin) {
		err := hook.SendIO(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Send via stdin failed - %s", err.Error())
		}
		os.Exit(0)
	} else {
		if !opts.IsSet("message") {
			fmt.Fprintf(os.Stderr, "'message' is required if no pipe from stdin")
			os.Exit(1)
		}
	}

	logrus.SetOutput(ioutil.Discard)
	logrus.AddHook(hook)

	if opts.IsSet("other") {
		logrus.WithFields(common.ToFields(opts.StringMap("other"))).Info(opts.String("message"))
	} else {
		logrus.Info(opts.String("message"))
	}

	// Flush the message to kafka and close the producer
	err = hook.Close()

	// Print any errors we received
	if err != nil {
		errors, _ := err.(sarama.ProducerErrors)
		for _, error := range errors {
			fmt.Fprintf(os.Stderr, "kafka-hook: %s\n", error)
		}
		os.Exit(1)
	}
	os.Exit(0)
}
