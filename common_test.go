package ref

import (
	"github.com/hedzr/logex"
	"github.com/hedzr/logex/logx/logrus"
	"testing"
)

func initLogger(t *testing.T) (deferFn func()) {
	// build.New(build.NewLoggerConfigWith(true, "logex", "debug"))
	l := logrus.New("debug", false, true)
	//l.SetLevel(log.DebugLevel)

	l.Infof("123, level=%v", l.GetLevel())
	//log.Infof("456")

	return logex.CaptureLog(t).Release
}
