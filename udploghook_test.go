package udploghook_test

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/mailgun/logrus-udplog"
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"
)

func TestUDPLogHook(t *testing.T) { TestingT(t) }

type UPDLogHookTests struct {
	server *udploghook.Server
	hook   *udploghook.UDPHook
	log    *logrus.Logger
}

func exists(haystack map[string]interface{}, needle string) bool {
	_, exists := haystack[needle]
	return exists
}

var _ = Suite(&UPDLogHookTests{})

func (s *UPDLogHookTests) SetUpTest(c *C) {
	var err error
	// Create a new udp server
	s.server, err = udploghook.NewServer("127.0.0.1", 0)
	c.Assert(err, IsNil)

	// Create the hook
	s.hook, err = udploghook.New(s.server.Host(), s.server.Port())
	s.hook.SetDebug(true)
	c.Assert(err, IsNil)

	// Tell logrus about our hook
	s.log = logrus.New()
	s.log.Out = ioutil.Discard
	s.log.Hooks.Add(s.hook)

}

func (s *UPDLogHookTests) TearDownTest(c *C) {
	s.server.Close()
}

func (s *UPDLogHookTests) TestUDPHookINFO(c *C) {
	s.log.Info("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(path.Base(req["filename"].(string)), Equals, "udploghook_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookINFO"), Equals, true)
}

func (s *UPDLogHookTests) TestUDPHookExported(c *C) {
	logrus.SetOutput(ioutil.Discard)
	logrus.AddHook(s.hook)
	logrus.Info("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(path.Base(req["filename"].(string)), Equals, "udploghook_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookExported"), Equals, true)
}

func (s *UPDLogHookTests) TestUDPHookContext(c *C) {
	s.log.WithFields(logrus.Fields{
		"http.request": "http://localhost",
		"domain":       "example.com",
		"bean":         true,
		"bar":          1,
	}).Error("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["logLevel"], Equals, "ERROR")
	c.Assert(path.Base(req["filename"].(string)), Equals, "udploghook_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookContext"), Equals, true)

	context := req["context"].(map[string]interface{})
	c.Assert(context["http"].(map[string]interface{})["request"], Equals, "http://localhost")
	c.Assert(context["domain"], Equals, "example.com")
	c.Assert(context["bean"], Equals, true)
	c.Assert(context["bar"], Equals, float64(1))

	// these fields shouldn't exist unless we explicity pass them
	c.Assert(exists(context, "excType"), Equals, false)
	c.Assert(exists(context, "excValue"), Equals, false)
	c.Assert(exists(context, "excText"), Equals, false)
}

func (s *UPDLogHookTests) TestUDPHookRequest(c *C) {
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
	c.Assert(http["headers"], DeepEquals, map[string]interface{}{"User-Agent": []interface{}{"test-agent"}})
	c.Assert(http["method"], DeepEquals, "POST")
	c.Assert(http["ip"], Equals, "192.0.2.1:1234")
	c.Assert(http["params"], DeepEquals, map[string]interface{}{"param1": []interface{}{"1"}, "param2": []interface{}{"2"}})
	c.Assert(http["size"], Equals, float64(4))
	c.Assert(http["url"], Equals, "http://example.com?param1=1&param2=2")
	c.Assert(http["useragent"], Equals, "test-agent")
}

func (s *UPDLogHookTests) TestFromErr(c *C) {
	cause := errors.New("foo")
	err := errors.Wrap(cause, "bar")

	s.log.WithFields(udploghook.FromErr(err)).Info("Info Called")

	req := s.server.GetRequest()
	c.Assert(path.Base(req["filename"].(string)), Equals, "udploghook_test.go")
	c.Assert(req["lineno"], Equals, float64(130))
	c.Assert(req["funcName"], Equals, "TestFromErr()")
	c.Assert(req["excType"], Equals, "*errors.withStack")
	c.Assert(req["excValue"], Equals, "bar: foo")
	c.Assert(strings.Contains(req["excText"].(string), "(*UPDLogHookTests).TestFromErr"), Equals, true)
	c.Assert(strings.Contains(req["excText"].(string), "github.com/mailgun/logrus-udplog/udploghook_test.go:129"), Equals, true)
}

func (s *UPDLogHookTests) TestTIDAsString(c *C) {
	s.log.WithFields(logrus.Fields{"tid": "foo"}).Info("Info Called")
	req := s.server.GetRequest()
	c.Assert(req["tid"], Equals, "foo")
}

func (s *UPDLogHookTests) TestTIDAsOther(c *C) {
	s.log.WithFields(logrus.Fields{"tid": 10}).Info("Info Called")
	req := s.server.GetRequest()

	// Should not exist since tid is not a string
	c.Assert(exists(req, "tid"), Equals, false)

	// Should instead be in context as an int
	context := req["context"].(map[string]interface{})
	c.Assert(context["tid"], Equals, float64(10))
}
