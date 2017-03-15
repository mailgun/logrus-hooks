package kafkahook

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Sirupsen/logrus"
	"github.com/mailgun/holster/errors"
	"github.com/mailgun/holster/stack"
	"github.com/mailgun/logrus-hooks/common"
	"github.com/mailru/easyjson/jwriter"
)

type KafkaHook struct {
	producer sarama.AsyncProducer
	hostName string
	appName  string
	topic    string
	pid      int
	debug    bool
}

type Config struct {
	Endpoints []string
	Topic     string
	Producer  sarama.AsyncProducer
}

func New(conf Config) (*KafkaHook, error) {
	var err error

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	kafkaConfig.Producer.Flush.Frequency = 200 * time.Millisecond // Flush batches every 200ms

	// If the user failed to provide a producer create one
	if conf.Producer == nil {
		conf.Producer, err = sarama.NewAsyncProducer(conf.Endpoints, kafkaConfig)
		if err != nil {
			return nil, errors.Wrap(err, "kafka producer error")
		}
	}

	h := KafkaHook{
		producer: conf.Producer,
		topic:    conf.Topic,
	}

	if h.hostName, err = os.Hostname(); err != nil {
		h.hostName = "unknown_host"
	}
	h.appName = filepath.Base(os.Args[0])
	h.pid = os.Getpid()

	return &h, nil
}

func (h *KafkaHook) Fire(entry *logrus.Entry) error {
	var caller *stack.FrameInfo
	var err error

	caller = common.GetLogrusCaller()

	rec := &common.LogRecord{
		AppName:   h.appName,
		HostName:  h.hostName,
		LogLevel:  strings.ToUpper(entry.Level.String()),
		FileName:  caller.File,
		FuncName:  caller.Func,
		LineNo:    caller.LineNo,
		Message:   entry.Message,
		Context:   nil,
		Timestamp: common.Number(float64(entry.Time.UnixNano()) / 1000000000),
		PID:       h.pid,
	}
	rec.FromFields(entry.Data)

	var w jwriter.Writer
	rec.MarshalEasyJSON(&w)
	if w.Error != nil {
		return errors.Wrap(err, "while marshalling json")
	}
	buf := w.Buffer.BuildBytes()

	if h.debug {
		fmt.Printf("%s\n", string(buf))
	}

	err = h.sendKafka(buf)
	if err != nil {
		return errors.Wrap(err, "while sending to kafka")
	}
	return nil
}

func (h *KafkaHook) sendKafka(buf []byte) error {
	select {
	case h.producer.Input() <- &sarama.ProducerMessage{
		Topic: h.topic,
		Key:   nil,
		Value: sarama.ByteEncoder(buf),
	}:
	case err := <-h.producer.Errors():
		return err
	}
	return nil
}

// Given an io reader send the contents of the reader to udplog
func (h *KafkaHook) SendIO(input io.Reader) error {
	// Append our identifier
	buf := bytes.NewBuffer([]byte(""))
	_, err := buf.ReadFrom(input)
	if err != nil {
		return errors.Wrap(err, "while reading from IO")
	}

	if h.debug {
		fmt.Printf("%s\n", buf.String())
	}

	err = h.sendKafka(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "while sending to kafka")
	}
	return nil
}

// Levels returns the available logging levels.
func (h *KafkaHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func (h *KafkaHook) SetDebug(set bool) {
	h.debug = set
}
