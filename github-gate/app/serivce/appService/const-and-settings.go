package appService

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
)

var (
	ErrorEmptyOrIncompleteJSONData = errors.New("Empty Or Incomplete JSON Data. ")
)

const (
	TaskTypeDownloadRepositoryByName           itask.Type = 10
	TaskTypeDownloadRepositoryByKeyWord        itask.Type = 11
	TaskTypeRepositoryAndRepositoriesByKeyWord itask.Type = 12
	CompositeTaskNewRepositoryWithExistWord           itask.Type = 100
)

//
// JSON
//

type JsonRepositoryName struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type JsonSingleTaskDownloadRepositoriesByName struct {
	Repositories []JsonRepositoryName `json:"repositories"`
}

type JsonSingleTaskDownloadRepositoriesByKeyWord struct {
	KeyWord string `json:"key_word"`
}

type JsonSingleTaskDownloadRepositoryAndRepositoriesByKeyWord struct {
	Repository JsonRepositoryName `json:"repository"`
	KeyWord    string             `json:"key_word"`
}
