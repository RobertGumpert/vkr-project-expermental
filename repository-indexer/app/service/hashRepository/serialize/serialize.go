package serialize

import (
	"errors"
	"fmt"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strconv"
	"strings"
)

func KeyWord(keyWord interface{}) (string, error) {
	var (
		convert, ok = keyWord.(string)
	)
	if !ok {
		return convert, errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	return convert, nil
}

func PositionKeyWord(position interface{}) (string, error) {
	var (
		pos, ok = position.(int64)
		convert string
	)
	if !ok {
		return convert, errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	convert = strconv.FormatInt(pos, 10)
	return convert, nil
}

func RepositoryID(id interface{}) (string, error) {
	var (
		i, ok   = id.(uint)
		convert string
	)
	if !ok {
		return convert, errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	convert = fmt.Sprint(i)
	return convert, nil
}

func NearestRepositories(repositories interface{}) (string, error) {
	var (
		repos, ok = repositories.(dataModel.NearestRepositoriesJSON)
		sl        = make([]string, 0)
		convert   string
	)
	if len(repos.Repositories) == 0 {
		return convert, nil
	}
	if !ok {
		return convert, errors.New("DOESN'T CONVERT 'STRING' TO STRING")
	}
	for id, distance := range repos.Repositories {
		var (
			distanceToString, idToString, objToString string
		)
		distanceToString = fmt.Sprintf("%f", distance)
		idToString = strconv.Itoa(int(id))
		objToString = strings.Join([]string{
			idToString,
			distanceToString,
		}, ":")
		sl = append(sl, objToString)
	}
	convert = strings.Join(sl, ",")
	return convert, nil
}
