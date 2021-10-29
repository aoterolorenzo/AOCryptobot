package helpers

import (
	"fmt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"strconv"
	"time"
)

type FileLogger struct { /* Your functions */
	telegramOutput bool
	telegramToken  string
	telegramChatId string
}

func NewFileLogger() *FileLogger {
	telegramOutput, _ := strconv.ParseBool(os.Getenv("telegramOutput"))
	var telegramToken string
	var telegramChatId string

	if telegramOutput {
		telegramToken = os.Getenv("telegramToken")
		if telegramToken == "" {
			log.Errorln("error: telegramOutput set to true but telegramToken parameter not found ")
			os.Exit(1)
		}
		telegramChatId = os.Getenv("telegramChatId")
		if telegramChatId == "" {
			log.Errorln("error: telegramOutput set to true but telegramChatId parameter not found ")
			os.Exit(1)
		}
	}

	return &FileLogger{
		telegramChatId: telegramChatId,
		telegramOutput: telegramOutput,
		telegramToken:  telegramToken,
	}
}

var defaultLogger *log.Logger
var Logger = *NewFileLogger()

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/bot_signal-trader/conf.env")
	if err != nil {
		log.Fatalln("Error loading go.env file", err)
	}
	logFile := os.Getenv("logFile")
	if logFile == "" {
		logFile = "bot.log"
	}
	//TODO: got log filename onto env var
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
	if l.telegramOutput {
		err := sendOnTelegramChannel(fmt.Sprintf("%s", args[0]), l.telegramToken, l.telegramChatId)
		if err != nil {
			log.Fatal(err)
		}
	}

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

func sendOnTelegramChannel(message string, token string, chatID string) error {

	b, err := tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		URL:    "",
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		return err
	}

	id, err := b.ChatByID(chatID)
	if err != nil {
		return err
	}
	_, err = b.Send(id, message)
	if err != nil {
		return err
	}

	return nil
}
