package internal

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger = logrus.New()

func init() {
	if os.Getenv("DEBUG") == "true" {
		Logger.SetLevel(logrus.DebugLevel)
	}
}
