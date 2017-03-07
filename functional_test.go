package udploghook

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	. "gopkg.in/check.v1"
)

type FunctionalTest struct {
	server *Server
}

var _ = Suite(&FunctionalTest{})

func (s *FunctionalTest) SetUpTest(c *C) {
	var err error
	s.server, err = NewServer("127.0.0.1", 0)
	c.Assert(err, IsNil)
}

func (s *FunctionalTest) TearDownTest(c *C) {
	s.server.Close()
}

func (s *FunctionalTest) TestUDPHookINFO(c *C) {
	hook, err := New(s.server.Host(), s.server.Port())
	c.Assert(err, IsNil)

	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.Info("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(path.Base(req["filename"].(string)), Equals, "functional_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookINFO"), Equals, true)
}

func (s *FunctionalTest) TestUDPHookExported(c *C) {
	hook, err := New(s.server.Host(), s.server.Port())
	c.Assert(err, IsNil)

	logrus.SetOutput(ioutil.Discard)
	logrus.AddHook(hook)
	logrus.Info("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(path.Base(req["filename"].(string)), Equals, "functional_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookExported"), Equals, true)
}

func (s *FunctionalTest) TestUDPHookContext(c *C) {
	hook, err := New(s.server.Host(), s.server.Port())
	c.Assert(err, IsNil)

	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.WithFields(logrus.Fields{
		"http.request": "http://localhost",
		"domain":       "example.com",
	}).Error("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["logLevel"], Equals, "ERROR")
	c.Assert(path.Base(req["filename"].(string)), Equals, "functional_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookContext"), Equals, true)

	context := req["context"].(map[string]interface{})
	c.Assert(context["domain"], Equals, "example.com")
	c.Assert(context["http"].(map[string]interface{})["request"], Equals, "http://localhost")
}

func (s *FunctionalTest) TestUDPHookRequest(c *C) {
	hook, err := New(s.server.Host(), s.server.Port())
	c.Assert(err, IsNil)

	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)

	body := bytes.NewBuffer([]byte("body"))
	req := httptest.NewRequest("POST", "http://example.com?param1=1&param2=2", body)
	req.Header.Add("User-Agent", "test-agent")
	req.ParseForm()

	log.WithFields(logrus.Fields{
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
