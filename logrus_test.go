package udploghook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func TestLogrusUDP(t *testing.T) { TestingT(t) }

type LogrusUDPSuite struct {
	server *Server
}

var _ = Suite(&LogrusUDPSuite{})

func (s *LogrusUDPSuite) TestFieldsToMap(c *C) {
	fields := logrus.Fields{
		"foo":  "bar",
		"bar":  1,
		"bean": true,
	}
	result := fieldsToMap(fields)
	c.Assert(result["foo"], Equals, "bar")
	c.Assert(result["bar"], Equals, 1)
	c.Assert(result["bean"], Equals, true)
}

func (s *LogrusUDPSuite) TestFieldsToMapNested(c *C) {
	fields := logrus.Fields{
		"foo":                "bar",
		"http.url":           "http://example.com",
		"http.response.code": 200,
		"bar":                1,
		"bean":               true,
	}
	result := fieldsToMap(fields)
	c.Assert(result["foo"], Equals, "bar")
	c.Assert(result["bar"], Equals, 1)
	c.Assert(result["bean"], Equals, true)
	c.Assert(result["http"].(map[string]interface{})["url"], Equals, "http://example.com")
	c.Assert(result["http"].(map[string]interface{})["response"].(map[string]interface{})["code"], Equals, 200)
}

func (s *LogrusUDPSuite) TestRequestToMap(c *C) {
	body := bytes.NewBuffer([]byte("body"))
	req := httptest.NewRequest("POST", "http://example.com?param1=1&param2=2", body)
	req.Header.Add("User-Agent", "test-agent")
	req.ParseForm()

	result := requestToMap(req)

	c.Assert(result["headers"], DeepEquals, http.Header{"User-Agent": []string{"test-agent"}})
	c.Assert(result["method"], Equals, "POST")
	c.Assert(result["ip"], Equals, "192.0.2.1:1234")
	c.Assert(result["params"], DeepEquals, url.Values{"param2": []string{"2"}, "param1": []string{"1"}})
	c.Assert(result["size"], Equals, int64(4))
	c.Assert(result["url"], Equals, "http://example.com?param1=1&param2=2")
	c.Assert(result["useragent"], Equals, "test-agent")
}
