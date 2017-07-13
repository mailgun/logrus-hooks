package common_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/mailgun/logrus-hooks/common"
	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func TestCommon(t *testing.T) { TestingT(t) }

type CommonTestSuite struct{}

var _ = Suite(&CommonTestSuite{})

func (s *CommonTestSuite) TestFieldsToMap(c *C) {
	fields := logrus.Fields{
		"foo":  "bar",
		"bar":  1,
		"bean": true,
	}
	r := common.LogRecord{}
	r.FromFields(fields)
	c.Assert(r.Context["foo"], Equals, "bar")
	c.Assert(r.Context["bar"], Equals, 1)
	c.Assert(r.Context["bean"], Equals, true)
}

func (s *CommonTestSuite) TestFieldsToMapNested(c *C) {
	fields := logrus.Fields{
		"foo":                "bar",
		"http.url":           "http://example.com",
		"http.response.code": 200,
		"bar":                1,
		"bean":               true,
	}
	r := common.LogRecord{}
	r.FromFields(fields)
	c.Assert(r.Context["foo"], Equals, "bar")
	c.Assert(r.Context["bar"], Equals, 1)
	c.Assert(r.Context["bean"], Equals, true)
	c.Assert(r.Context["http"].(map[string]interface{})["url"], Equals, "http://example.com")
	c.Assert(r.Context["http"].(map[string]interface{})["response"].(map[string]interface{})["code"], Equals, 200)
}

func (s *CommonTestSuite) TestRequestToMap(c *C) {
	body := bytes.NewBuffer([]byte("body"))
	req := httptest.NewRequest("POST", "http://example.com?param1=1&param2=2", body)
	req.Header.Add("User-Agent", "test-agent")
	req.ParseForm()

	result := common.RequestToMap(req)

	c.Assert(result["headers-json"], Equals, ""+
		"{\n"+
		"  \"User-Agent\": [\n"+
		"    \"test-agent\"\n"+
		"  ]\n"+
		"}")
	c.Assert(result["method"], Equals, "POST")
	c.Assert(result["ip"], Equals, "192.0.2.1:1234")
	c.Assert(result["params-json"], Equals, ""+
		"{\n"+
		"  \"param1\": [\n"+
		"    \"1\"\n"+
		"  ],\n"+
		"  \"param2\": [\n"+
		"    \"2\"\n"+
		"  ]\n"+
		"}")
	c.Assert(result["size"], Equals, int64(4))
	c.Assert(result["url"], Equals, "http://example.com?param1=1&param2=2")
	c.Assert(result["useragent"], Equals, "test-agent")
}

func (s *CommonTestSuite) TestAssertions(c *C) {
	fields := logrus.Fields{
		"tid":         "foo",
		"excValue":    "foo",
		"excType":     "foo",
		"excFuncName": "foo",
		"excLineno":   1,
		"excFileName": "foo",
		"category":    "foo-bar",
	}
	r := common.LogRecord{}

	// When
	r.FromFields(fields)

	// Then
	c.Assert(r.TID, Equals, "foo")
	c.Assert(r.ExcValue, Equals, "foo")
	c.Assert(r.ExcType, Equals, "foo")
	c.Assert(r.FuncName, Equals, "foo")
	c.Assert(r.LineNo, Equals, 1)
	c.Assert(r.FileName, Equals, "foo")
	c.Assert(r.Category, Equals, "foo-bar")
}

func (s *CommonTestSuite) TestNegativeAssertions(c *C) {
	fields := logrus.Fields{
		"tid":         10,
		"excValue":    10,
		"excType":     10,
		"excFuncName": 10,
		"excLineno":   "1",
		"excFileName": 10,
		"category":    10,
	}
	r := common.LogRecord{}

	// When
	r.FromFields(fields)

	// Then
	c.Assert(r.TID, Equals, "")
	c.Assert(r.ExcValue, Equals, "")
	c.Assert(r.ExcType, Equals, "")
	c.Assert(r.FuncName, Equals, "")
	c.Assert(r.LineNo, Equals, 0)
	c.Assert(r.FileName, Equals, "")
	c.Assert(r.Category, Equals, "")

	c.Assert(r.Context["tid"], Equals, 10)
	c.Assert(r.Context["excValue"], Equals, 10)
	c.Assert(r.Context["excType"], Equals, 10)
	c.Assert(r.Context["excFuncName"], Equals, 10)
	c.Assert(r.Context["excLineno"], Equals, "1")
	c.Assert(r.Context["excFileName"], Equals, 10)
}
