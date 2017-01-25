package logrusUDP_test

import (
	"io/ioutil"
	"strings"

	"path"

	"testing"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/thrawn01/logrusUDP"
	. "gopkg.in/check.v1"
)

func TestLogrusUDP(t *testing.T) { TestingT(t) }

type LogrusUDPSuite struct {
	server *logrusUDP.Server
}

var _ = Suite(&LogrusUDPSuite{})

func (s *LogrusUDPSuite) SetUpTest(c *C) {
	var err error
	s.server, err = logrusUDP.NewServer("127.0.0.1", 0)
	c.Assert(err, IsNil)
}

func (s *LogrusUDPSuite) TearDownTest(c *C) {
	s.server.Close()
}

func (s *LogrusUDPSuite) TestUDPHookINFO(c *C) {
	hook, err := logrusUDP.NewLogHook(s.server.Host(), s.server.Port())
	c.Assert(err, IsNil)

	log := logrus.New()
	log.Out = ioutil.Discard
	log.Hooks.Add(hook)
	log.Info("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(path.Base(req["filename"].(string)), Equals, "logrus_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookINFO"), Equals, true)
}

func (s *LogrusUDPSuite) TestUDPHookExported(c *C) {
	hook, err := logrusUDP.NewLogHook(s.server.Host(), s.server.Port())
	c.Assert(err, IsNil)

	logrus.SetOutput(ioutil.Discard)
	logrus.AddHook(hook)
	logrus.Info("this is a test")

	req := s.server.GetRequest()
	c.Assert(req["message"], Equals, "this is a test")
	c.Assert(req["context"], Equals, nil)
	c.Assert(req["logLevel"], Equals, "INFO")
	c.Assert(path.Base(req["filename"].(string)), Equals, "logrus_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookExported"), Equals, true)
}

func (s *LogrusUDPSuite) TestUDPHookContext(c *C) {
	hook, err := logrusUDP.NewLogHook(s.server.Host(), s.server.Port())
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
	c.Assert(path.Base(req["filename"].(string)), Equals, "logrus_test.go")
	c.Assert(strings.Contains(req["funcName"].(string), "TestUDPHookContext"), Equals, true)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(req["context"].(string)), &result)
	c.Assert(err, IsNil)

	c.Assert(result["http.request"], Equals, "http://localhost")
	c.Assert(result["domain"], Equals, "example.com")
}
