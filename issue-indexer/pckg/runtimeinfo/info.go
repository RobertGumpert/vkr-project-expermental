package runtimeinfo

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
)

func Runtime(skip int) (info string) {
	var (
		function = "undefined func"
		pckg     = "undefined package"
	)
	pc, f, lineInt, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	s := strings.Split(f, "/")
	f = s[len(s)-1]
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
	return fmt.Sprintf("LINE=[%s]; FUNC=[%s]; PACKAGE=[%s]; FILE=[%s]", strconv.Itoa(lineInt), function, pckg, f)
}

func LogError(err ...interface{}) {
	log.Println(Runtime(2), ", ERROR: ", err, "; ")
}

func LogInfo(info ...interface{}) {
	log.Println(Runtime(2), ", INFO: ", info, "; ")
}

func LogFatal(err ...interface{}) {
	str := fmt.Sprintf("%s%s%s%s", Runtime(2), ", FATAL: ",err,"; ")
	log.Println(str)
	panic(str)
}
