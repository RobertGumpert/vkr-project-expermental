package hashRepository

import (
	"github.com/RobertGumpert/gosimstor"
	"repository-indexer/app/service/hashRepository/deserialize"
	"repository-indexer/app/service/hashRepository/serialize"
	"strings"
)

const (
	dictionary          string = "dictionary"
	nearestRepositories string = "nearest"
)

type LocalHashStorage struct {
	storage         *gosimstor.Storage
	pathRootProject string
}

func NewLocalHashStorage(pathRootProject string) (*LocalHashStorage, error) {
	storage := &LocalHashStorage{pathRootProject: pathRootProject}
	storage.pathRootProject = strings.Join([]string{
		pathRootProject,
		"data",
		"storage",
	}, "/")
	localStorage, err := storage.createLocalStorage()
	if err != nil {
		return nil, err
	}
	storage.storage = localStorage
	return storage, nil
}

func (storage *LocalHashStorage) createLocalStorage() (*gosimstor.Storage, error) {
	return gosimstor.NewStorage(
		gosimstor.NewFileProvider(
			dictionary,
			storage.pathRootProject,
			1,
			serialize.KeyWord,
			serialize.PositionKeyWord,
			deserialize.KeyWord,
			deserialize.PositionKeyWord,
		),
		gosimstor.NewFileProvider(
			nearestRepositories,
			storage.pathRootProject,
			3,
			serialize.RepositoryID,
			serialize.NearestRepositories,
			deserialize.RepositoryID,
			deserialize.NearestRepositories,
		),
	)
}
