[![Build Status](https://drone.io/github.com/evalphobia/wizard/status.png)](https://drone.io/github.com/evalphobia/wizard/latest)  [![Coverage Status](https://coveralls.io/repos/evalphobia/logrus_fluent/badge.svg?branch=master&service=github)](https://coveralls.io/github/evalphobia/logrus_fluent?branch=master)


# Fluentd Hook for Logrus <img src="http://i.imgur.com/hTeVwmJ.png" width="40" height="40" alt=":walrus:" class="emoji" title=":walrus:"/>

## Usage

```go
import (
	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_fluent"
)

func main() {
	hook := logrus_fluent.NewHook("localhost", 24224)
	hook.SetLevels([]logrus.Level{
		logrus.PanicLevel,
		logrus.ErrorLevel,
	})

	logrus.AddHook(hook)
}
```


## Special fields

Some logrus fields have a special meaning in this hook.

- `tag` is used as a fluentd tag. (if `tag` is omitted, Entry.Message is used as a fluentd tag)
