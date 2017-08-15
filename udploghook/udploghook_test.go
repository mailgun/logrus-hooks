package udploghook_test

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mailgun/holster/errors"
	"github.com/mailgun/logrus-hooks/common"
	"github.com/mailgun/logrus-hooks/udploghook"
	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func TestUDPLogHook(t *testing.T) { TestingT(t) }

type UDPLogHookTests struct {
	server     *udploghook.Server
	udploghook *udploghook.UDPHook
	log        *logrus.Logger
}

var _ = Suite(&UDPLogHookTests{})

func (s *UDPLogHookTests) SetUpTest(c *C) {
	var err error
	// Create a new udp server
	s.server, err = udploghook.NewServer("127.0.0.1", 0)
	c.Assert(err, IsNil)

	// Create the udploghook
	s.udploghook, err = udploghook.New(s.server.Host(), s.server.Port())
	s.udploghook.SetDebug(true)
	c.Assert(err, IsNil)

	// Tell logrus about our udploghook
	s.log = logrus.New()
	s.log.Out = ioutil.Discard
	s.log.Hooks.Add(s.udploghook)

}

func (s *UDPLogHookTests) TearDownTest(c *C) {
	s.server.Close()
}

func (s *UDPLogHookTests) TestUDPHookINFO(c *C) {
	s.log.Info("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/udploghook/udploghook_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["funcName"].(string), Equals,
		"udploghook_test.(*UDPLogHookTests).TestUDPHookINFO")
}

func (s *UDPLogHookTests) TestUDPHookExported(c *C) {
	logrus.SetOutput(ioutil.Discard)
	logrus.AddHook(s.udploghook)
	logrus.Info("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/udploghook/udploghook_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["funcName"].(string), Equals,
		"udploghook_test.(*UDPLogHookTests).TestUDPHookExported")
}

func (s *UDPLogHookTests) TestUDPHookContext(c *C) {
	s.log.WithFields(logrus.Fields{
		"http.request": "http://localhost",
		"domain":       "example.com",
		"bean":         true,
		"bar":          1,
	}).Error("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["logLevel"], Equals, "ERROR")
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/udploghook/udploghook_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["funcName"].(string), Equals,
		"udploghook_test.(*UDPLogHookTests).TestUDPHookContext")

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

func (s *UDPLogHookTests) TestUDPHookRequest(c *C) {
	body := bytes.NewBuffer([]byte("body"))
	req := httptest.NewRequest("POST", "http://example.com?param1=1&param2=2", body)
	req.Header.Add("User-Agent", "test-agent")
	req.ParseForm()

	s.log.WithFields(logrus.Fields{
		"http": req,
	}).Error("Get Called")

	resp := s.server.GetRequest()
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

func (s *UDPLogHookTests) TestFromErr(c *C) {
	cause := errors.New("foo")
	err := errors.Wrap(cause, "bar")

	s.log.WithFields(errors.ToLogrus(err)).Info("Info Called")

	req := s.server.GetRequest()
	c.Assert(strings.HasSuffix(req["filename"].(string),
		"github.com/mailgun/logrus-hooks/udploghook/udploghook_test.go"),
		Equals, true, Commentf(req["filename"].(string)))
	c.Assert(req["lineno"], Equals, float64(147))
	c.Assert(req["funcName"].(string), Equals,
		"udploghook_test.(*UDPLogHookTests).TestFromErr")
	c.Assert(req["excType"], Equals, "*errors.fundamental")
	c.Assert(req["excValue"], Equals, "bar: foo")
	c.Assert(strings.Contains(req["excText"].(string), "(*UDPLogHookTests).TestFromErr"), Equals, true)
	c.Assert(strings.Contains(req["excText"].(string), "github.com/mailgun/logrus-hooks/udploghook/udploghook_test.go:147"), Equals, true)
}

func (s *UDPLogHookTests) TestTIDAsString(c *C) {
	s.log.WithFields(logrus.Fields{"tid": "foo"}).Info("Info Called")
	req := s.server.GetRequest()
	c.Assert(req["tid"], Equals, "foo")
}

func (s *UDPLogHookTests) TestTIDAsOther(c *C) {
	s.log.WithFields(logrus.Fields{"tid": 10}).Info("Info Called")
	req := s.server.GetRequest()

	// Should not exist since tid is not a string
	c.Assert(common.Exists(req, "tid"), Equals, false)

	// Should instead be in context as an int
	context := req["context"].(map[string]interface{})
	c.Assert(context["tid"], Equals, float64(10))
}
