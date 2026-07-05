package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type CleanTextFormatter struct{}

func (f *CleanTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	level := strings.ToUpper(entry.Level.String())

	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = 37 // White
	case logrus.WarnLevel:
		levelColor = 33 // Yellow
		level = "WARN"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31 // Red
	default:
		levelColor = 36 // Cyan
	}

	timestamp := entry.Time.Format("15:04:05")

	var trace, caller string
	if t, ok := entry.Data["trace"].(string); ok {
		trace = t
	}
	if c, ok := entry.Data["caller"].(string); ok {
		caller = c
	}

	duration := ""
	if d, ok := entry.Data["duration"]; ok {
		if dur, isDur := d.(time.Duration); isDur && dur > 10*time.Microsecond {
			duration = fmt.Sprintf(" [%.2fms]", float64(dur.Microseconds())/1000)
		} else if durInt, isInt := d.(int64); isInt && durInt > 10 {
			duration = fmt.Sprintf(" [%.2fms]", float64(durInt)/1000)
		}
	}

	rows := ""
	if r, ok := entry.Data["rows"].(int64); ok {
		rows = fmt.Sprintf(" [rows: %d]", r)
	}

	header := fmt.Sprintf("%s \033[36m%s %s\033[0m%s%s\n",
		timestamp, trace, caller, duration, rows)

	var sb strings.Builder
	for key, val := range entry.Data {
		switch key {
		case "service", "req_id", "trace", "caller", "duration", "rows":
			continue
		default:
			sb.WriteString(formatStructuredField(key, val))
		}
	}

	body := fmt.Sprintf("\033[%dm%-9s\033[0m%s\n%s\n", levelColor, ("[" + level + "]"), entry.Message, sb.String())

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
		log.SetFormatter(&CleanTextFormatter{})
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

var (
	l    *logrus.Logger
	once sync.Once
)

func Logger(options ...*LoggerOptions) *logrus.Logger {
	once.Do(func() {
		opt := &LoggerOptions{}
		if len(options) == 1 && options[0] != nil {
			opt = options[0]
		}

		if opt.AppName == "" {
			appEnv := os.Getenv("APP_ENV")
			if appEnv == "" {
				appEnv = "development"
			}
			opt.AppName = fmt.Sprintf("%s-%s", os.Getenv("APP_NAME"), appEnv)
		}

		l = NewLogger(opt)
	})
	return l
}

func formatStructuredField(key string, value any) string {
	// Render error values via .Error() to avoid confusing JSON output from
	// driver-specific error types (e.g. pgx produces protocol-level fields).
	if e, ok := value.(error); ok {
		value = e.Error()
	}

	indentBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("         \033[35m%s\033[0m: %v\n", key, value)
	}

	indentStr := string(indentBytes)

	if !strings.Contains(indentStr, "\n") {
		return fmt.Sprintf("         \033[35m%s\033[0m: %s\n", key, indentStr)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "         \033[35m%s\033[0m:\n", key)

	for line := range strings.SplitSeq(indentStr, "\n") {
		fmt.Fprintf(&sb, "          %s\n", line)
	}

	return sb.String()
}
