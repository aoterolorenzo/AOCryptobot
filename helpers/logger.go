package helpers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

type Logger struct { /* Your functions */
}

var defaultLogger *log.Logger

func init() {
	//TODO: got log filename onto env var
	f, err := os.OpenFile("bot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	plainFormatter := new(PlainFormatter)
	plainFormatter.TimestampFormat = "2006-01-02 15:04:05"
	plainFormatter.LevelDesc = []string{"PANC", "FATL", "E", "WARN", "INFO", "DEBG"}
	defaultLogger = log.New()
	defaultLogger.SetOutput(f)
	defaultLogger.SetFormatter(plainFormatter)
}

func (l *Logger) Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

func (l *Logger) Fatalln(args ...interface{}) {
	defaultLogger.Fatalln(args...)
}

func (l *Logger) Panicln(args ...interface{}) {
	defaultLogger.Panicln(args...)
}

func (l *Logger) Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}

func (l *Logger) Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
}

func (l *Logger) Traceln(args ...interface{}) {
	defaultLogger.Traceln(args...)
}

func (l *Logger) Debugln(args ...interface{}) {
	defaultLogger.Debugln(args...)
}

type PlainFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

func (f PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := fmt.Sprintf(entry.Time.Format(f.TimestampFormat))
	return []byte(fmt.Sprintf("%s %s %s\n", f.LevelDesc[entry.Level], timestamp, entry.Message)), nil
}
