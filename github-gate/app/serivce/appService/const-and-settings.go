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
	TaskTypeNewRepositoryWithExistKeyword      itask.Type = 100
	TaskTypeNewRepositoryWithNewKeyword        itask.Type = 101
)

//
// JSON
//

type JsonRepository struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type JsonNewRepositoryWithExistKeyword struct {
	Repositories []JsonRepository `json:"repositories"`
}

type JsonNewRepositoryWithNewKeyword struct {
	Keyword    string         `json:"keyword"`
	Repository JsonRepository `json:"repository"`
}

//
type JsonSingleTaskDownloadRepositoriesByKeyWord struct {
	KeyWord string `json:"key_word"`
}

type JsonSingleTaskDownloadRepositoryAndRepositoriesByKeyWord struct {
	Repository JsonRepository `json:"repository"`
	KeyWord    string         `json:"key_word"`
}
