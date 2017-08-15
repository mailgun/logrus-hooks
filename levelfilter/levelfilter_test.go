package levelfilter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/mailgun/logrus-hooks/kafkahook"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
	"strings"
)

func Test(t *testing.T) { TestingT(t) }

type LevelFilterSuite struct{}

var _ = Suite(&LevelFilterSuite{})

// Only levels with higher or the same severity are advertised by a level
// filtering hook.
func (s *LevelFilterSuite) TestLevels(c *C) {
	for i, tc := range []struct {
		original []logrus.Level
		level    logrus.Level
		filtered []logrus.Level
	}{
		0: {
			original: logrus.AllLevels,
			level:    logrus.PanicLevel,
			filtered: []logrus.Level{logrus.PanicLevel},
		}, 1: {
			original: logrus.AllLevels,
			level:    logrus.FatalLevel,
			filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel},
		}, 2: {
			original: logrus.AllLevels,
			level:    logrus.ErrorLevel,
			filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel},
		}, 3: {
			original: logrus.AllLevels,
			level:    logrus.WarnLevel,
			filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel},
		}, 4: {
			original: logrus.AllLevels,
			level:    logrus.InfoLevel,
			filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel},
		}, 5: {
			original: logrus.AllLevels,
			level:    logrus.DebugLevel,
			filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel},
		}, 6: {
			// Original missing levels stay missing when filtered.
			original: []logrus.Level{logrus.PanicLevel, logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel},
			level:    logrus.InfoLevel,
			filtered: []logrus.Level{logrus.PanicLevel, logrus.WarnLevel, logrus.InfoLevel},
		}, 7: {
			// It is ok to specify a level missing in the original list for filtering.
			original: []logrus.Level{logrus.PanicLevel, logrus.WarnLevel, logrus.DebugLevel},
			level:    logrus.InfoLevel,
			filtered: []logrus.Level{logrus.PanicLevel, logrus.WarnLevel},
		}} {
		fmt.Printf("Test case #%d\n", i)

		fakeHook := newFakeHook(tc.original)
		lf := New(fakeHook, tc.level)

		c.Assert(lf.Levels(), DeepEquals, tc.filtered)
	}
}

// Fire calls to a level filter hook are forwarded to the underlying hook.
func (s *LevelFilterSuite) TestFire(c *C) {
	fakeHook := newFakeHook(logrus.AllLevels)
	lf := New(fakeHook, logrus.PanicLevel)
	c.Assert(fakeHook.entries, DeepEquals, []*logrus.Entry(nil))

	// When
	e1 := &logrus.Entry{Message: "1"}
	lf.Fire(e1)
	e2 := &logrus.Entry{Message: "2"}
	lf.Fire(e2)
	e3 := &logrus.Entry{Message: "3"}
	lf.Fire(e3)

	// Then
	c.Assert(fakeHook.entries, DeepEquals, []*logrus.Entry{e1, e2, e3})
}

func (s *LevelFilterSuite) TestCallerInfoWithError(c *C) {
	kafkaHook, msgGetter := newKafkaHook(c)
	log := logrus.New()
	log.Out = ioutil.Discard

	lh := New(kafkaHook, logrus.InfoLevel)
	log.Hooks.Add(lh)

	err := errors.New("Kaboom!")

	// When
	log.WithError(err).Error("Error Called")

	// Then
	req := msgGetter()
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/levelfilter/levelfilter_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["funcName"], Equals,
		"levelfilter.(*LevelFilterSuite).TestCallerInfoWithError")
}

func (s *LevelFilterSuite) TestCallerInfo(c *C) {
	kafkaHook, msgGetter := newKafkaHook(c)
	log := logrus.New()
	log.Out = ioutil.Discard

	lh := New(kafkaHook, logrus.InfoLevel)
	log.Hooks.Add(lh)

	// When
	log.Info("Info Called")

	// Then
	req := msgGetter()
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/levelfilter/levelfilter_test.go"), Equals,
		true, Commentf(req["filename"].(string)))
	c.Assert(req["funcName"].(string), Equals,
		"levelfilter.(*LevelFilterSuite).TestCallerInfo")
}

type fakeHook struct {
	levels  []logrus.Level
	entries []*logrus.Entry
}

func newFakeHook(levels []logrus.Level) *fakeHook {
	return &fakeHook{levels: levels}
}

func (h *fakeHook) Levels() []logrus.Level {
	return h.levels
}

func (h *fakeHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry)
	return nil
}

func newKafkaHook(c *C) (logrus.Hook, func() map[string]interface{}) {
	// Setup our AsyncProducer Mock.
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	prod := mocks.NewAsyncProducer(c, saramaCfg)
	prod.ExpectInputAndSucceed()

	// Create the kafka hook.
	kafkaHook, err := kafkahook.New(kafkahook.Config{
		Producer: prod,
		Topic:    "test",
	})
	c.Assert(err, IsNil)

	// Create message getter.
	msgGetter := func() map[string]interface{} {
		return getMsg(prod)
	}

	return kafkaHook, msgGetter
}

func getMsg(producer *mocks.AsyncProducer) map[string]interface{} {
	var result map[string]interface{}
	msg := <-producer.Successes()
	buf, _ := msg.Value.Encode()

	fmt.Printf("%s\n", buf)
	if err := json.Unmarshal(buf, &result); err != nil {
		fmt.Printf("json.Unmarshal() error: %s\n", err)
	}
	return result
}
