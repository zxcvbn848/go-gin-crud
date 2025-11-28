package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Log 提供全域的 Logrus 實例
var Log = logrus.New()

func init() {
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	Log.SetOutput(os.Stdout)

	if level, ok := os.LookupEnv("LOG_LEVEL"); ok {
		parsedLevel, err := logrus.ParseLevel(level)
		if err == nil {
			Log.SetLevel(parsedLevel)
			return
		}
	}

	Log.SetLevel(logrus.InfoLevel)
}
