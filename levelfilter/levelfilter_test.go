package levelfilter

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type LevelFilterSuite struct{}

var _ = Suite(&LevelFilterSuite{})

// Only levels with higher or the same severity are advertised by a level
// filtering hook.
func (s *LevelFilterSuite) TestLevels(c *C) {
	for i, tc := range []struct {
		original []logrus.Level
		level    logrus.Level
		filtered []logrus.Level
	}{
		0: {
		original: logrus.AllLevels,
		level:    logrus.PanicLevel,
		filtered: []logrus.Level{logrus.PanicLevel},
	},
		1: {
		original: logrus.AllLevels,
		level:    logrus.FatalLevel,
		filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel},
	}, 2: {
		original: logrus.AllLevels,
		level:    logrus.ErrorLevel,
		filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel},
	}, 3: {
		original: logrus.AllLevels,
		level:    logrus.WarnLevel,
		filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel},
	}, 4: {
		original: logrus.AllLevels,
		level:    logrus.InfoLevel,
		filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel},
	}, 5: {
		original: logrus.AllLevels,
		level:    logrus.DebugLevel,
		filtered: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel},
	}, 6: {
		// Original missing levels stay missing when filtered.
		original: []logrus.Level{logrus.PanicLevel, logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel},
		level:    logrus.InfoLevel,
		filtered: []logrus.Level{logrus.PanicLevel, logrus.WarnLevel, logrus.InfoLevel},
	}, 7: {
		// It is ok to specify a level missing in the original list for filtering.
		original: []logrus.Level{logrus.PanicLevel, logrus.WarnLevel, logrus.DebugLevel},
		level:    logrus.InfoLevel,
		filtered: []logrus.Level{logrus.PanicLevel, logrus.WarnLevel},
	}} {
		fmt.Printf("Test case #%d", i)

		fakeHook := newFakeHook(tc.original)
		lf := New(fakeHook, tc.level)

		c.Assert(lf.Levels(), DeepEquals, tc.filtered)
	}
}

// Fire calls to a level filter hook are forwarded to the underlying hook.
func (s *LevelFilterSuite) TestFire(c *C) {
	fakeHook := newFakeHook(logrus.AllLevels)
	lf := New(fakeHook, logrus.PanicLevel)
	c.Assert(fakeHook.entries, DeepEquals, []*logrus.Entry(nil))

	// When
	e1 := &logrus.Entry{Message: "1"}
	lf.Fire(e1)
	e2 := &logrus.Entry{Message: "2"}
	lf.Fire(e2)
	e3 := &logrus.Entry{Message: "3"}
	lf.Fire(e3)

	// Then
	c.Assert(fakeHook.entries, DeepEquals, []*logrus.Entry{e1, e2, e3})
}

type fakeHook struct {
	levels []logrus.Level
	entries []*logrus.Entry
}

func newFakeHook(levels []logrus.Level) *fakeHook {
	return &fakeHook{levels: levels}
}

func (h *fakeHook) Levels() []logrus.Level {
	return h.levels
}

func (h *fakeHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry)
	return nil
}
