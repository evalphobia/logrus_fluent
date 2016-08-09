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
	host   string
	port   int
	levels []logrus.Level
	tag    *string
}

func NewHook(host string, port int) *fluentHook {
	return &fluentHook{
		host:   host,
		port:   port,
		levels: defaultLevels,
		tag:    nil,
	}
}

func getTagAndDel(entry *logrus.Entry, data logrus.Fields) string {
	var v interface{}
	var ok bool
	if v, ok = data[TagField]; !ok {
		return entry.Message
	}

	var val string
	if val, ok = v.(string); !ok {
		return entry.Message
	}
	delete(data, TagField)
	return val
}

func setLevelString(entry *logrus.Entry, data logrus.Fields) {
	data["level"] = entry.Level.String()
}

func setMessage(entry *logrus.Entry, data logrus.Fields) {
	if _, ok := data[MessageField]; !ok {
		data[MessageField] = entry.Message
	}
}

func (hook *fluentHook) Fire(entry *logrus.Entry) error {
	logger, err := fluent.New(fluent.Config{
		FluentHost: hook.host,
		FluentPort: hook.port,
	})
	if err != nil {
		return err
	}
	defer logger.Close()

	// Create a map for passing to FluentD
	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}

	setLevelString(entry, data)
	var tag string
	if hook.tag == nil {
		tag = getTagAndDel(entry, data)
		if tag != entry.Message {
			setMessage(entry, data)
		}
	} else {
		tag = *hook.tag
	}

	fluentData := ConvertToValue(data, TagName)
	err = logger.PostWithTime(tag, entry.Time, fluentData)
	return err
}

func (hook *fluentHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *fluentHook) SetLevels(levels []logrus.Level) {
	hook.levels = levels
}

func (hook *fluentHook) Tag() string {
	return *hook.tag
}

func (hook *fluentHook) SetTag(tag string) {
	hook.tag = &tag
}
