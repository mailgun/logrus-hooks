package logrus_hooks

import "github.com/mailgun/logrus-hooks/common"

func NewJSONFormater() *common.JSONFormater {
	return common.DefaultFormatter
}
