package deserialize

import (
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strconv"
	"strings"
)

func KeyWord(keyWord string) (interface{}, error) {
	return keyWord, nil
}

func PositionKeyWord(position string) (interface{}, error) {
	return strconv.ParseInt(position, 10, 64)
}

func RepositoryID(id string) (interface{}, error) {
	convert, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	return uint(convert), nil
}

func NearestRepositories(repositories string) (interface{}, error) {
	var (
		split   = strings.Split(repositories, ",")
		nearest = dataModel.NearestRepositoriesJSON{
			Repositories: make(map[uint]float64),
		}
	)
	if len(split) == 0 {
		return nearest, nil
	}
	for i := 0; i < len(split); i++ {
		var (
			objToString                  []string
			distanceToString, idToString string
			id                           uint
			distance                     float64
		)
		objToString = strings.Split(split[i], ":")
		if strings.TrimSpace(objToString[0]) != "" {
			idToString = objToString[0]
		} else {
			continue
		}
		if strings.TrimSpace(objToString[1]) != "" {
			distanceToString = objToString[1]
		} else {
			continue
		}
		id64, err := strconv.ParseUint(idToString, 10, 64)
		if err != nil {
			continue
		} else {
			id = uint(id64)
		}
		distance, err = strconv.ParseFloat(distanceToString, 64)
		if err != nil {
			continue
		}
		nearest.Repositories[id] = distance
	}
	return nearest, nil
}
