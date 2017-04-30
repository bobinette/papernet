package log

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

type logger struct {
	*logrus.Entry
}

func New(env string) Logger {
	l := logrus.New()

	if env == "prod" {
		l.Formatter = &logrus.JSONFormatter{}
		l.Level = logrus.InfoLevel
	} else {
		l.Formatter = &logrus.TextFormatter{}
		l.Level = logrus.DebugLevel
	}

	return logger{l.WithField("env", env)}
}

func (l logger) Print(args ...interface{}) {
	l.Println(args...)
}

func (l logger) Error(args ...interface{}) {
	l.Errorln(args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.Fatalln(args...)
}
