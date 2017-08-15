package kafkahook_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/mailgun/holster/errors"
	"github.com/mailgun/logrus-hooks/common"
	"github.com/mailgun/logrus-hooks/kafkahook"
	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func TestKafkaHook(t *testing.T) { TestingT(t) }

type KafkaHookTests struct {
	kafkahook *kafkahook.KafkaHook
	producer  *mocks.AsyncProducer
	log       *logrus.Logger
}

var _ = Suite(&KafkaHookTests{})

func (s *KafkaHookTests) SetUpTest(c *C) {
	var err error

	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true

	// Setup our AsyncProducer Mock
	s.producer = mocks.NewAsyncProducer(c, conf)
	s.producer.ExpectInputAndSucceed()

	// Create the hook
	s.kafkahook, err = kafkahook.New(kafkahook.Config{
		Producer: s.producer,
		Topic:    "test",
	})
	//s.kafkahook.SetDebug(true)
	c.Assert(err, IsNil)

	// Tell logrus about our udploghook
	s.log = logrus.New()
	s.log.Out = ioutil.Discard
	s.log.Hooks.Add(s.kafkahook)
}

func (s *KafkaHookTests) TestKafkaHookINFO(c *C) {
	s.log.Info("this is a test")

	req := GetMsg(s.producer)
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/kafkahook/kafkahook_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["funcName"].(string), Equals,
		"kafkahook_test.(*KafkaHookTests).TestKafkaHookINFO")
}

func (s *KafkaHookTests) TestKafkaHookExported(c *C) {
	logrus.SetOutput(ioutil.Discard)
	logrus.AddHook(s.kafkahook)
	logrus.Info("this is a test")

	req := GetMsg(s.producer)
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/kafkahook/kafkahook_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["funcName"].(string), Equals,
		"kafkahook_test.(*KafkaHookTests).TestKafkaHookExported")
}

func (s *KafkaHookTests) TestKafkaHookContext(c *C) {
	s.log.WithFields(logrus.Fields{
		"http.request": "http://localhost",
		"domain":       "example.com",
		"bean":         true,
		"bar":          1,
	}).Error("this is a test")

	req := GetMsg(s.producer)
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["lineno"], Equals, float64(91))
	c.Assert(req["logLevel"], Equals, "ERROR")
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/kafkahook/kafkahook_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["funcName"].(string), Equals,
		"kafkahook_test.(*KafkaHookTests).TestKafkaHookContext")

	context := req["context"].(map[string]interface{})
	c.Assert(context["http"].(map[string]interface{})["request"], Equals, "http://localhost")
	c.Assert(context["domain"], Equals, "example.com")
	c.Assert(context["bean"], Equals, true)
	c.Assert(context["bar"], Equals, float64(1))

	// these fields shouldn't exist unless we explicitly pass them
	c.Assert(common.Exists(context, "excType"), Equals, false)
	c.Assert(common.Exists(context, "excValue"), Equals, false)
	c.Assert(common.Exists(context, "excText"), Equals, false)
}

func (s *KafkaHookTests) TestKafkaHookRequest(c *C) {
	body := bytes.NewBuffer([]byte("body"))
	req := httptest.NewRequest("POST", "http://example.com?param1=1&param2=2", body)
	req.Header.Add("User-Agent", "test-agent")
	req.ParseForm()

	s.log.WithFields(logrus.Fields{
		"http": req,
	}).Error("Get Called")

	resp := GetMsg(s.producer)
	c.Assert(resp["message"], Equals, "Get Called")
	c.Assert(resp["logLevel"], Equals, "ERROR")

	http := resp["context"].(map[string]interface{})["http"].(map[string]interface{})
	c.Assert(http["headers-json"], Equals, ""+
		"{\n"+
		"  \"User-Agent\": [\n"+
		"    \"test-agent\"\n"+
		"  ]\n"+
		"}")
	c.Assert(http["method"], DeepEquals, "POST")
	c.Assert(http["ip"], Equals, "192.0.2.1:1234")
	c.Assert(http["params-json"], Equals, ""+
		"{\n"+
		"  \"param1\": [\n"+
		"    \"1\"\n"+
		"  ],\n"+
		"  \"param2\": [\n"+
		"    \"2\"\n"+
		"  ]\n"+
		"}")
	c.Assert(http["size"], Equals, float64(4))
	c.Assert(http["url"], Equals, "http://example.com?param1=1&param2=2")
	c.Assert(http["useragent"], Equals, "test-agent")
}

func (s *KafkaHookTests) TestFromErr(c *C) {
	cause := errors.New("foo")
	err := errors.Wrap(cause, "bar")

	s.log.WithFields(errors.ToLogrus(err)).Info("Info Called")

	req := GetMsg(s.producer)
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/kafkahook/kafkahook_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["lineno"], Equals, float64(154))
	c.Assert(req["funcName"], Equals, "kafkahook_test.(*KafkaHookTests).TestFromErr")
	c.Assert(req["excType"], Equals, "*errors.fundamental")
	c.Assert(req["excValue"], Equals, "bar: foo")
	c.Assert(strings.Contains(req["excText"].(string), "(*KafkaHookTests).TestFromErr"), Equals, true)
	c.Assert(strings.Contains(req["excText"].(string), "github.com/mailgun/logrus-hooks/kafkahook/kafkahook_test.go"), Equals, true)
}

func GetMsg(producer *mocks.AsyncProducer) map[string]interface{} {
	var result map[string]interface{}
	msg := <-producer.Successes()
	buf, _ := msg.Value.Encode()

	fmt.Printf("%s\n", buf)
	if err := json.Unmarshal(buf, &result); err != nil {
		fmt.Printf("json.Unmarshal() error: %s\n", err)
	}
	return result
}
