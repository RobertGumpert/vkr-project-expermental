package appService

import (
	"errors"
	"strconv"
	"strings"
)

func ToStringKeyWord(keyWord interface{}) (string, error) {
	var (
		convert, ok = keyWord.(string)
	)
	if !ok {
		return convert,  errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	return convert,  nil
}

func ToStringPositionKeyWord(position interface{}) (string, error) {
	var (
		pos, ok = position.(int64)
		convert string
	)
	if !ok {
		return convert,  errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	convert = strconv.FormatInt(pos, 10)
	return convert,  nil
}

func ToStringRepositoryName(name interface{}) (string, error) {
	var (
		convert, ok = name.(string)
	)
	if !ok {
		return convert,  errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	return convert,  nil
}

func ToStringNearestRepositories(repositories interface{}) (string, error) {
	var (
		repos, ok = repositories.([]string)
		convert string
	)
	if !ok {
		return convert,  errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	convert = strings.Join(repos, ",")
	return convert,  nil
}