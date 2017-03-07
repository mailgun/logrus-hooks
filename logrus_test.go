package udploghook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
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
	r := logRecord{}
	r.fromFields(fields)
	c.Assert(r.Context["foo"], Equals, "bar")
	c.Assert(r.Context["bar"], Equals, 1)
	c.Assert(r.Context["bean"], Equals, true)
}

func (s *LogrusUDPSuite) TestFieldsToMapNested(c *C) {
	fields := logrus.Fields{
		"foo":                "bar",
		"http.url":           "http://example.com",
		"http.response.code": 200,
		"bar":                1,
		"bean":               true,
	}
	r := logRecord{}
	r.fromFields(fields)
	c.Assert(r.Context["foo"], Equals, "bar")
	c.Assert(r.Context["bar"], Equals, 1)
	c.Assert(r.Context["bean"], Equals, true)
	c.Assert(r.Context["http"].(map[string]interface{})["url"], Equals, "http://example.com")
	c.Assert(r.Context["http"].(map[string]interface{})["response"].(map[string]interface{})["code"], Equals, 200)
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

// If there is a logrus field `err` of `error` type, then it is parsed and
// Exc(Value|Type|Text) fields of the respective UDPlog record are populated.
func (s *LogrusUDPSuite) TestErrAsError(c *C) {
	cause := errors.New("foo")
	err := errors.Wrap(cause, "bar")
	fields := logrus.Fields{
		"err": err,
	}
	r := logRecord{}

	// When
	r.fromFields(fields)

	// Then
	c.Assert(r.ExcValue, Equals, "bar: foo")
	c.Assert(r.ExcType, Equals, "*errors.withStack")
	c.Assert(r.ExcText, Equals, `foo
github.com/mailgun/logrus-udplog.(*LogrusUDPSuite).TestErrAsError
	/Users/maxim/workspace/goworld/src/github.com/mailgun/logrus-udplog/logrus_test.go:73
runtime.call32
	/usr/local/go/src/runtime/asm_amd64.s:479
reflect.Value.call
	/usr/local/go/src/reflect/value.go:434
reflect.Value.Call
	/usr/local/go/src/reflect/value.go:302
gopkg.in/check%2ev1.(*suiteRunner).forkTest.func1
	/Users/maxim/workspace/goworld/src/gopkg.in/check.v1/check.go:772
gopkg.in/check%2ev1.(*suiteRunner).forkCall.func1
	/Users/maxim/workspace/goworld/src/gopkg.in/check.v1/check.go:666
runtime.goexit
	/usr/local/go/src/runtime/asm_amd64.s:2086
bar
github.com/mailgun/logrus-udplog.(*LogrusUDPSuite).TestErrAsError
	/Users/maxim/workspace/goworld/src/github.com/mailgun/logrus-udplog/logrus_test.go:74
runtime.call32
	/usr/local/go/src/runtime/asm_amd64.s:479
reflect.Value.call
	/usr/local/go/src/reflect/value.go:434
reflect.Value.Call
	/usr/local/go/src/reflect/value.go:302
gopkg.in/check%2ev1.(*suiteRunner).forkTest.func1
	/Users/maxim/workspace/goworld/src/gopkg.in/check.v1/check.go:772
gopkg.in/check%2ev1.(*suiteRunner).forkCall.func1
	/Users/maxim/workspace/goworld/src/gopkg.in/check.v1/check.go:666
runtime.goexit
	/usr/local/go/src/runtime/asm_amd64.s:2086`)
}

// If there is a logrus field `err` of any type but `error` type, then it is
// added to the context along with other fields.
func (s *LogrusUDPSuite) TestErrAsOther(c *C) {
	fields := logrus.Fields{
		"err": "foo",
	}
	r := logRecord{}

	// When
	r.fromFields(fields)

	// Then
	c.Assert(r.Context["err"], Equals, "foo")
}

func (s *LogrusUDPSuite) TestTIDAsString(c *C) {
	fields := logrus.Fields{
		"tid": "foo",
	}
	r := logRecord{}

	// When
	r.fromFields(fields)

	// Then
	c.Assert(r.TID, Equals, "foo")
}

func (s *LogrusUDPSuite) TestTIDAsOther(c *C) {
	fields := logrus.Fields{
		"tid": 10,
	}
	r := logRecord{}

	// When
	r.fromFields(fields)

	// Then
	c.Assert(r.Context["tid"], Equals, 10)
}
