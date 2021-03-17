package task

import (
	"fmt"
	"strings"
)

const (
	RunnableType Type = 0
	DeferType    Type = 1
	TriggerType  Type = 2
	DependType   Type = 3
)

func GenerateUniqueKey(number int, behaviors ...Type) string {
	var key = fmt.Sprintf("key[%d]", number)
	for _, behavior := range behaviors {
		switch behavior {
		case RunnableType:
			key = fmt.Sprintf("%s-%s", key, "run")
			break
		case DeferType:
			key = fmt.Sprintf("%s-%s", key, "defer")
			break
		case TriggerType:
			key = fmt.Sprintf("%s-%s", key, "trigger")
			break
		case DependType:
			key = fmt.Sprintf("%s-%s", key, "depend")
			break
		}
	}
	return key
}

func AddExecutionBehavior(key string, behaviors ...Type) string {
	for _, behavior := range behaviors {
		switch behavior {
		case RunnableType:
			key = fmt.Sprintf("%s-%s", key, "run")
			break
		case DeferType:
			key = fmt.Sprintf("%s-%s", key, "defer")
			break
		case TriggerType:
			key = fmt.Sprintf("%s-%s", key, "trigger")
			break
		case DependType:
			key = fmt.Sprintf("%s-%s", key, "depend")
			break
		}
	}
	return key
}

func SwapRunnableAndDefer(key string) string {
	if strings.Contains(key, "defer") {
		key = strings.ReplaceAll(key, "defer", "run")
	}
	if strings.Contains(key, "run") {
		key = strings.ReplaceAll(key, "run", "defer")
	}
	return key
}
