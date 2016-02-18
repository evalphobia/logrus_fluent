package logrus_fluent

import (
	"github.com/Sirupsen/logrus"
	"github.com/fluent/fluent-logger-golang/fluent"
)

const (
	TagName      = "fluent"
	TagField     = "tag"
	MessageField = "message"
)

var defaultLevels = []logrus.Level{
	logrus.PanicLevel,
	logrus.FatalLevel,
	logrus.ErrorLevel,
	logrus.WarnLevel,
	logrus.InfoLevel,
}

type fluentHook struct {
	//host   string
	//port   int
	levels []logrus.Level
    Fluent *fluent.Fluent
}

func NewHook(host string, port int) *fluentHook {
    logger, _ := fluent.New(fluent.Config{FluentHost:host, FluentPort:port})
	return &fluentHook{
		//host:   host,
		//port:   port,
        Fluent: logger,
		levels: defaultLevels,
	}
}

func getTagAndDel(entry *logrus.Entry) string {
	var v interface{}
	var ok bool
	if v, ok = entry.Data[TagField]; !ok {
		return entry.Message
	}

	var val string
	if val, ok = v.(string); !ok {
		return entry.Message
	}
	delete(entry.Data, TagField)
	return val
}

func setLevelString(entry *logrus.Entry) {
	entry.Data["level"] = entry.Level.String()
}

func setMessage(entry *logrus.Entry) {
	if _, ok := entry.Data[MessageField]; !ok {
		entry.Data[MessageField] = entry.Message
	}
}

func (hook *fluentHook) Fire(entry *logrus.Entry) error {
	// logger, err := fluent.New(fluent.Config{
	// 	FluentHost: hook.host,
	// 	FluentPort: hook.port,
	// })
	// if err != nil {
	// 	return err
	// }
	// defer logger.Close()
    logger := hook.Fluent
	setLevelString(entry)
	tag := getTagAndDel(entry)
	if tag != entry.Message {
		setMessage(entry)
	}

	data := ConvertToValue(entry.Data, TagName)
	err := logger.PostWithTime(tag, entry.Time, data)
	return err
}

func (hook *fluentHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *fluentHook) SetLevels(levels []logrus.Level) {
	hook.levels = levels
}
