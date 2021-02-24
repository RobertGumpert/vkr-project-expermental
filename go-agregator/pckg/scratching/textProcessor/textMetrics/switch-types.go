package textMetrics

import (
	"fmt"
	"go-agregator/pckg/runtimeinfo"
)

func addPanic() {
	panic(
		fmt.Sprintf(
			"%s, PANIC : %s",
			runtimeinfo.Runtime(2),
			"Type not number or not 64 size.",
		),
	)
}

func switchToFloat64(value interface{}) float64 {
	switch value.(type) {
	case int64:
		v := value.(int64)
		return float64(v)
	case float64:
		return value.(float64)
	default:
		addPanic()
	}
	return 0
}
