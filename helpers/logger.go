package helpers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"time"
)

type FileLogger struct { /* Your functions */
}

var defaultLogger *log.Logger
var Logger = FileLogger{}

func init() {
	//TODO: got log filename onto env var
	f, err := os.OpenFile("bot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	plainFormatter := new(PlainFormatter)
	plainFormatter.TimestampFormat = "2006-01-02 15:04:05"
	plainFormatter.LevelDesc = []string{"PANIC", "FATAL", "ERROR", "WARN", "INFO ", "DEBUG"}
	defaultLogger = log.New()
	defaultLogger.SetOutput(f)
	defaultLogger.SetFormatter(plainFormatter)
	defaultLogger.SetLevel(log.TraceLevel)
}

func (l *FileLogger) Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

func (l *FileLogger) Fatalln(args ...interface{}) {
	defaultLogger.Fatalln(args...)
}

func (l *FileLogger) Panicln(args ...interface{}) {
	defaultLogger.Panicln(args...)
}

func (l *FileLogger) Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}

func (l *FileLogger) Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
	sendOnTelegramChannel(fmt.Sprintf("%s", args[0]))
}

func (l *FileLogger) Traceln(args ...interface{}) {
	defaultLogger.Traceln(args...)
}

func (l *FileLogger) Debugln(args ...interface{}) {
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

func sendOnTelegramChannel(message string) {
	b, err := tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		URL: "",

		Token:  "1609124058:AAFUiuaD7Aop6BvZYIfxOy8-jNTaPV6xmCo",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	/*id , err :=*/
	b.ChatByID("-1001407056413")

	//b.Send(id, message)

}
