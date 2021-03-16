package generateTaskKey

import (
	"fmt"
	"strings"
)

type ExecutionBehavior int

const (
	RunnableBehavior  ExecutionBehavior = 0
	DeferBehavior     ExecutionBehavior = 1
	TriggeredBehavior ExecutionBehavior = 2
	DependBehavior    ExecutionBehavior = 3
)

func GenerateUniqueKey(number int, behaviors ...ExecutionBehavior) string {
	var key = fmt.Sprintf("key[%d]", number)
	for _, behavior := range behaviors {
		switch behavior {
		case RunnableBehavior:
			key = fmt.Sprintf("%s-%s", key, "run")
			break
		case DeferBehavior:
			key = fmt.Sprintf("%s-%s", key, "defer")
			break
		case TriggeredBehavior:
			key = fmt.Sprintf("%s-%s", key, "trigger")
			break
		case DependBehavior:
			key = fmt.Sprintf("%s-%s", key, "depend")
			break
		}
	}
	return key
}

func AddExecutionBehavior(key string, behaviors ...ExecutionBehavior) string {
	for _, behavior := range behaviors {
		switch behavior {
		case RunnableBehavior:
			key = fmt.Sprintf("%s-%s", key, "run")
			break
		case DeferBehavior:
			key = fmt.Sprintf("%s-%s", key, "defer")
			break
		case TriggeredBehavior:
			key = fmt.Sprintf("%s-%s", key, "trigger")
			break
		case DependBehavior:
			key = fmt.Sprintf("%s-%s", key, "depend")
			break
		}
	}
	return key
}

func ChangeRunnableAndDefer(key string) string {
	if strings.Contains(key, "defer") {
		key = strings.ReplaceAll(key, "defer", "run")
	}
	if strings.Contains(key, "run") {
		key = strings.ReplaceAll(key, "run", "defer")
	}
	return key
}
