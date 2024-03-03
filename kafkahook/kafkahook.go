package kafkahook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/mailgun/holster/v3/errors"
	"github.com/mailgun/holster/v3/setter"
	"github.com/mailgun/logrus-hooks/common"
	"github.com/sirupsen/logrus"
)

const bufferSize = 150

type KafkaHook struct {
	produce chan []byte
	conf    Config
	debug   bool

	// Sync stuff
	wg   sync.WaitGroup
	once sync.Once
}

type Config struct {
	Endpoints []string
	Topic     string
	Producer  sarama.AsyncProducer
	Formatter logrus.Formatter
}

func NewWithContext(ctx context.Context, conf Config) (hook *KafkaHook, err error) {
	done := make(chan struct{})
	go func() {
		hook, err = New(conf)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return hook, fmt.Errorf(
			"connect timeout while connecting to kafka peers %s",
			strings.Join(conf.Endpoints, ","),
		)
	case <-done:
		return hook, err
	}
}

func New(conf Config) (*KafkaHook, error) {
	// If no formatter defined, use the default
	setter.SetDefault(&conf.Formatter, common.DefaultFormatter)
	var err error

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy
	kafkaConfig.Producer.Flush.Frequency = 200 * time.Millisecond
	kafkaConfig.Producer.Retry.Backoff = 10 * time.Second
	kafkaConfig.Producer.Retry.Max = 6
	kafkaConfig.Producer.Return.Errors = true

	// If the user failed to provide a producer create one
	if conf.Producer == nil {
		conf.Producer, err = sarama.NewAsyncProducer(conf.Endpoints, kafkaConfig)
		if err != nil {
			return nil, errors.Wrap(err, "kafka producer error")
		}
	}

	h := KafkaHook{
		produce: make(chan []byte, bufferSize),
		conf:    conf,
	}

	h.wg.Add(1)
	go func() {
		for {
			select {
			case err := <-conf.Producer.Errors():
				msg, _ := err.Msg.Value.Encode()
				_, _ = fmt.Fprintf(os.Stderr, "[kafkahook] produce error '%s' for: %s\n", err.Err, string(msg))

			case buf, ok := <-h.produce:
				if !ok {
					if err := conf.Producer.Close(); err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "[kafkahook] producer close error: %s\n", err)
					}
					h.wg.Done()
					return
				}
				conf.Producer.Input() <- &sarama.ProducerMessage{
					Value: sarama.ByteEncoder(buf),
					Topic: conf.Topic,
					Key:   nil,
				}
			}
		}
	}()
	return &h, nil
}

func (h *KafkaHook) Fire(entry *logrus.Entry) error {
	buf, err := h.conf.Formatter.Format(entry)
	if err != nil {
		return errors.Wrap(err, "while formatting entry")
	}

	if h.debug {
		fmt.Printf("%s\n", string(buf))
	}

	err = h.sendKafka(buf)
	if err != nil {
		return errors.Wrap(err, "while sending")
	}
	return nil
}

func (h *KafkaHook) sendKafka(buf []byte) error {
	select {
	case h.produce <- buf:
	default:
		// If the producer input channel buffer is full, then we better drop
		// a log record than block program execution.
		_, _ = fmt.Fprintf(os.Stderr, "[kafkahook] buffer overflow: %s\n", string(buf))
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
		return errors.Wrap(err, "while sending")
	}
	return nil
}

// Levels returns the available logging levels.
func (h *KafkaHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *KafkaHook) SetDebug(set bool) {
	h.debug = set
}

// Close the kakfa producer and flush any remaining logs
func (h *KafkaHook) Close() error {
	var err error
	h.once.Do(
		func() {
			close(h.produce)
			h.wg.Wait()
		},
	)
	return err
}
