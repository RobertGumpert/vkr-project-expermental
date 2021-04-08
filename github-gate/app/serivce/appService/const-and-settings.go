package appService

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
)


const(
	ApiTaskDownloadRepositoryByName itask.Type = 10
)

//
//
//

var(
	ErrorEmptyOrIncompleteJSONData = errors.New("Empty Or Incomplete JSON Data. ")
)

//
// JSON
//

type JsonRepositoryName struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type ApiJsonDownloadRepositoriesByName struct {
	Repositories []JsonRepositoryName `json:"repositories"`
}
