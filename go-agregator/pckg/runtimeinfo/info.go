package runtimeinfo

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

func Runtime(skip int) (info string) {
	var (
		function = "undefined func"
		pckg     = "undefined package"
	)
	pc, _, lineInt, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	function = runtime.FuncForPC(pc).Name()
	if strings.Contains(function, "/") {
		//
		split := strings.Split(function, "/")
		function = split[len(split)-1]
		//
		functionSplit := strings.Split(function, ".")
		function = functionSplit[len(functionSplit)-1]
		//
		split = split[0 : len(split)-1]
		split = append(split, functionSplit[0])
		pckg = strings.Join(split, "/")
	} else {
		if strings.Contains(function, ".") {
			functionSplit := strings.Split(function, ".")
			function = functionSplit[len(functionSplit)-1]
			pckg = functionSplit[0]
		}
	}
	return fmt.Sprintf("LINE=[%s]; FUNC=[%s]; PACKAGE=[%s]", strconv.Itoa(lineInt), function, pckg)
}
