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
			Repositories: make([]dataModel.RepositoryModel, 0),
		}
	)
	if len(split) == 0 {
		return nearest, nil
	}
	for i := 0; i < len(split); i++ {
		id, err := strconv.ParseUint(split[i], 10, 64)
		if err != nil {
			continue
		}
		repository := dataModel.RepositoryModel{}
		repository.ID = uint(id)
		nearest.Repositories = append(nearest.Repositories, repository)
	}
	return nearest, nil
}
