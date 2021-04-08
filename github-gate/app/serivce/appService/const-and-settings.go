package appService

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
)

var (
	ErrorEmptyOrIncompleteJSONData = errors.New("Empty Or Incomplete JSON Data. ")
)


const (
	SingleTaskDownloadRepositoryByName    itask.Type = 10
	SingleTaskDownloadRepositoryByKeyWord itask.Type = 11
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
