package logger

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/sirupsen/logrus"
)

type CleanTextFormatter struct {
	Trace  string
	Caller string
}

func (f *CleanTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = 37
	case logrus.WarnLevel:
		levelColor = 33
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31
	default:
		levelColor = 36
	}

	timestamp := entry.Time.Format("2006/01/02 15:04:05")
	level := strings.ToUpper(entry.Level.String())

	var trace, caller string
	if f.Trace != "" || f.Caller != "" {
		trace = f.Trace
		caller = f.Caller
	} else {
		if t, ok := entry.Data["trace"].(string); ok {
			trace = t
		}
		if c, ok := entry.Data["caller"].(string); ok {
			caller = c
		}
	}

	header := fmt.Sprintf("\n%s \033[%dm%s %s\033[0m\n",
		timestamp, 36, trace, caller)

	delete(entry.Data, "trace")
	delete(entry.Data, "caller")
	delete(entry.Data, "service")

	var dataStr string
	if len(entry.Data) > 0 {
		if val, ok := entry.Data["error"]; ok && val != nil {
			dataStr += fmt.Sprintf("[err: %v] ", val)
		}
		if val, ok := entry.Data["trace_id"]; ok && val != nil {
			dataStr += fmt.Sprintf("[trace_id: %s] ", val)
		}
		if val, ok := entry.Data["data"]; ok {
			dataStr += fmt.Sprintf("[data: %s] ", utils.Stringify(val))
		}
	}

	body := fmt.Sprintf("\033[%dm[%s] \033[0m %s %s\n", levelColor, level, entry.Message, dataStr)

	return []byte(header + body), nil
}

type GlobalFieldsHook struct {
	AppName string
	AppEnv  string
	Trace   string
	Caller  string
}

func (h *GlobalFieldsHook) Levels() []logrus.Level {
	if h.AppEnv == "development" {
		return logrus.AllLevels
	}
	return []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (h *GlobalFieldsHook) Fire(entry *logrus.Entry) error {
	entry.Data["service"] = h.AppName

	if h.Caller != "" || h.Trace != "" {
		entry.Data["caller"] = h.Caller
		entry.Data["trace"] = h.Trace
		entry.Caller = nil

	} else if _, alreadySet := entry.Data["trace"]; alreadySet {
		entry.Caller = nil

	} else if entry.HasCaller() {
		entry.Data["caller"] = path.Base(entry.Caller.Function)
		entry.Data["trace"] = fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
	}

	return nil
}

type LoggerOptions struct {
	AppName string
	Trace   string
	Caller  string
}

func NewLogger(options ...*LoggerOptions) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	option := &LoggerOptions{}
	if len(options) == 1 && options[0] != nil {
		option = options[0]
	}

	if option.Trace == "" && option.Caller == "" {
		log.SetReportCaller(true)
	}

	if appEnv == "development" {
		log.SetFormatter(&CleanTextFormatter{
			Trace:  option.Trace,
			Caller: option.Caller,
		})
	} else {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	log.AddHook(&GlobalFieldsHook{
		AppName: option.AppName,
		AppEnv:  appEnv,
		Trace:   option.Trace,
		Caller:  option.Caller,
	})

	return log
}

var l *logrus.Logger

func Logger(options ...*LoggerOptions) *logrus.Logger {
	if l == nil {
		opt := &LoggerOptions{}
		if len(options) == 1 && options[0] != nil {
			opt = options[0]
		}

		if opt.AppName == "" {
			appEnv := os.Getenv("APP_ENV")
			if appEnv == "" {
				appEnv = "development"
			}
			opt.AppName = fmt.Sprintf("ai_assistant-%s", appEnv)
		}

		l = NewLogger(opt)
	}
	return l
}
