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

// FluentHook is logrus hook for fluentd.
type FluentHook struct {
	// Fluent is actual fluentd logger.
	// If set, this logger is used for logging.
	// otherwise new logger is created every time.
	Fluent *fluent.Fluent

	host   string
	port   int
	levels []logrus.Level
	tag    *string
}

// New returns initialized logrus hook for fluentd with persistent fluentd logger.
func New(host string, port int) (*FluentHook, error) {
	fd, err := fluent.New(fluent.Config{FluentHost: host, FluentPort: port})
	if err != nil {
		return nil, err
	}

	return &FluentHook{
		levels: defaultLevels,
		Fluent: fd,
	}, nil
}

// NewHook returns initialized logrus hook for fluentd.
// (** deperecated: use New() **)
func NewHook(host string, port int) *FluentHook {
	return &FluentHook{
		host:   host,
		port:   port,
		levels: defaultLevels,
		tag:    nil,
	}
}

// Levels returns logging level to fire this hook.
func (hook *FluentHook) Levels() []logrus.Level {
	return hook.levels
}

// SetLevels sets logging level to fire this hook.
func (hook *FluentHook) SetLevels(levels []logrus.Level) {
	hook.levels = levels
}

// Tag returns custom static tag.
func (hook *FluentHook) Tag() string {
	if hook.tag == nil {
		return ""
	}

	return *hook.tag
}

// SetTag sets custom static tag to override tag in the message fields.
func (hook *FluentHook) SetTag(tag string) {
	hook.tag = &tag
}

// Fire is invoked by logrus and sends log to fluentd logger.
func (hook *FluentHook) Fire(entry *logrus.Entry) error {
	var logger *fluent.Fluent
	var err error

	switch {
	case hook.Fluent != nil:
		logger = hook.Fluent
	default:
		logger, err = fluent.New(fluent.Config{
			FluentHost: hook.host,
			FluentPort: hook.port,
		})
		if err != nil {
			return err
		}
		defer logger.Close()
	}

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
