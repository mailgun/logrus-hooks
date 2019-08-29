package logrus_hooks_test

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	hooks "github.com/mailgun/logrus-hooks"
	"github.com/mailgun/logrus-hooks/common"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewJSONFormater(t *testing.T) {
	var log = logrus.New()
	var b bytes.Buffer
	output := bufio.NewWriter(&b)
	log.SetOutput(output)

	log.SetFormatter(hooks.NewJSONFormater())

	log.Info("This is a test message")

	output.Flush()
	fmt.Println(b.String())

	rec := common.LogRecord{}

	err := easyjson.Unmarshal(b.Bytes(), &rec)
	assert.Nil(t, err)

	assert.Equal(t, "This is a test message", rec.Message)
	assert.Equal(t, "INFO", rec.LogLevel)
	assert.Contains(t, rec.FuncName, "TestNewJSONFormater")
	assert.Equal(t, "logrus", rec.Category)
}
